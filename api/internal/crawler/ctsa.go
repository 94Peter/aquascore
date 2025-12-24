package crawler

import (
	"errors"
	"fmt"
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

const apiSleepDuration = 100 * time.Millisecond

func NewCtsaCrawler(opts ...Option) (*ctsaCrawler, error) {
	crawler := &ctsaCrawler{}
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
	_, _ = client.Get("https://ctsa.utk.com.tw/CTSA/public/race/game_data.aspx")
	crawler.client = client
	return crawler, nil
}

type ctsaCrawler struct {
	baseUrl     string
	persistence Persistence
	isTest      bool
	client      *http.Client
}

func (c *ctsaCrawler) Crawl() error {
	raceIDs := c.getRaceIDs()

	if len(raceIDs) == 0 {
		log.Fatal("❌ 未能成功獲取任何比賽 ID，程序終止。")
	}
	fmt.Printf("✅ 成功找到 %d 個比賽 ID，開始逐一 POST 請求...\n", len(raceIDs))
	fmt.Println("---------------------------------------------------------")
	for _, raceID := range raceIDs {
		err := c.postForDetails(raceID)
		if err != nil {
			log.Printf("❌ POST 請求失敗: %v", err)
		} else {
			log.Printf("✅ POST 請求成功: %s", raceID.Name)
		}
		time.Sleep(apiSleepDuration)
		if c.isTest {
			break
		}
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

func (c *ctsaCrawler) getInitialData() (map[string]string, error) {
	resp, err := c.client.Get(c.baseUrl)
	if err != nil {
		return nil, fmt.Errorf("GET 請求失敗: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GET 請求返回非預期狀態碼: %d %s", resp.StatusCode, resp.Status)
	}

	if c.isTest {
		u, _ := url.Parse(c.baseUrl)
		storedCookies := c.client.Jar.Cookies(u)
		fmt.Printf("Jar 中儲存的 Cookie 數量: %d\n", len(storedCookies))
		if len(storedCookies) > 0 {
			fmt.Printf("儲存的 Cookie: %s = %s\n", storedCookies[0].Name, storedCookies[0].Value)
		}
	}

	doc, err := htmlquery.Parse(resp.Body)
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

func (c *ctsaCrawler) getRaceIDs() []activeInfo {
	resp, err := c.client.Get(c.baseUrl)
	if err != nil {
		log.Printf("GET 請求失敗: %v", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Printf("GET 請求返回非預期狀態碼: %d", resp.StatusCode)
		return nil
	}

	doc, err := htmlquery.Parse(resp.Body)
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

func (c *ctsaCrawler) postForDetails(active activeInfo) error {
	hiddenFields, err := c.getInitialData()
	if err != nil {
		return err
	}

	// 構造表單數據。
	// 根據網頁的表單結構，它可能需要發送以下隱藏欄位以及選中的 ddlRace 值：
	// __EVENTTARGET, __EVENTARGUMENT, __VIEWSTATE, __EVENTVALIDATION, ddlRace
	//
	// 這裡我們**只發送**最關鍵的 ddlRace 欄位，在某些簡單的應用中可能可行。
	// 在複雜的 ASP.NET 頁面中，您可能需要先GET頁面來獲取 __VIEWSTATE 和 __EVENTVALIDATION 等隱藏欄位，並將它們包含在 POST 請求體中。
	// 由於複雜度較高，這裡先演示簡單的 POST。

	form := url.Values{}
	// 網站通常用這個欄位來傳遞選擇的比賽 ID
	form.Add("ctl00$ContentPlaceHolder1$DD_Activity_ID", active.ID)
	for k, v := range hiddenFields {
		form.Add(k, v)
	}
	// 如果需要，還可能需要這些：
	// form.Add("__EVENTTARGET", "ddlRace")
	// form.Add("__EVENTARGUMENT", "")
	// form.Add("__VIEWSTATE", "從GET響應中解析出的值")
	// form.Add("__EVENTVALIDATION", "從GET響應中解析出的值")
	req, err := http.NewRequest("POST", c.baseUrl, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("構造請求失敗: %w", err)
	}

	// 設置 HTTP Headers
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", fmt.Sprintf("%d", len(form.Encode())))

	// 發送請求

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("POST 請求失敗: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("POST 請求返回非預期狀態碼: %d %s", resp.StatusCode, resp.Status)
	}

	// 這裡應該是解析返回的 HTML/JSON/資料的邏輯
	// 為了演示，我們僅打印一個成功的標記，並可以讀取響應體（Body）的一部分來驗證
	//
	// 例如：
	// bodyBytes, _ := io.ReadAll(resp.Body)
	// 這裡只打印前 200 字節作為檢查
	// fmt.Printf("   [Body Snippet]: %s...\n", bodyBytes[:200])
	// fmt.Println(string(bodyBytes))

	doc, err := htmlquery.Parse(resp.Body)
	if err != nil {
		return fmt.Errorf("HTML 解析失敗: %w", err)
	}
	xpath := "//table[@id='ctl00_ContentPlaceHolder1_GridView1']/tbody/tr[position() > 1]"
	dataRows := htmlquery.Find(doc, xpath)
	var races []raceInfo
	base, _ := url.Parse(c.baseUrl)
	for _, trNode := range dataRows {
		raceNameNode := htmlquery.FindOne(trNode, "./td[3]//a")

		raceName := "N/A"
		if raceNameNode != nil {
			raceName = htmlquery.InnerText(raceNameNode)
		}

		// b. 獲取成績報告連結
		// 成績報告連結是 <a> 標籤，文字內容為 '成績報告'
		// 使用相對 XPath: 找到當前 tr 內文字為 '成績報告' 的 <a> 標籤
		linkNode := htmlquery.FindOne(trNode, ".//a[text()='成績報告']")

		relativeURL := ""
		absoluteURL := "N/A"

		if linkNode != nil {
			relativeURL = htmlquery.SelectAttr(linkNode, "href")
			pathUrl, _ := url.Parse(relativeURL)
			// 將相對路徑轉換為絕對路徑
			absoluteURL = base.ResolveReference(pathUrl).String()
		}

		// 只有當成功找到成績報告連結時才記錄
		if absoluteURL != "N/A" && strings.TrimSpace(raceName) != "" {
			races = append(races, raceInfo{
				CompetitionName: active.Name,
				RaceName:        strings.TrimSpace(raceName),
				ScoreReportURL:  absoluteURL,
			})
		}
	}
	// 💡 實際應用中:
	// 您應該在這裡使用 htmlquery.Parse(bytes.NewReader(bodyBytes)) 來解析 HTML
	// 然後使用 XPath 查詢新的 HTML 內容，獲取表格數據等詳細資料。

	// 定义最大的并发数
	const maxConcurrency = 5
	// 创建一个容量为 maxConcurrency 的 channel 作为信号量
	// channel 的容量决定了可以同时运行的 goroutine 数量
	semaphore := make(chan struct{}, maxConcurrency)
	// 用于等待所有 goroutine 完成的 WaitGroup
	var wg sync.WaitGroup
	// 用于收集第一个遇到的错误的 channel
	errChan := make(chan error, 1)
	if c.isTest {
		races = races[89:90]
	}

	for _, race := range races {
		select {
		case err := <-errChan:
			// 从 errChan 接收到错误，关闭信号量 channel，停止新的 goroutine 启动
			close(semaphore)
			// 等待已启动的 goroutine 完成（可选，取决于具体需求）
			wg.Wait()
			return err
		default:
			// 没有错误，继续
		}

		wg.Add(1)
		// 尝试发送到信号量 channel。如果 channel 已满（即已有 maxConcurrency 个 goroutine 在运行），
		// 这一步会阻塞，直到有 goroutine 完成并从 channel 接收（释放信号）。
		semaphore <- struct{}{}
		go func(race raceInfo) {
			defer wg.Done()
			// 在 goroutine 结束时释放信号，将一个值从 channel 接收出去，允许新的 goroutine 启动
			defer func() { <-semaphore }()
			ok, err := c.persistence.IsCrawled(race.ScoreReportURL)
			if err != nil {
				sendNonBlockingError(fmt.Errorf("check crawled fail: %w", err), errChan)
				return
			}
			if ok {
				return
			}
			dbrace, err := createRace(race)
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
		}(race)
	}
	// 等待所有 goroutine 完成
	wg.Wait()

	// 关闭 errChan
	close(errChan)

	// 检查是否有错误发生
	if err, ok := <-errChan; ok {
		return err
	}
	return nil
}

func createRace(info raceInfo) (*Race, error) {
	resp, err := http.Get(info.ScoreReportURL)
	if err != nil {
		return nil, fmt.Errorf("GET 請求失敗: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GET 請求返回非預期狀態碼: %d %s", resp.StatusCode, resp.Status)
	}
	doc, err := htmlquery.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("HTML 解析失敗: %w", err)
	}
	race, err := newRaceBuilder(doc, info).CreateRace()
	if err != nil {
		return nil, err
	}

	return race, nil
}

func printNode(n *html.Node, depth int) {
	if depth > 4 {
		return // 限制深度
	}

	// 打印標籤和屬性
	if n.Type == html.ElementNode {
		attrs := ""
		for _, a := range n.Attr {
			attrs += fmt.Sprintf(" %s=%q", a.Key, a.Val)
		}
		fmt.Printf("%s<%s%s>\n", strings.Repeat("  ", depth), n.Data, attrs)
	} else if n.Type == html.TextNode && strings.TrimSpace(n.Data) != "" {
		// 打印非空白文字節點
		fmt.Printf("%s#text: %q\n", strings.Repeat("  ", depth), strings.TrimSpace(n.Data))
	}

	// 遞迴處理子節點
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		printNode(c, depth+1)
	}
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
	if len(parts) != 2 {
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
	re := regexp.MustCompile(`(\d{2}:\d{2}\.\d{2})`)
	matches := re.FindAllString(text, -1)
	if len(matches) == 0 {
		return nil, fmt.Errorf("找不到時間紀錄: %s", text)
	}
	if len(matches) == 1 {
		if strings.Contains(text, "大會紀錄") {
			gameRecordStr := matches[0]
			gameRecord, err := parseTimeDuration(gameRecordStr)
			if err != nil {
				return nil, fmt.Errorf("轉換大會紀錄失敗: %w", err)
			}
			return &raceRecord{gameRecord: gameRecord}, nil
		} else if strings.Contains(text, "全國紀錄") {
			nationalRecordStr := matches[0]
			nationalRecord, err := parseTimeDuration(nationalRecordStr)
			if err != nil {
				return nil, fmt.Errorf("轉換全國紀錄失敗: %w", err)
			}
			return &raceRecord{nationalRecord: nationalRecord}, nil
		}
	}

	var records raceRecord
	gameRecordStr := matches[0]
	nationalRecordStr := matches[1]
	records.gameRecord, err = parseTimeDuration(gameRecordStr)
	if err != nil {
		return nil, fmt.Errorf("轉換大會紀錄失敗: %w", err)
	}

	// 2. 轉換全國紀錄
	records.nationalRecord, err = parseTimeDuration(nationalRecordStr)
	if err != nil {
		return nil, fmt.Errorf("轉換全國紀錄失敗: %w", err)
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
	re := regexp.MustCompile(`(\d{2,3}/\d{2}/\d{2})`)
	match := re.FindStringSubmatch(text)

	if len(match) < 2 {
		return time.Time{}, fmt.Errorf("日期格式錯誤或找不到日期: %s", text)
	}
	datePart := match[1] // 例如: "114/01/11"

	// 2. 轉換民國紀年為西元紀年
	parts := regexp.MustCompile(`/`).Split(datePart, 3)
	if len(parts) != 3 {
		return time.Time{}, fmt.Errorf("日期分割錯誤: %s", datePart)
	}

	rocYearStr := parts[0]
	monthDayPart := parts[1] + "/" + parts[2] // "01/11"

	rocYear, err := strconv.Atoi(rocYearStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("解析民國年失敗: %w", err)
	}

	// 核心轉換邏輯: 西元 = 民國 + 1911
	adYear := rocYear + 1911
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

		if len(tds) < 8 {
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
			rank, err := strconv.Atoi(rankStr)
			if err != nil {
				return nil, fmt.Errorf("convert rank to int failed: %w", err)
			}
			result.Rank = rank
		}

		// 處理 Score (積點)
		if !b.info.IsQualifier() && scoreStr != "" {
			score, err := strconv.Atoi(scoreStr)
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
	if len(matches) > 2 {
		r.AgeGroup = strings.ReplaceAll(matches[1], " ", "") // "11&12"
		r.Gender = strings.TrimSpace(matches[5])             // "女子組"
		// 移除已匹配的部分
		remainingStr = strings.Replace(remainingStr, matches[0], "", 1)
		remainingStr = strings.TrimSpace(remainingStr)
	}
	matches = strings.Split(remainingStr, " ")
	if len(matches) == 3 {
		r.EventType = matches[1]
		r.Type = matches[2]
	}
	re := regexp.MustCompile(`^(\d+年)(.*)`)

	matches = re.FindStringSubmatch(strings.ReplaceAll(b.info.CompetitionName, " ", ""))

	if len(matches) < 3 {
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
