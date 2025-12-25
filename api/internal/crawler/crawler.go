package crawler

import "time"

type Race struct {
	Organizer       string
	Year            string
	Type            string
	CompetitionName string
	Gender          string
	AgeGroup        string
	EventType       string
	EventName       string
	GamesRecord     time.Duration
	NationalRecord  time.Duration
	Time            time.Time
	Results         []*RaceResult
}

type RaceResult struct {
	Unit   string
	Name   []string
	Record time.Duration
	Rank   int32
	Score  int32
	Note   string
}

type Persistence interface {
	PersistRace(race *Race) error
	CrawlLog(url string) error
	IsCrawled(url string) (bool, error)
}

func sendNonBlockingError(err error, errChan chan error) {
	select {
	case errChan <- err:
	default:
	}
}
