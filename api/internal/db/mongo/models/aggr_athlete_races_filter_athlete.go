package models

import (
	"time"

	"github.com/94peter/vulpes/db/mgo"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func NewAggrAthleteJoinRacesFilterByAthlete() *AggrAthleteJoinRacesFilterByAthlete {
	return &AggrAthleteJoinRacesFilterByAthlete{
		Index: raceResultCollection,
	}
}

type AggrAthleteJoinRacesFilterByAthlete struct {
	mgo.Index       `bson:"-"`
	RaceID          string    `bson:"race_id"`
	CompetitionName string    `bson:"competition_name"`
	PoolType        string    `bson:"pool_type"`
	EventName       string    `bson:"event_name"`
	EventType       string    `bson:"event_type"`
	EventDate       time.Time `bson:"event_date"`
	Record          float64   `bson:"record"`
	Rank            int       `bson:"rank"`
	Score           int       `bson:"score"`
	Note            string    `bson:"note"`
}

func (a *AggrAthleteJoinRacesFilterByAthlete) GetPipeline(q bson.M) mongo.Pipeline {
	pipeline := mongo.Pipeline{
		// Stage 1: $match - Filter the Competition Events
		{
			{Key: "$match", Value: q},
		},
		// Stage 2: $lookup - Join with Race Results
		{
			{Key: "$lookup", Value: bson.M{
				"from":         raceCollectionName, // Assuming raceResult.C().Name() returns the collection name
				"localField":   "race_id",
				"foreignField": "_id",
				"as":           "results",
			}},
		},
		{
			{Key: "$unwind", Value: "$results"},
		},
		// Stage 4: $project - Reshape the output document
		{
			{Key: "$project", Value: bson.M{
				"race_id":          bson.M{"$toString": "$results._id"},
				"competition_name": "$results.competition_name",
				"event_name":       "$results.event_name",
				"event_type":       "$results.event_type",
				"event_date":       "$results.time",
				"pool_type":        "$results.pool_type",
				"record":           "$record",
				"rank":             "$rank",
				"score":            "$score",
				"note":             "$note",
			}},
		},
	}
	return pipeline
}
