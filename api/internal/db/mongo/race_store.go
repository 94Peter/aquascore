package mongo

import (
	"context"
	"fmt"

	"aquascore/api/internal/db/mongo/models"

	"github.com/94peter/vulpes/db/mgo"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

type RaceStore interface {
	SaveRace(ctx context.Context, race *models.Race) (bson.ObjectID, error)
	SaveRaceResults(ctx context.Context, results []*models.RaceResult) error
	GetAthleteNames(ctx context.Context) ([]string, error)
	GetYears(ctx context.Context) ([]string, error)
	GetCompetitions(ctx context.Context, year string, athlete string) ([]string, error)
	GetAthleteRaces(
		ctx context.Context, athleteName, competitionName, year string) ([]*models.AggrAthleteJoinRacesFilterByRace, error)
	GetAllAthleteRaces(ctx context.Context, athleteName string) ([]*models.AggrAthleteJoinRacesFilterByAthlete, error)
	GetRaceWithResultsByID(ctx context.Context, raceID string) (*models.AggrRaceWithResult, error)
}

func newRaceStore(tracer trace.Tracer) RaceStore {
	if tracer == nil {
		tracer = noop.NewTracerProvider().Tracer("noop")
	}
	return &raceStore{
		tracer: tracer,
	}
}

type raceStore struct {
	tracer trace.Tracer
}

func spanErrorHandler(err error, span trace.Span) error {
	if err != nil {
		span.RecordError(err)
		return err
	}
	span.SetStatus(codes.Ok, "ok")
	return nil
}

func (rs *raceStore) startTracer(ctx context.Context, name string) (context.Context, trace.Span) {
	ctx, span := rs.tracer.Start(ctx, name, trace.WithSpanKind(trace.SpanKindInternal))
	return ctx, span
}

func (rs *raceStore) GetYears(ctx context.Context) ([]string, error) {
	race := models.NewRace()
	ctx, span := rs.startTracer(ctx, "RaceStore.GetYears")
	defer span.End()
	result, err := mgo.Distinct[string](ctx, race.C(), "year", bson.M{})
	if err := spanErrorHandler(err, span); err != nil {
		return nil, err
	}
	return result, spanErrorHandler(nil, span)
}

func (rs *raceStore) GetAthleteNames(ctx context.Context) ([]string, error) {
	ctx, span := rs.startTracer(ctx, "RaceStore.GetAthleteNames")
	defer span.End()
	raceResult := models.NewRaceResult()
	result, err := mgo.Distinct[string](ctx, raceResult.C(), "name", bson.M{})
	if err := spanErrorHandler(err, span); err != nil {
		return nil, err
	}
	return result, spanErrorHandler(nil, span)
}

func (rs *raceStore) GetCompetitions(ctx context.Context, year string, athlete string) ([]string, error) {
	ctx, span := rs.startTracer(ctx, "RaceStore.GetCompetitions")
	defer span.End()
	var result []string
	// return all race in year
	if athlete != "" {
		aggr := models.NewAggrAthleteJoinRacesDistinctByCompetitionName(athlete)
		docs, err := mgo.PipeFind(ctx, aggr, bson.M{"year": year})
		if err := spanErrorHandler(err, span); err != nil {
			return nil, err
		}
		result = make([]string, 0, len(docs))
		for _, doc := range docs {
			result = append(result, doc.CompetitionName)
		}
		return result, spanErrorHandler(nil, span)
	}

	race := models.NewRace()
	result, err := mgo.Distinct[string](ctx, race.C(), "competition_name", bson.M{"year": year})
	if err := spanErrorHandler(err, span); err != nil {
		return nil, err
	}
	return result, spanErrorHandler(nil, span)
}

func (rs *raceStore) GetAthleteRaces(
	ctx context.Context, athleteName, competitionName, year string,
) ([]*models.AggrAthleteJoinRacesFilterByRace, error) {
	ctx, span := rs.startTracer(ctx, "RaceStore.GetAthleteRaces")
	defer span.End()
	aggr := models.NewAggrAthleteJoinRacesFilterByRace(athleteName)
	result, err := mgo.PipeFind(ctx, aggr, bson.M{"competition_name": competitionName, "year": year})
	if err := spanErrorHandler(err, span); err != nil {
		return nil, err
	}
	return result, spanErrorHandler(nil, span)
}

func (rs *raceStore) GetAllAthleteRaces(
	ctx context.Context, athleteName string,
) ([]*models.AggrAthleteJoinRacesFilterByAthlete, error) {
	ctx, span := rs.startTracer(ctx, "RaceStore.GetAllAthleteRaces")
	defer span.End()
	query := bson.M{
		"name": bson.M{"$in": []string{athleteName}},
	}
	aggr := models.NewAggrAthleteJoinRacesFilterByAthlete()
	result, err := mgo.PipeFind(ctx, aggr, query)
	if err := spanErrorHandler(err, span); err != nil {
		return nil, err
	}
	return result, spanErrorHandler(nil, span)
}

func (rs *raceStore) GetRaceWithResultsByID(ctx context.Context, raceID string) (*models.AggrRaceWithResult, error) {
	ctx, span := rs.startTracer(ctx, "RaceStore.GetRaceWithResultsByID")
	defer span.End()
	oid, err := bson.ObjectIDFromHex(raceID)
	if err != nil {
		return nil, spanErrorHandler(fmt.Errorf("invalid raceID: %w", err), span)
	}
	raceResult := models.NewAggrRaceWithResult()
	err = mgo.PipeFindOne(ctx, raceResult, bson.M{"_id": oid})
	if err != nil {
		return nil, spanErrorHandler(fmt.Errorf("failed to find race: %w", err), span)
	}
	return raceResult, spanErrorHandler(nil, span)
}

func (rs *raceStore) SaveRace(ctx context.Context, race *models.Race) (bson.ObjectID, error) {
	ctx, span := rs.startTracer(ctx, "SaveRace to mongo")
	defer span.End()
	r, err := mgo.Save(ctx, race)
	if err != nil {
		return bson.NilObjectID, spanErrorHandler(fmt.Errorf("failed to save race: %w", err), span)
	}
	return r.ID, spanErrorHandler(nil, span)
}

func (rs *raceStore) SaveRaceResults(ctx context.Context, results []*models.RaceResult) error {
	ctx, span := rs.startTracer(ctx, "SaveRaceResults to mongo")
	defer span.End()
	raceResult := models.NewRaceResult()
	bulk, err := mgo.NewBulkOperation(raceResult.C())
	if err != nil {
		return spanErrorHandler(fmt.Errorf("failed to create bulk operation: %w", err), span)
	}
	for _, r := range results {
		if r == nil {
			continue
		}
		bulk = bulk.InsertOne(r)
	}
	_, err = bulk.Execute(ctx)
	if err != nil {
		return spanErrorHandler(fmt.Errorf("failed to execute bulk operation: %w", err), span)
	}
	return spanErrorHandler(nil, span)
}
