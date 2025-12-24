# AquaScore 🏊‍♂️

AquaScore is a comprehensive swimming performance analysis platform designed to crawl, store, analyze, and visualize swimming competition data. It focuses on data from the Chinese Taipei Swimming Association (CTSA) and provides deep insights into athlete performance, progression, and race comparisons.

## 🚀 Features

-   **Data Crawling**: Automated crawling of swimming race results from official sources (CTSA).
-   **Performance Analysis**:
    -   **Overview**: Track an athlete's personal bests (PB), stability, and performance trends over time.
    -   **Comparison**: Compare specific race results against national records, games records, and competitors.
-   **Visualizations**: Interactive charts for performance trends and race result distributions.
-   **Modern Stack**: Built with a microservices architecture using Go, Python, and React.

## 🏗 Architecture

AquaScore is composed of three main services:

1.  **API Service (`api`)**:
    -   **Language**: Go (Golang)
    -   **Role**: Acts as the central backend gateway and crawler.
    -   **Responsibilities**:
        -   Exposes RESTful endpoints for the frontend.
        -   Crawls and parses raw HTML race data.
        -   Manages data persistence in MongoDB.
        -   Communicates with the Analysis Service via gRPC.
2.  **Analysis Service (`analysis`)**:
    -   **Language**: Python
    -   **Role**: Dedicated gRPC service for data processing.
    -   **Responsibilities**:
        -   Performs statistical analysis (Pandas/NumPy).
        -   Calculates stability scores, trends, and record comparisons.
3.  **Frontend (`frontend`)**:
    -   **Language**: TypeScript, React (Vite)
    -   **Role**: User Interface.
    -   **Responsibilities**:
        -   Displays athlete profiles, race history, and analytical charts.

## 🛠 Tech Stack

-   **Backend (API)**: Go 1.24, Gin, Mongo Driver, OpenTelemetry.
-   **Backend (Analysis)**: Python 3.12, gRPC, Pandas, NumPy.
-   **Frontend**: React 18, TypeScript, Tailwind CSS, Vite.
-   **Database**: MongoDB.
-   **Infrastructure**: Docker, Docker Compose, Buf (Protobuf).
-   **Observability**: OpenTelemetry (Jaeger).

## 🏁 Getting Started

### Prerequisites

-   Docker & Docker Compose
-   Go 1.24+ (for local dev)
-   Python 3.12+ (for local dev)
-   Node.js 20+ (for local dev)

### Running with Docker Compose (Recommended)

To start the entire stack (Database, API, Analysis, Frontend, and Jaeger):

```bash
docker-compose up --build
```

Once running, access the services:

-   **Frontend**: [http://localhost:3000](http://localhost:3000) (or the port defined in docker-compose)
-   **API**: [http://localhost:8080](http://localhost:8080)
-   **Jaeger UI**: [http://localhost:16686](http://localhost:16686)

### Local Development

#### 1. Database
Start a local MongoDB instance:
```bash
docker run -d -p 27017:27017 --name mongodb mongo:latest
```

#### 2. Analysis Service (Python)
```bash
cd analysis
python -m venv env
source env/bin/activate
pip install -r requirements.txt
# Run the gRPC server
python -m server
```

#### 3. API Service (Go)
```bash
cd api
# Ensure .aquascore.yaml is configured correctly for localhost
go run main.go server
```
*To run the crawler:*
```bash
go run main.go crawler --year 114
```

#### 4. Frontend (React)
```bash
cd frontend
npm install
npm run dev
```

## 📂 Project Structure

```
AquaScore/
├── api/            # Go backend & Crawler
├── analysis/       # Python gRPC analysis service
├── frontend/       # React application
├── proto/          # Protocol Buffer definitions
├── api/cmd/        # CLI commands (server, crawler)
└── ...
```

## 📝 License

[MIT](LICENSE)