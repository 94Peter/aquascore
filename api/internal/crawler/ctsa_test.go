package crawler

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"testing"
	"time"

	"github.com/antchfx/htmlquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_getRaceIDs(t *testing.T) {
	mockP := &mockPersistence{
		persisRace:      func(*Race) error { return nil },
		persistCrawlLog: func(string) error { return nil },
		isCrawled:       func(string) (bool, error) { return false, nil },
	}
	crawler, err := NewCtsaCrawler(
		withGetResponse(func(string) (io.Reader, error) {
			file, err := os.Open("test_file/ctsa/get_race_ids.html")
			if err != nil {
				log.Fatal(err)
			}
			// Defer the closing of the file until the main function returns
			defer file.Close()
			data, err := io.ReadAll(file)
			if err != nil {
				return nil, fmt.Errorf("failed to read all data: %w", err)
			}
			return bytes.NewReader(data), nil
		}),
		WithPersistence(mockP),
	)
	require.NoError(t, err)
	info := crawler.getRaceIDs(t.Context())
	assert.Len(t, info, 15)
	assert.Equal(t, "114年全國中區(1)游泳錦標賽", info[0].Name)
	assert.Equal(t, "151", info[0].ID)
}

func Test_createRace(t *testing.T) {
	mockP := &mockPersistence{
		persisRace:      func(*Race) error { return nil },
		persistCrawlLog: func(string) error { return nil },
		isCrawled:       func(string) (bool, error) { return false, nil },
	}
	crawler, err := NewCtsaCrawler(
		withGetResponse(func(string) (io.Reader, error) {
			file, err := os.Open("test_file/ctsa/record_1.html")
			if err != nil {
				log.Fatal(err)
			}
			// Defer the closing of the file until the main function returns
			defer file.Close()
			data, err := io.ReadAll(file)
			if err != nil {
				return nil, fmt.Errorf("failed to read all data: %w", err)
			}
			return bytes.NewReader(data), nil
		}),
		WithPersistence(mockP),
	)
	require.NoError(t, err)
	race, err := crawler.createRace(t.Context(), raceInfo{
		CompetitionName: "114年全國南區(1)游泳錦標賽",
		RaceName:        "11 & 12歲級女子組200公尺自由式 計時決賽",
	})
	require.NoError(t, err)
	assert.Equal(t, "11&12歲級", race.AgeGroup)
	expectTimeDuration, _ := parseTimeDuration("01:59.93")
	assert.Equal(t, expectTimeDuration, race.NationalRecord)
	assert.Len(t, race.Results, 36)

	crawler, err = NewCtsaCrawler(
		withGetResponse(func(string) (io.Reader, error) {
			file, err := os.Open("test_file/ctsa/record_2.html")
			if err != nil {
				log.Fatal(err)
			}
			// Defer the closing of the file until the main function returns
			defer file.Close()
			data, err := io.ReadAll(file)
			if err != nil {
				return nil, fmt.Errorf("failed to read all data: %w", err)
			}
			return bytes.NewReader(data), nil
		}),
		WithPersistence(mockP),
	)
	require.NoError(t, err)
	race, err = crawler.createRace(t.Context(), raceInfo{
		CompetitionName: "114年全國春季游泳錦標賽",
		RaceName:        "18及以上歲級男子組400公尺混合式 計時決賽",
	})
	require.NoError(t, err)
	assert.Equal(t, "18及以上歲級", race.AgeGroup)
	expectTimeDuration, _ = parseTimeDuration("04:15.86")
	assert.Equal(t, expectTimeDuration, race.NationalRecord)
	assert.Len(t, race.Results, 14)
}

func TestParseTimeDuration(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
		wantErr  bool
	}{
		{"05:34.22", 5*time.Minute + 34*time.Second + 220*time.Millisecond, false},
		{"00:55.00", 55 * time.Second, false},
		{"invalid", 0, true},
		{"44.3", 0, true}, // Missing dot part or wrong format based on Split(":") length check
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseTimeDuration(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, got)
			}
		})
	}
}

func Test_parseRaceList(t *testing.T) {
	mockP := &mockPersistence{
		persisRace:      func(*Race) error { return nil },
		persistCrawlLog: func(string) error { return nil },
		isCrawled:       func(string) (bool, error) { return false, nil },
	}
	file, err := os.Open("test_file/ctsa/get_race.html")
	require.NoError(t, err)
	defer file.Close()

	doc, err := htmlquery.Parse(file)
	require.NoError(t, err)

	crawler, err := NewCtsaCrawler(
		WithPersistence(mockP),
	)
	require.NoError(t, err)
	raceSlice := crawler.parseRaceList(doc, "114年全國春季游泳錦標賽")
	assert.Len(t, raceSlice, 176)
	assert.Equal(t, "11 & 12歲級女子組游泳 400公尺混合式 計時決賽", raceSlice[0].RaceName)
}

// mockPersistence for testing crawler functions that need Persistence interface
type mockPersistence struct {
	persisRace      func(race *Race) error
	persistCrawlLog func(url string) error
	isCrawled       func(url string) (bool, error)
}

func (m *mockPersistence) PersistRace(race *Race) error {
	return m.persisRace(race)
}

func (m *mockPersistence) CrawlLog(url string) error {
	return m.persistCrawlLog(url)
}

func (m *mockPersistence) IsCrawled(url string) (bool, error) {
	return m.isCrawled(url)
}

// TestCtsaCrawler_Crawl is a basic integration-style test for the crawler setup
// It uses mock persistence and ensures the crawler can be initialized and called.
func TestCtsaCrawler_Crawl(t *testing.T) {
	mockP := &mockPersistence{
		persisRace:      func(*Race) error { return nil },
		persistCrawlLog: func(string) error { return nil },
		isCrawled:       func(string) (bool, error) { return false, nil },
	}

	// Using a dummy URL for initialization, actual HTTP calls will be mocked in a true integration test.
	crawler, err := NewCtsaCrawler(WithBaseURL("http://dummy.url"), WithPersistence(mockP))
	require.NoError(t, err)
	assert.NotNil(t, crawler)

	// Note: Actual Crawl() method requires external network calls,
	// which are outside the scope of a strict unit test.
	// This test primarily checks the setup and ensures no immediate panics.
	// For full functionality testing, network calls would need to be mocked extensively
	// or an integration test would be required.
	// As this is a placeholder/setup check, we don't call Crawl() here.
}
