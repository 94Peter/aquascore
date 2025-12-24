package models

import (
	"time"

	"github.com/94peter/vulpes/db/mgo"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

const raceCollectionName = "race"

var raceCollection = mgo.NewCollectDef(raceCollectionName, func() []mongo.IndexModel {
	return []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "competition_name", Value: 1}, {Key: "year", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "year", Value: 1}},
		},
	}
})

func init() {
	mgo.RegisterIndex(raceCollection)
}

func NewRace() *Race {
	return &Race{
		Index: raceCollection,
		ID:    bson.NewObjectID(),
	}
}

type Race struct {
	mgo.Index `bson:"-"`
	ID        bson.ObjectID `bson:"_id,omitempty"`
	// 預賽 / 決賽
	Type            string        // 賽事類型 (預賽/決賽)
	Organizer       string        // 主辦單位
	Year            string        // 年份
	CompetitionName string        `bson:"competition_name"` // 競賽名稱
	Gender          string        // 性別組別
	PoolType        string        `bson:"pool_type"`
	AgeGroup        string        `bson:"age_group"`       // 年齡組別
	EventType       string        `bson:"event_type"`      // 項目類型
	EventName       string        `bson:"event_name"`      // 項目名稱
	GamesRecord     time.Duration `bson:"games_record"`    // 大會紀錄
	NationalRecord  time.Duration `bson:"national_record"` // 全國紀錄
	Time            time.Time     // 賽事時間
	CreatedAt       time.Time     `bson:"created_at"` // 創建時間
}

func (s *Race) GetId() any {
	if s.ID.IsZero() {
		return nil
	}
	return s.ID
}

func (s *Race) SetId(id any) {
	oid, ok := id.(bson.ObjectID)
	if !ok {
		return
	}
	s.ID = oid
}

func (p *Race) Validate() error {
	return nil
}
