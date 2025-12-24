package models

import (
	"time"

	"github.com/94peter/vulpes/db/mgo"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

const raceResultCollectionName = "raceResult"

var raceResultCollection = mgo.NewCollectDef(raceResultCollectionName, func() []mongo.IndexModel {
	return []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "name", Value: 1}},
		},
	}
})

func init() {
	mgo.RegisterIndex(raceResultCollection)
}

func NewRaceResult() *RaceResult {
	return &RaceResult{
		Index: raceResultCollection,
		ID:    bson.NewObjectID(),
	}
}

type RaceResult struct {
	mgo.Index `bson:"-"`
	ID        bson.ObjectID `bson:"_id,omitempty"`
	Unit      string        // 單位
	Name      []string      // 選手姓名
	Record    time.Duration // 成績
	Rank      int           // 名次
	Score     int           // 分數
	Note      string        // 備註
	RaceId    bson.ObjectID `bson:"race_id,omitempty"` // 賽事ID
}

func (s *RaceResult) GetId() any {
	if s.ID.IsZero() {
		return nil
	}
	return s.ID
}

func (s *RaceResult) SetId(id any) {
	oid, ok := id.(bson.ObjectID)
	if !ok {
		return
	}
	s.ID = oid
}

func (p *RaceResult) Validate() error {
	return nil
}
