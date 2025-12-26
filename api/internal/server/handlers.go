package server

import (
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"aquascore/api/internal/db/mongo"
	"aquascore/api/internal/db/mongo/models"

	analysisv1 "buf.build/gen/go/aqua/analysis/protocolbuffers/go/analysis/v1"
	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// APIHandler holds the dependencies for API handlers.
type apiHandler struct {
	raceStore  mongo.RaceStore
	grpcClient GrpcClient
}

type AthleteRaceResult struct {
	RaceID    string  `json:"race_id"`
	EventName string  `json:"event_name"`
	Record    float64 `json:"record"`
	Rank      int     `json:"rank"`
	Score     int     `json:"score"`
	Note      string  `json:"note"`
}

// NewAPIHandler creates a new APIHandler.
func initAPIHandler(router gin.IRoutes, db *mongo.Stores, grpcClient GrpcClient) {
	handler := &apiHandler{
		raceStore:  db.RaceStore,
		grpcClient: grpcClient,
	}
	router.GET("/athletes", handler.GetAthletes)
	router.GET("/years", handler.GetYears)
	router.GET("/competitions", handler.GetCompetitions)
	router.GET("/athletes/:athlete_name/races", handler.GetAthleteRaces)
	router.GET("/athletes/:athlete_name/performance-overview", handler.GetAthletePerformanceOverview)
	router.GET("/race/:race_id/comparison", handler.GetRaceComparison)
}

// GetAthletes handles the GET /athletes endpoint.
func (h *apiHandler) GetAthletes(c *gin.Context) {
	// Get all unique athlete names
	names, err := h.raceStore.GetAthleteNames(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve athletes"})
		return
	}

	// Return an empty array instead of null if no athletes are found
	if names == nil {
		names = []string{}
	}

	c.JSON(http.StatusOK, names)
}

// GetYears handles the GET /years endpoint.
func (h *apiHandler) GetYears(c *gin.Context) {
	years, err := h.raceStore.GetYears(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve years"})
		return
	}

	// Return an empty array instead of null if no years are found
	if years == nil {
		years = []string{}
	}

	c.JSON(http.StatusOK, years)
}

// GetCompetitions handles the GET /competitions endpoint.
func (h *apiHandler) GetCompetitions(c *gin.Context) {
	year := c.Query("year")
	if year == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "year query parameter is required"})
		return
	}

	athlete := c.Query("athlete") // Read the singular athlete parameter

	competitionNames, err := h.raceStore.GetCompetitions(c.Request.Context(), year, athlete)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve competitions"})
		return
	}

	// Transform the string slice into a slice of objects to match frontend expectations
	competitions := make([]gin.H, 0, len(competitionNames))

	for _, name := range competitionNames {
		competitions = append(competitions, gin.H{"name": name})
	}

	c.JSON(http.StatusOK, competitions)
}

// GetAthleteRaces handles the GET /athletes/:athlete_name/races endpoint.
func (h *apiHandler) GetAthleteRaces(c *gin.Context) {
	athleteName := c.Param("athlete_name")
	competitionName := c.Query("competition_name")
	year := c.Query("year")

	if competitionName == "" || year == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "competition_name and year query parameters are required"})
		return
	}

	races, err := h.raceStore.GetAthleteRaces(c.Request.Context(), athleteName, competitionName, year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve athlete races"})
		return
	}

	// Return an empty array instead of null if no races are found
	if races == nil {
		c.JSON(http.StatusOK, []string{})
		return
	}

	results := make([]AthleteRaceResult, len(races))
	for i, race := range races {
		results[i] = AthleteRaceResult{
			RaceID:    race.RaceID,
			EventName: race.EventName,
			Record:    race.Record / float64(time.Second),
			Rank:      race.Rank,
			Score:     race.Score,
			Note:      race.Note,
		}
	}

	c.JSON(http.StatusOK, results)
}

// GetAthletePerformanceOverview handles the GET /athletes/:athlete_name/performance-overview endpoint.
func (h *apiHandler) GetAthletePerformanceOverview(c *gin.Context) {
	athleteName := c.Param("athlete_name")

	races, err := h.raceStore.GetAllAthleteRaces(c.Request.Context(), athleteName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve athlete races"})
		return
	}
	req := mapRacesToAnalyzePerformanceOverviewRequest(athleteName, races)
	res, err := h.grpcClient.AnalyzePerformanceOverview(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to analyze performance"})
		return
	}
	c.JSON(http.StatusOK, mapAnalysisToResponse(res.EventAnalyses))
}

func mapRacesToAnalyzePerformanceOverviewRequest(
	athleteName string,
	races []*models.AggrAthleteJoinRacesFilterByAthlete,
) *analysisv1.AnalyzePerformanceOverviewRequest {
	performanceResults := make([]*analysisv1.PerformanceResult, 0, len(races))
	appendCount := 0
	for _, race := range races {
		if race.Record == 0 {
			continue
		}
		performanceResults = append(performanceResults, &analysisv1.PerformanceResult{
			EventDate:       timestamppb.New(race.EventDate),
			ResultTime:      race.Record / float64(time.Second),
			EventType:       fmt.Sprintf("%s(%s)", race.EventType, race.PoolType),
			CompetitionName: fmt.Sprintf("%s %s", race.CompetitionName, race.EventName),
		})
		appendCount++
	}
	performanceResults = performanceResults[:appendCount]

	return &analysisv1.AnalyzePerformanceOverviewRequest{
		AthleteName: athleteName,
		Results:     performanceResults,
	}
}

func mapAnalysisToResponse(analyses []*analysisv1.EventPerformanceAnalysis) []map[string]any {
	output := make([]map[string]any, 0, len(analyses))
	for _, analysis := range analyses {
		// mins := time.Duration(analysis.PersonalBest) / time.Minute
		// secs := time.Duration(analysis.PersonalBest) % time.Minute / time.Second
		recent_races := make([]map[string]any, 0, len(analysis.RecentRaces))
		for _, race := range analysis.RecentRaces {
			recent_races = append(recent_races, map[string]any{
				"time":             race.Time,
				"date":             race.Date,
				"competition_name": race.CompetitionName,
			})
		}
		output = append(output, map[string]any{
			"event_name": analysis.EventName,
			"personal_best": map[string]any{
				"time": analysis.PersonalBest.Time,
				"unit": "s",
				"date": analysis.PersonalBest.Date,
			},
			"analysis": map[string]any{
				"stability": map[string]any{
					"value": analysis.Analysis.Stability.Value,
					"unit":  analysis.Analysis.Stability.Unit,
					"label": analysis.Analysis.Stability.Label,
				},
				"pb_freshness": map[string]any{
					"days_since_pb": analysis.Analysis.PbFreshness.DaysSincePb,
					"label":         analysis.Analysis.PbFreshness.Label,
				},
				"trend": map[string]any{
					"value": analysis.Analysis.Trend.Value,
					"unit":  analysis.Analysis.Trend.Unit,
					"label": analysis.Analysis.Trend.Label,
				},
			},
			"recent_races": recent_races,
			"charts": map[string]any{
				"sparkline": analysis.Charts.Sparkline,
				"trend_chart": map[string]any{
					"dates":   analysis.Charts.TrendChart.Dates,
					"times":   analysis.Charts.TrendChart.Times,
					"pb_line": analysis.Charts.TrendChart.PbLine,
				},
			},
		})
	}
	return output
}

// GetRaceComparison handles the GET /race/:race_id/comparison endpoint.
func (h *apiHandler) GetRaceComparison(c *gin.Context) {
	raceIDHex := c.Param("race_id")

	athleteName := c.Query("athlete_name")
	if athleteName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "athlete_name query parameter is required"})
		return
	}
	raceWithResult, err := h.raceStore.GetRaceWithResultsByID(c.Request.Context(), raceIDHex)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve race"})
		return
	}

	req, err := mapRaceWithResultToAnalyzeResultComparisonRequest(athleteName, raceWithResult)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	res, err := h.grpcClient.AnalyzeResultComparison(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to analyze result comparison"})
		return
	}
	c.JSON(http.StatusOK,
		mapAnalyzeResultComparisonResponseToResponse(raceWithResult, req, res))
}

func mapRaceWithResultToAnalyzeResultComparisonRequest(
	athleteName string,
	race *models.AggrRaceWithResult,
) (*analysisv1.AnalyzeResultComparisonRequest, error) {
	competitionResults := make([]*analysisv1.RaceResult, len(race.Results))
	var targetResult *analysisv1.RaceResult
	var resultSize int
	for _, result := range race.Results {
		if result.Record == 0 {
			continue
		}
		var resultAthleteName string
		if len(result.Name) == 1 {
			resultAthleteName = result.Name[0]
		} else {
			resultAthleteName = strings.Join(result.Name, ",")
		}
		res := &analysisv1.RaceResult{
			AthleteName: resultAthleteName,
			RecordTime:  result.Record.Seconds(),
			Rank:        result.Rank,
		}
		var hasAdded bool
		if result.Rank == 1 || result.Rank == 2 ||
			result.Rank == 3 || result.Rank == 6 ||
			result.Rank == 7 || result.Rank == 8 {
			competitionResults[resultSize] = res
			hasAdded = true
			resultSize++
		}

		if slices.Contains(result.Name, athleteName) {
			targetResult = res
			if !hasAdded {
				competitionResults[resultSize] = res
				resultSize++
			}
		}
	}
	if targetResult == nil {
		return nil, errors.New("target athlete not found in this race")
	}
	nationalRecord := race.NationalRecord.Seconds()
	gamesRecord := race.GamesRecord.Seconds()
	return &analysisv1.AnalyzeResultComparisonRequest{
		TargetResult:       targetResult,
		CompetitionResults: competitionResults[:resultSize],
		Records: &analysisv1.RecordMarks{
			NationalRecord: &nationalRecord,
			GamesRecord:    &gamesRecord,
		},
	}, nil
}

func mapAnalyzeResultComparisonResponseToResponse(
	race *models.AggrRaceWithResult,
	req *analysisv1.AnalyzeResultComparisonRequest,
	res *analysisv1.AnalyzeResultComparisonResponse,
) map[string]any {
	competitor_comparison := make([]map[string]any, 0, len(res.ResultsComparison)-1)
	var nationalRecordDiff, gamesRecordDiff *float64
	for _, comp := range res.ResultsComparison {
		if comp.AthleteName == req.TargetResult.AthleteName {
			nationalRecordDiff = comp.DiffFromNationalRecord
			gamesRecordDiff = comp.DiffFromGamesRecord
			continue
		}

		competitor_comparison = append(competitor_comparison, map[string]any{
			"rank":             comp.Rank,
			"athlete_name":     comp.AthleteName,
			"record_time":      comp.RecordTime,
			"diff_from_target": comp.DiffFromTarget,
			"diff_label":       getDiffLabel(comp.DiffFromTarget),
		})
	}
	return map[string]any{
		"target_result": map[string]any{
			"athlete_name":     req.TargetResult.AthleteName,
			"record_time":      req.TargetResult.RecordTime,
			"rank":             req.TargetResult.Rank,
			"competition_name": race.CompetitionName,
			"event_name":       race.EventName,
			"date":             race.Time.Format("2006-01-02"),
		},
		"records": map[string]any{
			"national_record": map[string]any{
				"time": req.Records.NationalRecord,
				"diff": nationalRecordDiff,
			},
			"games_record": map[string]any{
				"time": req.Records.NationalRecord,
				"diff": gamesRecordDiff,
			},
		},
		"competitor_comparison": competitor_comparison,
	}
}

func getDiffLabel(diff *float64) string {
	if diff == nil {
		return ""
	}
	switch {
	case *diff < -1.0:
		return "far_ahead"
	case *diff < 0.0:
		return "slightly_ahead"
	case *diff == 0.0:
		return "your_result"
	case *diff < 1.0:
		return "slightly_behind"
	case *diff >= 1.0:
		return "far_behind"
	}
	return ""
}
