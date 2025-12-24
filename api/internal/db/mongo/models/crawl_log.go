package models

import (
	"time"

	"github.com/94peter/vulpes/db/mgo"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const crawlLogCollectionName = "crawlLog"

var crawlLogCollection = mgo.NewCollectDef(crawlLogCollectionName, func() []mongo.IndexModel {
	return []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "url", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	}
})

func init() {
	mgo.RegisterIndex(crawlLogCollection)
}

func NewCrawlLog() *CrawlLog {
	return &CrawlLog{
		Index: crawlLogCollection,
		ID:    bson.NewObjectID(),
	}
}

type CrawlLog struct {
	mgo.Index `bson:"-"`
	ID        bson.ObjectID `bson:"_id"`
	Url       string
	CreatedAt time.Time `bson:"createdAt"`
}

func (s *CrawlLog) GetId() any {
	if s.ID.IsZero() {
		return nil
	}
	return s.ID
}

func (s *CrawlLog) SetId(id any) {
	oid, ok := id.(bson.ObjectID)
	if !ok {
		return
	}
	s.ID = oid
}

func (p *CrawlLog) Validate() error {
	return nil
}
