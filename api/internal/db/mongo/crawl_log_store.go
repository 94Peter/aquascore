package mongo

import (
	"context"
	"errors"
	"fmt"

	"aquascore/internal/db/mongo/models"

	"github.com/94peter/vulpes/db/mgo"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Query interface {
	Query() bson.M
}

type CrawlLogStore interface {
	SaveCrawlLog(ctx context.Context, crawlLog *models.CrawlLog) error
	FindOneCrawlLog(ctx context.Context, q Query) (*models.CrawlLog, error)
}

func newCrawlLogStore() CrawlLogStore {
	return &crawlLogStore{}
}

type crawlLogStore struct{}

func (*crawlLogStore) SaveCrawlLog(ctx context.Context, crawlLog *models.CrawlLog) error {
	_, err := mgo.Save(ctx, crawlLog)
	if err != nil {
		return fmt.Errorf("save crawl log error: %w", err)
	}
	return nil
}

func (*crawlLogStore) FindOneCrawlLog(ctx context.Context, q Query) (*models.CrawlLog, error) {
	crawlLog := models.NewCrawlLog()
	err := mgo.FindOne(ctx, crawlLog, q.Query())
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, fmt.Errorf("find crawl log error: %w", err)
	}
	return crawlLog, nil
}

func NewCrawlLogQueryByUrl(url string) Query {
	return &queryCrawlLogByUrl{url: url}
}

type queryCrawlLogByUrl struct {
	url string
}

func (q *queryCrawlLogByUrl) Query() bson.M {
	return bson.M{"url": q.url}
}
