package mongo

import (
	"context"

	"github.com/94peter/vulpes/db/mgo"
	"go.opentelemetry.io/otel"

	"aquascore/internal/db"
)

type Stores struct {
	CrawlLogStore CrawlLogStore
	RaceStore     RaceStore
}

var store *Stores

func IniMongodb(ctx context.Context, uri string, dbName string) (db.CloseDbFunc, error) {
	tracer := otel.Tracer("Mongodb")
	err := mgo.InitConnection(ctx, dbName, tracer, mgo.WithURI(uri), mgo.WithMinPoolSize(50), mgo.WithMaxPoolSize(100))
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

	return func(ctx context.Context) error {
		return mgo.Close(ctx)
	}, nil
}

func InjectStore(f func(*Stores)) {
	f(store)
}
