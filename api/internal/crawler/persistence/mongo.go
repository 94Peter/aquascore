package persistence

import (
	"context"
	"fmt"
	"time"

	"aquascore/api/internal/crawler"
	"aquascore/api/internal/db/mongo"
	"aquascore/api/internal/db/mongo/models"

	"go.mongodb.org/mongo-driver/v2/bson"
)

const defaultTimeout = time.Second * 5

func NewMongoPersistence(raceStore mongo.RaceStore, crawlLogStore mongo.CrawlLogStore) crawler.Persistence {
	return &mongoPersistence{raceStore, crawlLogStore}
}

type mongoPersistence struct {
	raceStore     mongo.RaceStore
	crawlLogStore mongo.CrawlLogStore
}

func (m *mongoPersistence) PersistRace(race *crawler.Race) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	raceId, err := m.raceStore.SaveRace(ctx, raceToModelRace(race))
	if err != nil {
		return fmt.Errorf("save race fail: %w", err)
	}
	raceResults := make([]*models.RaceResult, len(race.Results))
	for i, raceResult := range race.Results {
		raceResults[i] = raceResultToModelRaceResult(raceId, raceResult)
	}
	err = m.raceStore.SaveRaceResults(ctx, raceResults)
	if err != nil {
		return fmt.Errorf("save race results fail: %w", err)
	}
	return nil
}

func (m *mongoPersistence) CrawlLog(url string) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	crawlLog := models.NewCrawlLog()
	crawlLog.URL = url
	crawlLog.CreatedAt = time.Now()
	err := m.crawlLogStore.SaveCrawlLog(ctx, crawlLog)
	if err != nil {
		return fmt.Errorf("save crawl log fail: %w", err)
	}
	return nil
}

func (m *mongoPersistence) IsCrawled(url string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	crawlLog, err := m.crawlLogStore.FindOneCrawlLog(ctx, mongo.NewCrawlLogQueryByUrl(url))
	if err != nil {
		return false, fmt.Errorf("get crawl log fail: %w", err)
	}
	if crawlLog == nil {
		return false, nil
	}
	return true, nil
}

func raceToModelRace(race *crawler.Race) *models.Race {
	modelRace := models.NewRace()
	modelRace.Organizer = race.Organizer
	modelRace.Type = race.Type
	modelRace.Year = race.Year
	modelRace.CompetitionName = race.CompetitionName
	modelRace.Gender = race.Gender
	modelRace.AgeGroup = race.AgeGroup
	modelRace.EventType = race.EventType
	modelRace.EventName = race.EventName
	modelRace.GamesRecord = race.GamesRecord
	modelRace.NationalRecord = race.NationalRecord
	modelRace.Time = race.Time
	modelRace.CreatedAt = time.Now()
	return modelRace
}

func raceResultToModelRaceResult(raceId bson.ObjectID, raceResult *crawler.RaceResult) *models.RaceResult {
	if len(raceResult.Name) == 0 {
		return nil
	}
	modelRaceResult := models.NewRaceResult()
	modelRaceResult.Name = raceResult.Name
	modelRaceResult.Unit = raceResult.Unit
	modelRaceResult.RaceId = raceId
	modelRaceResult.Note = raceResult.Note
	modelRaceResult.Rank = raceResult.Rank
	modelRaceResult.Record = raceResult.Record
	modelRaceResult.Score = raceResult.Score
	return modelRaceResult
}
