package models

import (
	"time"

	"github.com/94peter/vulpes/db/mgo"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func NewAggrAthleteJoinRacesFilterByRace(athleteName string) *AggrAthleteJoinRacesFilterByRace {
	return &AggrAthleteJoinRacesFilterByRace{
		Index:       raceCollection,
		athleteName: athleteName,
	}
}

type AggrAthleteJoinRacesFilterByRace struct {
	mgo.Index       `bson:"-"`
	RaceID          string    `bson:"race_id"`
	CompetitionName string    `bson:"competition_name"`
	EventName       string    `bson:"event_name"`
	EventType       string    `bson:"event_type"`
	EventDate       time.Time `bson:"event_date"`
	Record          float64   `bson:"record"`
	Rank            int       `bson:"rank"`
	Score           int       `bson:"score"`
	Note            string    `bson:"note"`
	athleteName     string
}

func (a *AggrAthleteJoinRacesFilterByRace) GetPipeline(q bson.M) mongo.Pipeline {
	pipeline := mongo.Pipeline{
		// Stage 1: $match - Filter the Competition Events
		{
			{Key: "$match", Value: q},
		},
		// Stage 2: $lookup - Join with Race Results
		{
			{Key: "$lookup", Value: bson.M{
				"from":         raceResultCollectionName, // Assuming raceResult.C().Name() returns the collection name
				"localField":   "_id",
				"foreignField": "race_id",
				"as":           "results",
			}},
		},
		// Stage 3: $unwind - Deconstruct the 'results' array
		{
			{Key: "$unwind", Value: "$results"},
		},
		// Stage 4: $match - Filter the unwound results by athlete name
		{
			{Key: "$match", Value: bson.M{
				"results.name": bson.M{"$in": []string{a.athleteName}},
			}},
		},
		// Stage 5: $project - Reshape the output document
		{
			{Key: "$project", Value: bson.M{
				"race_id":          bson.M{"$toString": "$_id"},
				"competition_name": "$competition_name",
				"event_name":       "$event_name",
				"event_type":       "$event_type",
				"event_date":       "$time",
				"record":           "$results.record",
				"rank":             "$results.rank",
				"score":            "$results.score",
				"note":             "$results.note",
			}},
		},
	}
	return pipeline
}
