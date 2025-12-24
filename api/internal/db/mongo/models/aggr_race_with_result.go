package models

import (
	"time"

	"github.com/94peter/vulpes/db/mgo"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func NewAggrRaceWithResult() *AggrRaceWithResult {
	return &AggrRaceWithResult{
		Index: raceCollection,
	}
}

type AggrRaceWithResult struct {
	mgo.Index `bson:"-"`
	ID        bson.ObjectID `bson:"_id,omitempty"`
	// 預賽 / 決賽
	Type            string        // 賽事類型 (預賽/決賽)
	Organizer       string        // 主辦單位
	Year            string        // 年份
	CompetitionName string        `bson:"competition_name"` // 競賽名稱
	Gender          string        // 性別組別
	AgeGroup        string        `bson:"age_group"`       // 年齡組別
	EventType       string        `bson:"event_type"`      // 項目類型
	EventName       string        `bson:"event_name"`      // 項目名稱
	GamesRecord     time.Duration `bson:"games_record"`    // 大會紀錄
	NationalRecord  time.Duration `bson:"national_record"` // 全國紀錄
	Time            time.Time     // 賽事時間
	CreatedAt       time.Time     `bson:"created_at"` // 創建時間
	Results         []*struct {
		Unit   string        `bson:"unit"`   // 單位
		Name   []string      `bson:"name"`   // 選手姓名
		Record time.Duration `bson:"record"` // 成績
		Rank   int           `bson:"rank"`   // 名次
		Score  int           `bson:"score"`  // 分數
		Note   string        `bson:"note"`   // 備註
	} `bson:"results"` // 結果
}

func (a *AggrRaceWithResult) GetPipeline(q bson.M) mongo.Pipeline {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: q}},
		{
			{Key: "$lookup", Value: bson.M{
				"from":         raceResultCollectionName, // Assuming raceResult.C().Name() returns the collection name
				"localField":   "_id",
				"foreignField": "race_id",
				"as":           "results",
			}},
		},
	}
	return pipeline
}
