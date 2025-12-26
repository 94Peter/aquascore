package crawler

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

const (
	apiSleepDuration             = 100 * time.Millisecond
	defaultDelay                 = time.Second * 5
	notApplicable                = "N/A"
	expectedTimeSplitParts       = 2
	expectedDateRegexMatchGroups = 2
	maxDateSplitLimit            = 3
	expectedDateSplitParts       = 3
	rocYearOffset                = 1911
	minRaceResultColumns         = 8
	minAgeGenderRegexMatches     = 2
)

func NewCtsaCrawler(opts ...Option) (*ctsaCrawler, error) {
	crawler := &ctsaCrawler{}
	var err error
	for _, opt := range opts {
		opt(crawler)
	}
	if crawler.persistence == nil {
		return nil, errors.New("persistence is nil")
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}

	// 2. 創建一個帶有 Jar 的 http.Client
	// 這個 Client 會自動管理 Cookie 的接收和發送
	client := &http.Client{
		Jar: jar, // 將 Jar 設置給 Client
	}
	crawler.client = client
	return crawler, nil
}

func (c *ctsaCrawler) getResponse(ctx context.Context, url string) (io.Reader, error) {
	if c.mockGetResponse != nil {
		return c.mockGetResponse(url)
	}
	myctx, cancel := context.WithTimeout(ctx, defaultDelay)
	defer cancel()
	req, err := http.NewRequestWithContext(myctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET 請求返回非預期狀態碼: %d %s", resp.StatusCode, resp.Status)
	}
	// 建立一個記憶體暫存區
	var buf bytes.Buffer

	// io.Copy 會在 context 有效期間，把資料從網路串流搬到記憶體 buf
	// 如果此時 context cancel，io.Copy 會回傳錯誤
	if _, err := io.Copy(&buf, resp.Body); err != nil {
		return nil, fmt.Errorf("暫存資料失敗 (可能超時或連線中斷): %w", err)
	}
	return bytes.NewReader(buf.Bytes()), nil
}

type ctsaCrawler struct {
	baseUrl         string
	persistence     Persistence
	client          *http.Client
	mockGetResponse func(url string) (io.Reader, error)
}

func (c *ctsaCrawler) Crawl(ctx context.Context) error {
	raceIDs := c.getRaceIDs(ctx)

	if len(raceIDs) == 0 {
		log.Fatal("❌ 未能成功獲取任何比賽 ID，程序終止。")
	}
	fmt.Printf("✅ 成功找到 %d 個比賽 ID，開始逐一 POST 請求...\n", len(raceIDs))
	fmt.Println("---------------------------------------------------------")
	for _, raceID := range raceIDs {
		err := c.postForDetails(ctx, raceID)
		if err != nil {
			log.Printf("❌ POST 請求失敗: %v", err)
		} else {
			log.Printf("✅ POST 請求成功: %s", raceID.Name)
		}
		time.Sleep(apiSleepDuration)
	}
	return nil
}

type activeInfo struct {
	ID   string
	Name string
}

type raceInfo struct {
	CompetitionName string // 例如：114年全國南區(1)游泳錦標賽
	RaceName        string // 例如：11 & 12歲級女子組游泳 200公尺自由式 計時決賽
	ScoreReportURL  string // 成績報告的絕對 URL 連結
}

func (info *raceInfo) IsQualifier() bool {
	return strings.Contains(info.RaceName, "預賽") || strings.Contains(info.RaceName, "快組計時決賽")
}

func (c *ctsaCrawler) getInitialData(ctx context.Context) (map[string]string, error) {
	body, err := c.getResponse(ctx, c.baseUrl)
	if err != nil {
		return nil, fmt.Errorf("GET 請求失敗: %w", err)
	}

	doc, err := htmlquery.Parse(body)
	if err != nil {
		return nil, fmt.Errorf("HTML 解析失敗: %w", err)
	}

	// 1. 獲取隱藏欄位 (Hidden Inputs)
	hiddenInputs := htmlquery.Find(doc, "//input[@type='hidden']")
	hiddenFields := make(map[string]string)
	for _, n := range hiddenInputs {
		name := htmlquery.SelectAttr(n, "name")
		value := htmlquery.SelectAttr(n, "value")

		if name != "" {
			hiddenFields[name] = value
		}
	}
	return hiddenFields, nil
}

func (c *ctsaCrawler) getRaceIDs(ctx context.Context) []activeInfo {
	body, err := c.getResponse(ctx, c.baseUrl)
	if err != nil {
		log.Printf("GET 請求失敗: %v", err)
		return nil
	}

	doc, err := htmlquery.Parse(body)
	if err != nil {
		log.Printf("HTML 解析失敗: %v", err)
		return nil
	}

	// 假設 <select> 的 ID 是 "ddlRace" 或其他類似名稱
	// 根據常見的 ASP.NET 網站結構，我猜測 ID 可能是 ddlRace 或其他
	// 由於沒有源碼，我們嘗試直接尋找所有 <option>
	// 實際操作中，應該找到 <select name="ddlRace"> 或 <select id="ddlRace">
	// 這裡使用更通用的 XPath: 尋找所有具有 value 屬性的 <option>
	list := htmlquery.Find(doc, "//select[@name='ctl00$ContentPlaceHolder1$DD_Activity_ID']/option")

	var actives []activeInfo
	for _, n := range list {
		// 提取 value 屬性
		id := htmlquery.SelectAttr(n, "value")
		// 忽略第一個通常是 "請選擇" 或空值的 option
		name := htmlquery.InnerText(n)
		if id != "" && id != "0" {
			actives = append(actives, activeInfo{ID: id, Name: name})
		}
	}

	return actives
}

func (c *ctsaCrawler) postForDetails(ctx context.Context, active activeInfo) error {
	races, err := c.fetchRaceList(ctx, active)
	if err != nil {
		return err
	}
	return c.processRaces(ctx, races)
}

func (c *ctsaCrawler) fetchRaceList(ctx context.Context, active activeInfo) ([]raceInfo, error) {
	hiddenFields, err := c.getInitialData(ctx)
	if err != nil {
		return nil, err
	}

	form := url.Values{}
	form.Add("ctl00$ContentPlaceHolder1$DD_Activity_ID", active.ID)
	for k, v := range hiddenFields {
		form.Add(k, v)
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultDelay)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseUrl, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("構造請求失敗: %w", err)
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", fmt.Sprintf("%d", len(form.Encode())))

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("POST 請求失敗: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("POST 請求返回非預期狀態碼: %d %s", resp.StatusCode, resp.Status)
	}

	doc, err := htmlquery.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("HTML 解析失敗: %w", err)
	}

	return c.parseRaceList(doc, active.Name), nil
}

func (c *ctsaCrawler) parseRaceList(doc *html.Node, competitionName string) []raceInfo {
	xpath := "//table[@id='ctl00_ContentPlaceHolder1_GridView1']/tbody/tr[position() > 1]"
	dataRows := htmlquery.Find(doc, xpath)
	var races []raceInfo
	base, _ := url.Parse(c.baseUrl)

	for _, trNode := range dataRows {
		raceNameNode := htmlquery.FindOne(trNode, "./td[3]//a")

		raceName := notApplicable
		if raceNameNode != nil {
			raceName = htmlquery.InnerText(raceNameNode)
		}

		linkNode := htmlquery.FindOne(trNode, ".//a[text()='成績報告']")

		absoluteURL := notApplicable

		if linkNode != nil {
			relativeURL := htmlquery.SelectAttr(linkNode, "href")
			pathUrl, _ := url.Parse(relativeURL)
			absoluteURL = base.ResolveReference(pathUrl).String()
		}

		if absoluteURL != notApplicable && strings.TrimSpace(raceName) != "" {
			races = append(races, raceInfo{
				CompetitionName: competitionName,
				RaceName:        strings.TrimSpace(raceName),
				ScoreReportURL:  absoluteURL,
			})
		}
	}
	return races
}

func (c *ctsaCrawler) processRaces(ctx context.Context, races []raceInfo) error {
	const maxConcurrency = 5
	semaphore := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup
	errChan := make(chan error, 1)

	for _, race := range races {
		select {
		case err := <-errChan:
			close(semaphore)
			wg.Wait()
			return err
		default:
		}

		wg.Add(1)
		semaphore <- struct{}{}
		go c.processSingleRace(ctx, race, &wg, semaphore, errChan)
	}
	wg.Wait()
	close(errChan)

	if err, ok := <-errChan; ok {
		return err
	}
	return nil
}

func (c *ctsaCrawler) processSingleRace(
	ctx context.Context,
	race raceInfo,
	wg *sync.WaitGroup,
	semaphore chan struct{},
	errChan chan error,
) {
	defer wg.Done()
	defer func() { <-semaphore }()

	ok, err := c.persistence.IsCrawled(race.ScoreReportURL)
	if err != nil {
		sendNonBlockingError(fmt.Errorf("check crawled fail: %w", err), errChan)
		return
	}
	if ok {
		return
	}
	dbrace, err := c.createRace(ctx, race)
	if err != nil {
		sendNonBlockingError(fmt.Errorf("generate race %s [%s] fail: %w", race.CompetitionName, race.RaceName, err), errChan)
		return
	}
	err = c.persistence.PersistRace(dbrace)
	if err != nil {
		sendNonBlockingError(fmt.Errorf("persistence race fail: %w", err), errChan)
		return
	}
	err = c.persistence.CrawlLog(race.ScoreReportURL)
	if err != nil {
		sendNonBlockingError(fmt.Errorf("persistence crawl log fail: %w", err), errChan)
		return
	}
}

func (c *ctsaCrawler) createRace(ctx context.Context, info raceInfo) (*Race, error) {
	body, err := c.getResponse(ctx, info.ScoreReportURL)
	if err != nil {
		return nil, fmt.Errorf("GET 請求失敗: %w", err)
	}
	doc, err := htmlquery.Parse(body)
	if err != nil {
		return nil, fmt.Errorf("HTML 解析失敗: %w", err)
	}
	race, err := newRaceBuilder(doc, info).CreateRace()
	if err != nil {
		return nil, err
	}

	return race, nil
}

func newRaceBuilder(doc *html.Node, info raceInfo) *raceBuilder {
	return &raceBuilder{doc: doc, info: info}
}

type raceBuilder struct {
	info raceInfo
	doc  *html.Node
}

type raceRecord struct {
	gameRecord     time.Duration
	nationalRecord time.Duration
}

func parseTimeDuration(timeStr string) (time.Duration, error) {
	// 格式範例: "05:34.22"
	parts := strings.Split(timeStr, ":")
	if len(parts) != expectedTimeSplitParts {
		return 0, fmt.Errorf("時間格式錯誤，預期為 mm:ss.SS，實際為: %s", timeStr)
	}

	minutes := parts[0]
	secondsWithMillis := parts[1]

	// 構造 Duration 字串: Go 的 time.ParseDuration 接受 "5m34.22s" 這樣的格式
	durationString := fmt.Sprintf("%sm%ss", minutes, secondsWithMillis)

	return time.ParseDuration(durationString)
}

func (b *raceBuilder) getRecord() (*raceRecord, error) {
	// 3. 提取並清洗時間字串
	// 完整的文字內容是 " 大會紀錄：05:34.22   全國紀錄：04:40.21 " (包含換行和空格)
	text, err := b.innerText(
		"/html/body/form/div[3]/span/div[1]/table/tbody/tr[2]/td[3]",
		"/html/body/form/div[1]/span/div[1]/table/tbody/tr[2]/td[3]",
	)
	if err != nil {
		return nil, err
	}

	var records raceRecord

	gameRecordReg := regexp.MustCompile(`大會紀錄：\s*(\d{2}:\d{2}\.\d{2})`)
	gameMatch := gameRecordReg.FindStringSubmatch(text)
	const expectMatchSize = 2
	if len(gameMatch) == expectMatchSize {
		records.gameRecord, err = parseTimeDuration(gameMatch[1])
		if err != nil {
			return nil, fmt.Errorf("轉換大會紀錄失敗: %w", err)
		}
	}

	nationalRecordReg := regexp.MustCompile(`全國紀錄：\s*(\d{2}:\d{2}\.\d{2})`)
	nationalMatch := nationalRecordReg.FindStringSubmatch(text)
	if len(nationalMatch) == expectMatchSize {
		records.nationalRecord, err = parseTimeDuration(nationalMatch[1])
		if err != nil {
			return nil, fmt.Errorf("轉換全國紀錄失敗: %w", err)
		}
	}
	return &records, nil
}

func (b *raceBuilder) getOrganizer() (string, error) {
	return b.innerText(
		"/html/body/form/div[3]/span/h1/text()[1]",
		"/html/body/form/div[1]/span/h1/text()[1]")
}

func (b *raceBuilder) getTime() (time.Time, error) {
	text, err := b.innerText(
		"/html/body/form/div[3]/span/div[1]/table/tbody/tr[1]/td[3]",
		"/html/body/form/div[1]/span/div[1]/table/tbody/tr[1]/td[3]",
	)
	if err != nil {
		return time.UnixMicro(0), err
	}
	re := regexp.MustCompile(`(\d{2,3}/\d{2}/\d{2})`)
	match := re.FindStringSubmatch(text)

	if len(match) < expectedDateRegexMatchGroups {
		return time.Time{}, fmt.Errorf("日期格式錯誤或找不到日期: %s", text)
	}
	datePart := match[1] // 例如: "114/01/11"

	// 2. 轉換民國紀年為西元紀年
	parts := regexp.MustCompile(`/`).Split(datePart, maxDateSplitLimit)
	if len(parts) != expectedDateSplitParts {
		return time.Time{}, fmt.Errorf("日期分割錯誤: %s", datePart)
	}

	rocYearStr := parts[0]
	monthDayPart := parts[1] + "/" + parts[2] // "01/11"

	rocYear, err := strconv.Atoi(rocYearStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("解析民國年失敗: %w", err)
	}

	// 核心轉換邏輯: 西元 = 民國 + 1911
	adYear := rocYear + rocYearOffset
	adYearStr := strconv.Itoa(adYear)

	// 3. 構造西元日期字串 (例如: "2025/01/11")
	adDateStr := adYearStr + "/" + monthDayPart

	// 4. 解析為 time.Time
	// 使用 "2006/01/02" 作為標準 Go 時間格式範例
	t, err := time.Parse("2006/01/02", adDateStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("轉換為 time.Time 失敗: %w", err)
	}

	return t, nil
}

func (b *raceBuilder) getResult() ([]*RaceResult, error) {
	list, err := b.listElement(
		"/html/body/form/div[3]/span/div[2]/table/tbody/tr[position() > 1]",
		"/html/body/form/div[1]/span/div[2]/table/tbody/tr[position() > 1]")
	if err != nil {
		return nil, err
	}
	results := make([]*RaceResult, 0, len(list))
	for _, n := range list {
		tds := htmlquery.Find(n, "/td/font") // 選擇 tr 下所有 td 內的 font 標籤

		if len(tds) < minRaceResultColumns {
			// 跳過格式不正確的行
			continue
		}
		result := RaceResult{
			Unit: strings.TrimSpace(htmlquery.InnerText(tds[2])),
			Name: strings.Split(strings.TrimSpace(htmlquery.InnerText(tds[3])), " "),
			Note: strings.TrimSpace(htmlquery.InnerText(tds[7])),
		}
		recordStr := strings.TrimSpace(htmlquery.InnerText(tds[4]))
		rankStr := strings.TrimSpace(htmlquery.InnerText(tds[5]))
		scoreStr := strings.TrimSpace(htmlquery.InnerText(tds[6]))
		if recordStr != "" {
			duration, err := parseTimeDuration(recordStr)
			if err != nil {
				return nil, err
			}
			result.Record = duration
			// 如果解析失敗 (如 "逾時" 的空字串)，Record 保持為 0 (零值)
		}
		// 處理 Rank (名次)
		if !b.info.IsQualifier() && rankStr != "" {
			rank, err := stringToInt32(rankStr)
			if err != nil {
				return nil, fmt.Errorf("convert rank to int failed: %w", err)
			}
			result.Rank = rank
		}

		// 處理 Score (積點)
		if !b.info.IsQualifier() && scoreStr != "" {
			score, err := stringToInt32(scoreStr)
			if err != nil {
				return nil, fmt.Errorf("convert score to int failed: %w", err)
			}
			result.Score = score
		}
		if len(result.Name) != 0 {
			results = append(results, &result)
		}
	}
	return results, nil
}

func stringToInt32(s string) (int32, error) {
	val, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(val), nil
}

func (b *raceBuilder) CreateRace() (*Race, error) {
	organizer, err := b.getOrganizer()
	if err != nil {
		return nil, err
	}
	records, err := b.getRecord()
	if err != nil {
		return nil, err
	}
	t, err := b.getTime()
	if err != nil {
		return nil, err
	}
	var r Race
	r.Organizer = organizer
	r.GamesRecord = records.gameRecord
	r.NationalRecord = records.nationalRecord
	r.Time = t
	reAgeGender := regexp.MustCompile(`(([\s\d]+[\s&~及]+[\s\d\p{Han}]+歲級)|([\s\p{Han}]+級)|(排名賽))(.+?組)`)
	matches := reAgeGender.FindStringSubmatch(b.info.RaceName)
	r.EventName = b.info.RaceName
	remainingStr := b.info.RaceName
	if len(matches) > minAgeGenderRegexMatches {
		r.AgeGroup = strings.ReplaceAll(matches[1], " ", "") // "11&12"
		r.Gender = strings.TrimSpace(matches[5])             // "女子組"
		// 移除已匹配的部分
		remainingStr = strings.Replace(remainingStr, matches[0], "", 1)
		remainingStr = strings.TrimSpace(remainingStr)
	}
	const expectedRaceNameSplitParts = 3
	matches = strings.Split(remainingStr, " ")
	if len(matches) == expectedRaceNameSplitParts {
		r.EventType = matches[1]
		r.Type = matches[2]
	}
	re := regexp.MustCompile(`^(\d+年)(.*)`)

	matches = re.FindStringSubmatch(strings.ReplaceAll(b.info.CompetitionName, " ", ""))

	if len(matches) < expectedRaceNameSplitParts {
		// 如果沒有匹配或匹配不完整，返回原始字串作為名稱，年份為空
		return nil, errors.New("比賽名稱格式錯誤")
	}

	// matches[1] 是年份部分，例如 "114年"
	rocYear := matches[1]

	// matches[2] 是名稱部分，需要去除可能的首尾空格
	r.CompetitionName = strings.TrimSpace(matches[2])

	// 為了輸出您要求的格式，我們將 "114年" 中的 "年" 去掉，只留下 "114"
	r.Year = strings.TrimSuffix(rocYear, "年")
	results, err := b.getResult()
	if err != nil {
		return nil, err
	}
	r.Results = results
	return &r, nil
}

func (b *raceBuilder) innerText(xpath, xpath2 string) (string, error) {
	recordNode := htmlquery.FindOne(b.doc, xpath)
	if recordNode == nil {
		recordNode = htmlquery.FindOne(b.doc, xpath2)
		if recordNode == nil {
			return "", fmt.Errorf("找不到包含 '大會紀錄' 的元素")
		}
	}
	return htmlquery.InnerText(recordNode), nil
}

func (b *raceBuilder) listElement(xpath, xpath2 string) ([]*html.Node, error) {
	list := htmlquery.Find(b.doc, xpath)
	if list == nil {
		list = htmlquery.Find(b.doc, xpath2)
		if list == nil {
			return nil, fmt.Errorf("找不到任何成績資料行")
		}
	}
	return list, nil
}
