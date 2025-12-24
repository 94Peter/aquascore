package crawler

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

const targetURL = "https://ctsa.utk.com.tw/CTSA_114/public/race/game_data.aspx"

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

func TestAdd(t *testing.T) {
	mockPersisce := &mockPersistence{
		persisRace: func(race *Race) error {
			fmt.Println(race.CompetitionName, race.EventName, len(race.Results), race.AgeGroup, race.Gender, race.Results[0].Name[0])
			assert.NotNil(t, race)
			return nil
		},
		persistCrawlLog: func(url string) error {
			return nil
		},
		isCrawled: func(url string) (bool, error) {
			return false, nil
		},
	}
	crawler, err := NewCtsaCrawler(WithBaseURL(targetURL), withTest(true), WithPersistence(mockPersisce))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(crawler.Crawl())
	assert.False(t, true)
}

func Test_reg(t *testing.T) {
	// reAgeGender := regexp.MustCompile(`(((\d+\s*&\s*\d+|\d+\s*~\s*\d+|\d+及以上|\d+及以下)歲級)|([\s\p{Han}]+級))(.+?組)`)
	regStr := `(([\s\d]+[\s&~及]+[\s\d\p{Han}]+歲級)|([\s\p{Han}]+級))(.+?組)`
	reAgeGender := regexp.MustCompile(regStr)
	matches := reAgeGender.FindStringSubmatch("國小低年級女子組游泳 100公尺蝶式 計時決賽")
	fmt.Println(matches)
	matches = reAgeGender.FindStringSubmatch("13 & 14歲級女子組游泳 200公尺混合式 計時決賽")
	fmt.Println(matches[4])
	fmt.Println(matches)
	assert.False(t, true)
}

func BenchmarkCrawl(b *testing.B) {
	mockPersisce := &mockPersistence{
		persisRace: func(race *Race) error {
			fmt.Println(race.CompetitionName, race.EventName, len(race.Results), race.AgeGroup, race.Gender, race.Results[0].Name[0])
			
			return nil
		},
		persistCrawlLog: func(url string) error {
			return nil
		},
		isCrawled: func(url string) (bool, error) {
			return false, nil
		},
	}
	for i := 0; i < b.N; i++ {
		
	crawler, err := NewCtsaCrawler(WithBaseURL(targetURL), withTest(true), WithPersistence(mockPersisce))
	if err != nil {
		b.Fatal(err)
	}
	fmt.Println(crawler.Crawl())
	}
}
