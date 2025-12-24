package models

import (
	"github.com/94peter/vulpes/db/mgo"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func NewAggrAthleteJoinRacesDistinctByCompetitionName(athleteName string) *AggrAthleteJoinRacesDistinctByRace {
	return &AggrAthleteJoinRacesDistinctByRace{
		Index:       raceCollection,
		athleteName: athleteName,
	}
}

type AggrAthleteJoinRacesDistinctByRace struct {
	mgo.Index       `bson:"-"`
	CompetitionName string `bson:"competition_name"`

	athleteName string
}

func (a *AggrAthleteJoinRacesDistinctByRace) GetPipeline(q bson.M) mongo.Pipeline {
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
		// --- 新增：使用 $group 達到 Distinct 效果 ---
		// Stage 5: $group - 依照 competition_name 分組（去重）
		{
			{Key: "$group", Value: bson.M{
				"_id": "$competition_name",
			}},
		},

		// Stage 6: $project - 重新整理欄位名稱（選用）
		// 因為 $group 會把結果放在 _id，如果你希望欄位名還是 competition_name，可以再 project 一次
		{
			{Key: "$project", Value: bson.M{
				"_id":              0,
				"competition_name": "$_id",
			}},
		},
	}
	return pipeline
}
