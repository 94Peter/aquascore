package mongo

import (
	"context"

	"aquascore/api/internal/db"

	"github.com/94peter/vulpes/db/mgo"
	"go.opentelemetry.io/otel"
)

type Stores struct {
	CrawlLogStore CrawlLogStore
	RaceStore     RaceStore
}

var store *Stores

const (
	minDBPoolSize = 50
	maxDBPoolSize = 100
)

func IniMongodb(ctx context.Context, uri string, dbName string) (db.CloseDbFunc, error) {
	tracer := otel.Tracer("Mongodb")
	err := mgo.InitConnection(ctx, dbName, tracer,
		mgo.WithURI(uri), mgo.WithMinPoolSize(minDBPoolSize), mgo.WithMaxPoolSize(maxDBPoolSize))
	if err != nil {
		return nil, err
	}
	err = mgo.SyncIndexes(ctx)
	if err != nil {
		return nil, err
	}

	raceStoreTracer := otel.Tracer("RaceStore")
	store = &Stores{
		CrawlLogStore: newCrawlLogStore(),
		RaceStore:     newRaceStore(raceStoreTracer),
	}

	return mgo.Close, nil
}

func InjectStore(f func(*Stores)) {
	f(store)
}
