# technical_spec.md

## 1. 總覽 (Overview)

本文件旨在定義 AquaScore 儀表板後端系統的技術架構、組件和實現方法。後端將採用 Golang 作為 API 層，並透過 gRPC 與 Python 數據分析服務進行通訊，充分利用 Python 強大的數據分析生態系來實現 wireframe 中定義的核心功能。數據庫將採用 MongoDB。

本文檔的目標讀者為後端開發工程師，旨在提供一個清晰的開發藍圖。

## 2. 系統架構 (System Architecture)

我們將採用現代化的、前後端分離的微服務架構。新的數據流將由 Golang API 層負責與數據庫進行所有交互。

*   **API 層 (API Layer)**: 使用 **Golang**。Go 負責處理前端請求、直接從 **MongoDB** 讀取數據。它將獲取到的原始數據透過 gRPC 傳遞給 Python 服務進行分析，並將最終結果回傳給前端。
*   **數據分析服務 (Data Analysis Service)**: 使用 **Python**。此服務作為一個純粹的計算引擎，接收來自 Go API 層的數據，利用 Pandas/NumPy 等函式庫進行核心的數據計算，並將分析結果返回。**它不再直接與數據庫互動**。
*   **前端 (Frontend)**: 任何現代 JavaScript 框架（如 React, Vue, or Svelte）。前端負責渲染 UI，並通過 RESTful API 與 Go API 層進行數據交換。
*   **數據庫 (Database)**: **MongoDB**。由 Golang API 層進行讀取和寫入。

### 架構圖

```
+----------------+      +------------------------+      +-----------------------+ -----(gRPC)-----> +-------------------------------+
|                |      |                        |      |                       |                  |                               |
|  用戶瀏覽器     |-(REST)->|  前端 (React/Vue.js)   |-(REST)->|    API 層 (Golang)    |                  |  數據分析服務 (Python)        |
| (Web Browser)  |      |                        |      |                       |                  |  (Pandas, NumPy, gRPC Server) |
|                |      |                        |      +-----------+-----------+                  |                               |
+----------------+      +------------------------+                  |                              +-------------------------------+
                                                                        |
                                                                        | (Driver)
                                                                        v
                                                              +----------------+
                                                              |                |
                                                              |  數據庫        |
                                                              |  (MongoDB)     |
                                                              |                |
                                                              +----------------+
```

## 3. 後端技術棧 (Backend Technology Stack)

### 3.1 API 層 (Golang)

| 組件 | 技術/函式庫 | 理由 |
| :--- | :--- | :--- |
| **語言** | Golang | 高效能、高併發，適合處理 API 請求和數據庫 I/O。 |
| **Web 框架** | Gin / net/http | 輕量級、高效能的 Web 框架。 |
| **數據庫互動** | `go.mongodb.org/mongo-driver` | 官方的 MongoDB Go 驅動程式，用於與數據庫進行高效能的數據操作。|
| **gRPC** | `google.golang.org/grpc` | 用於與 Python 數據分析服務进行高效、類型安全的 RPC 通訊。 |

### 3.2 數據分析服務 (Python)

| 組件 | 技術/函式庫 | 理由 |
| :--- | :--- | :--- |
| **語言** | Python 3.11+ | 利用其成熟的數據分析生態系。 |
| **gRPC 框架** | `grpcio` | 實現 gRPC 伺服器，接收來自 Go API 層的數據。 |
| **數據分析** | Pandas, NumPy | 核心組件，用於實現所有性能指標計算。 |
| **數據模型** | Pydantic | (可選) 用於在服務內部驗證傳入數據的結構。 |


## 4. API 設計 (API and gRPC Design)

### 4.1 RESTful API 端點 (由 Go 提供)

The Go API layer provides the following RESTful endpoints. For detailed specifications, including request/response formats and schemas, please see the `openapi.yaml` file.

*   `GET /athletes`: Fetches a list of all athletes.
*   `GET /years`: Fetches a list of available competition years.
*   `GET /competitions?year={year}`: Fetches competitions for a specific year.
*   `GET /athletes/{athlete_name}/races?competition_name={competition_name}&year={year}`: Fetches all race results for a specific athlete in a given competition and year.
*   `GET /athletes/{athlete_name}/performance-overview`: Fetches a detailed performance analysis for an athlete.
*   `GET /race/{race_id}/comparison`: Fetches a comparison analysis for a specific race.

### 4.2 gRPC 服務定義 (Python Service)

Go API 層與 Python 數據分析服務之間透過 gRPC 進行通訊。Python 服務的方法現在接收結構化的 Protobuf 訊息，而不是序列化的 JSON 字串。

```protobuf
syntax = "proto3";

package analysis;

import "google/protobuf/timestamp.proto";

option go_package = "aquascore/internal/generated/analysis;analysis";

// AnalysisService 定義了需要複雜計算的數據分析 RPC
service AnalysisService {
    // 分析運動員的整體表現 (PB, 趨勢, 穩定度等)
    rpc AnalyzePerformanceOverview(AnalyzePerformanceOverviewRequest) returns (AnalyzePerformanceOverviewResponse);

    // 分析單一成績與其他選手、紀錄的比較
    rpc AnalyzeResultComparison(AnalyzeResultComparisonRequest) returns (AnalyzeResultComparisonResponse);
}

// --- 用於整體表現分析的訊息 ---

// 傳入的單筆成績數據
message PerformanceResult {
    google.protobuf.Timestamp event_date = 1;
    double result_time = 2;
    string event_name = 3; // e.g., "50公尺自由式"
}

message AnalyzePerformanceOverviewRequest {
    string athlete_name = 1; // 運動員名稱
    repeated PerformanceResult results = 2; // 該運動員的所有成績列表 (可包含多個項目)
}

// 返回的單一項目表現分析結果
message EventPerformanceAnalysis {
    string event_name = 1;
    double personal_best = 2;
    google.protobuf.Timestamp personal_best_date = 3;
    double stability = 4; // 穩定度 (變異係數)
    double trend = 5;     // 近期趨勢 (與PB的秒差)
}

message AnalyzePerformanceOverviewResponse {
    repeated EventPerformanceAnalysis event_analyses = 1;
}

// --- 用於成績比較分析的訊息 ---

// 傳入的單筆比賽成績
message RaceResult {
    string athlete_name = 1;
    double record_time = 2; // 秒數
    int32 rank = 3;         // 名次
}

// 傳入的紀錄標記 (全國紀錄/大會紀錄)
message RecordMarks {
    optional double national_record = 1;
    optional double games_record = 2;
}

message AnalyzeResultComparisonRequest {
    RaceResult target_result = 1;                   // 當前要比較的成績紀錄
    repeated RaceResult competition_results = 2;    // 該場比賽所有選手的成績
    RecordMarks records = 3;                        // 相關的紀錄
}

// 返回的單一選手成績對比結果
message SingleResultComparison {
    string athlete_name = 1;
    double record_time = 2;
    int32 rank = 3;
    // 使用 optional，如果紀錄不存在，則客戶端不會收到該字段
    optional double diff_from_national_record = 4; // 與全國紀錄的秒差
    optional double diff_from_games_record = 5;    // 與大會紀錄的秒差
}

message AnalyzeResultComparisonResponse {
    repeated SingleResultComparison results_comparison = 1;
}
```

## 5. 核心功能實現邏輯 (使用 Pandas)

後端的 **Python 數據分析服務**在從 gRPC 請求中獲取數據後，將使用 Pandas DataFrame 進行核心指標的計算。

#### 範例：計算 50 公尺自由式的各項指標

```python
import pandas as pd
import json

# 1. 在 gRPC 服務的實現中，從 request 的 JSON 字串解析數據，載入 DataFrame
# def AnalyzePerformanceOverview(self, request, context):
#     data = json.loads(request.results_json)
#     df = pd.DataFrame(data)

# 假設 df 結構如下：
#    event_date    result_time
# 0  2024-10-19    24.98
# 1  2024-08-15    25.30
# 2  2024-07-01    25.10
# ...

# 確保 'event_date' 是 datetime 類型，以便排序
df['event_date'] = pd.to_datetime(df['event_date'])

# 2. 計算個人最佳 (PB)
pb = df['result_time'].min()  # -> 24.98

# 3. 計算穩定度 (以變異係數 CV 為例)
last_n_races = df.sort_values('event_date', ascending=False).head(5)
std_dev = last_n_races['result_time'].std()
mean_val = last_n_races['result_time'].mean()
stability_cv = (std_dev / mean_val) * 100 if mean_val != 0 else 0

# 4. 計算近三場趨勢
last_3_races = df.sort_values('event_date', ascending=False).head(3)
last_3_mean = last_3_races['result_time'].mean()
trend_vs_pb = last_3_mean - pb  # -> e.g., +0.33s

# 5. 將所有計算結果組合成 JSON 字串，並在 gRPC 回應中返回
# result_payload = {"pb": pb, "stability": stability_cv, "trend": trend_vs_pb}
# return AnalyzePerformanceOverviewResponse(overview_json=json.dumps(result_payload))
```

## 6. 數據庫模型 (MongoDB Document Schema)

這些模型定義了存儲在 MongoDB 中的文檔結構。Golang API 層將使用對應的 struct 來從 MongoDB 讀取數據，然後將其序列化以通過 gRPC 傳遞給 Python 服務。

#### Go Struct 範例

```go
// 在 Go 中，你可以定義以下 struct 來映射 BSON 文檔
package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Race struct {
    ID              primitive.ObjectID `bson:"_id,omitempty"`
    Name            string             `bson:"name"`
    Year            string             `bson:"year"`
    Date            primitive.DateTime `bson:"date"`
    Event           string             `bson:"event"`
    EventType       string             `bson:"event_type"`
    AgeGroup        string             `bson:"age_group,omitempty"`
    Gender          string             `bson:"gender,omitempty"`
    NationalRecord  float64            `bson:"national_record,omitempty"`
    GamesRecord     float64            `bson:"games_record,omitempty"`
}

type RaceResult struct {
    ID      primitive.ObjectID `bson:"_id,omitempty"`
    RaceID  primitive.ObjectID `bson:"race_id"`
    Name    string             `bson:"name"`
    Record  float64            `bson:"record"`
    Rank    int32              `bson:"rank,omitempty"`
}
```

## 7. 開發與部署建議

*   **環境管理**:
    *   **Golang**: 使用 Go Modules (`go.mod`) 管理依賴。
    *   **Python**: 使用 Poetry 或標準的 `venv` 來管理專案依賴。
*   **gRPC 開發流程**:
    *   在 `.proto` 檔案中定義服務和訊息。
    *   使用 `protoc` 編譯器為 Golang 和 Python 生成客戶端和伺服器代碼。
*   **測試**:
    *   **Golang**: 使用內建的 `testing` 套件為 API 端點和數據庫交互編寫單元測試與整合測試。
    *   **Python**: 使用 `pytest` 框架，並針對核心數據分析邏輯（接收模擬數據）進行單元測試。
*   **部署**:
    *   **容器化**: 將 Go API 服務和 Python 數據分析服務分別打包成獨立的 Docker image。
    *   **服務運行**:
        *   **Go**: 直接運行編譯後的二進制文件。
        *   **Python**: 運行 gRPC 伺服器。
    *   **服務協調**: 使用 Docker Compose 或 Kubernetes 來管理和串連這兩個服務以及 MongoDB 數據庫。
    *   **MongoDB 部署**: 考慮使用 MongoDB Atlas (雲端服務) 或自行搭建副本集 (Replica Set) 以確保高可用性和數據持久性。
