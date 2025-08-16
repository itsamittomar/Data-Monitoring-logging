# Distributed Logging & Monitoring System

A production-ready distributed logging and monitoring system built in Go, featuring real-time log ingestion, metrics aggregation, alerting, and dashboard APIs.

## 🚀 Features

- **Log Ingestion Pipeline**: Kafka → ElasticSearch with batch processing
- **Metrics Aggregation**: Prometheus-like time series collection and aggregation
- **Alerting System**: Rule-based monitoring with configurable thresholds
- **Dashboard Backend**: RESTful APIs for frontend integration
- **Scalable Architecture**: Microservices design with independent scaling
- **Production Ready**: Docker deployment, health checks, graceful shutdown

## 📋 Table of Contents

- [Architecture](#architecture)
- [Quick Start](#quick-start)
- [Services](#services)
- [Configuration](#configuration)
- [API Documentation](#api-documentation)
- [Alert Rules](#alert-rules)
- [Deployment](#deployment)
- [Development](#development)
- [Data Models](#data-models)

## 🏗️ Architecture

```
┌─────────────────┐    ┌──────────────┐    ┌─────────────────┐
│   Applications  │    │     Kafka    │    │  Log Ingestion  │
│   (Log Sources) │───▶│   Message    │───▶│    Service      │
│                 │    │    Queue     │    │                 │
└─────────────────┘    └──────────────┘    └─────────┬───────┘
                                                     │
                                                     ▼
┌─────────────────┐    ┌──────────────┐    ┌─────────────────┐
│   Dashboard     │    │   Alerting   │    │ ElasticSearch   │
│   Backend       │    │   System     │    │   Storage       │
│   (REST API)    │    │(Rules Engine)│    │                 │
└─────────┬───────┘    └──────┬───────┘    └─────────────────┘
          │                   │
          ▼                   ▼
┌─────────────────┐    ┌──────────────┐
│   Metrics       │    │    Alert     │
│  Aggregation    │    │Notifications │
│   System        │    │              │
└─────────────────┘    └──────────────┘
```

### Component Flow

1. **Log Ingestion**: Applications → Kafka → Ingestion Service → ElasticSearch
2. **Metrics Collection**: Sources → Metrics Service → Time Series Storage
3. **Alerting**: Rules Engine → Metric Queries → Alert Evaluation → Notifications
4. **Dashboard**: Frontend → REST API → Data Sources → JSON Response

## 🚀 Quick Start

### Prerequisites

- Go 1.24+
- Docker & Docker Compose
- Make (optional)

### 1. Clone and Setup

```bash
git clone <repository-url>
cd awesomeProject6
```

### 2. Start Infrastructure

```bash
# Start Kafka, ElasticSearch, and Kibana
docker-compose up kafka elasticsearch kibana -d

# Wait for services to be ready (30-60 seconds)
```

### 3. Build Services

```bash
# Using Make
make build

# Or manually
go mod tidy
go build -o bin/ingestion ./cmd/ingestion
go build -o bin/metrics ./cmd/metrics
go build -o bin/alerting ./cmd/alerting
go build -o bin/dashboard ./cmd/dashboard
```

### 4. Run Services

```bash
# Terminal 1: Log Ingestion
./bin/ingestion

# Terminal 2: Metrics Service
./bin/metrics

# Terminal 3: Alerting Service
./bin/alerting

# Terminal 4: Dashboard API
./bin/dashboard
```

### 5. Verify Setup

```bash
# Check dashboard health
curl http://localhost:8080/api/v1/health

# Check metrics endpoint
curl http://localhost:9090/metrics
```

## 🔧 Services

### Log Ingestion Service (`cmd/ingestion`)

**Purpose**: Consumes logs from Kafka and stores them in ElasticSearch

**Features**:
- Batch processing (100 logs/batch or 5s timeout)
- Automatic index creation with daily rotation
- Graceful shutdown with pending batch processing
- Consumer group for load distribution

**Process Flow**:
1. Connect to Kafka consumer group
2. Read log entries from configured topic
3. Parse JSON log entries into structured format
4. Batch logs for efficient ElasticSearch indexing
5. Create daily indices (e.g., `logs-2024.08.16`)

### Metrics Service (`cmd/metrics`)

**Purpose**: Collect, store, and aggregate metrics data

**Features**:
- In-memory time series storage
- Multiple aggregation functions
- Automatic data pruning (24h retention)
- Prometheus-compatible endpoints

**Aggregation Functions**:
- `Sum`: Total value over time period
- `Average`: Mean value over time period
- `Max`: Maximum value in time period
- `Rate`: Events per second
- `Percentile`: P95, P99 calculations

### Alerting Service (`cmd/alerting`)

**Purpose**: Monitor metrics and trigger alerts based on rules

**Features**:
- JSON-based rule configuration
- Flexible query language
- Configurable evaluation intervals
- Alert state management

**Alert Evaluation Process**:
1. Load alert rules from JSON file
2. Execute metric queries using aggregation functions
3. Compare results against thresholds
4. Generate alerts when conditions are met
5. Send notifications through alert channel

### Dashboard Service (`cmd/dashboard`)

**Purpose**: Provide REST API for frontend dashboards

**Features**:
- RESTful endpoints for logs and metrics
- CORS support for web frontends
- Request logging and monitoring
- Health check endpoints

## ⚙️ Configuration

### Main Configuration (`config.yaml`)

```yaml
kafka:
  brokers:
    - "localhost:9092"
  topic: "logs"

elasticsearch:
  urls:
    - "http://localhost:9200"
  index: "logs"
  username: ""
  password: ""

metrics:
  port: 9090
  path: "/metrics"
  retention_days: 15

alerting:
  rules_path: "alert_rules.json"
  check_interval: "30s"

dashboard:
  port: 8080
```

### Environment Variables

- `CONFIG_PATH`: Path to config.yaml (default: "./config.yaml")
- `LOG_LEVEL`: Logging level (debug, info, warn, error)
- `KAFKA_BROKERS`: Comma-separated Kafka broker addresses
- `ELASTICSEARCH_URL`: ElasticSearch connection URL

## 📊 API Documentation

### Dashboard API (Port 8080)

#### Search Logs
```http
GET /api/v1/logs/search?service=api&level=error&from=1692172800&to=1692176400&limit=100
```

**Parameters**:
- `service`: Filter by service name
- `level`: Filter by log level (debug, info, warn, error)
- `from`: Start timestamp (Unix)
- `to`: End timestamp (Unix)
- `limit`: Maximum results (default: 100)

**Response**:
```json
{
  "logs": [],
  "total": 0,
  "query": {
    "service": "api",
    "level": "error",
    "from": "1692172800",
    "to": "1692176400",
    "limit": "100"
  }
}
```

#### Query Metrics
```http
GET /api/v1/metrics/query?metric=cpu_usage&function=avg&duration=1h&service=api
```

**Parameters**:
- `metric`: Metric name
- `function`: Aggregation function (sum, avg, max, rate, p95, p99)
- `duration`: Time window (e.g., "1h", "30m", "5s")
- Additional label filters as query parameters

**Response**:
```json
{
  "metric": "cpu_usage",
  "function": "avg",
  "value": 45.2,
  "duration": "1h",
  "labels": {"service": "api"}
}
```

#### Query Metrics Range
```http
GET /api/v1/metrics/range?metric=memory_usage&from=1692172800&to=1692176400&step=60
```

**Parameters**:
- `metric`: Metric name
- `from`: Start timestamp (Unix)
- `to`: End timestamp (Unix)
- `step`: Step interval in seconds

#### Health Check
```http
GET /api/v1/health
```

**Response**:
```json
{
  "status": "healthy",
  "timestamp": 1692176400,
  "version": "1.0.0"
}
```

### Metrics API (Port 9090)

#### Prometheus Metrics
```http
GET /metrics
```
Returns Prometheus-formatted metrics

## 🚨 Alert Rules

### Rule Configuration

Create alert rules in `alert_rules.json`:

```json
[
  {
    "name": "HighErrorRate",
    "query": "rate error_count service=api",
    "threshold": 10.0,
    "operator": ">",
    "duration": "5m",
    "labels": {
      "severity": "critical",
      "team": "backend"
    },
    "annotations": {
      "summary": "High error rate detected",
      "description": "Error rate is above 10 errors per second for the API service"
    }
  }
]
```

### Query Language

Alert queries use a simple syntax:
```
<function> <metric_name> [label=value] [label2=value2]
```

**Examples**:
- `sum request_count service=api` - Total requests for API service
- `avg cpu_usage host=web01` - Average CPU for specific host
- `p95 response_time` - 95th percentile response time across all services
- `rate error_count` - Error rate per second

### Operators

- `>`: Greater than
- `>=`: Greater than or equal
- `<`: Less than
- `<=`: Less than or equal
- `==`: Equal to
- `!=`: Not equal to

## 🐳 Deployment

### Docker Compose (Recommended)

```bash
# Start entire stack
docker-compose up -d

# View logs
docker-compose logs -f

# Scale services
docker-compose up --scale log-ingestion=3 -d

# Stop services
docker-compose down
```

### Manual Docker Build

```bash
# Build images
make docker-build

# Run individual containers
docker run -d --name ingestion logging-system-ingestion
docker run -d --name dashboard -p 8080:8080 logging-system-dashboard
```

### Kubernetes Deployment

```yaml
# Example deployment for ingestion service
apiVersion: apps/v1
kind: Deployment
metadata:
  name: log-ingestion
spec:
  replicas: 3
  selector:
    matchLabels:
      app: log-ingestion
  template:
    metadata:
      labels:
        app: log-ingestion
    spec:
      containers:
      - name: ingestion
        image: logging-system-ingestion:latest
        env:
        - name: KAFKA_BROKERS
          value: "kafka:9092"
```

## 💻 Development

### Project Structure

```
awesomeProject6/
├── cmd/                    # Main applications
│   ├── ingestion/         # Log ingestion service
│   ├── metrics/           # Metrics collection service
│   ├── alerting/          # Alerting service
│   └── dashboard/         # Dashboard API service
├── pkg/                   # Shared packages
│   ├── kafka/            # Kafka consumer
│   ├── elasticsearch/    # ElasticSearch client
│   ├── prometheus/       # Metrics collection & aggregation
│   ├── rules/            # Alert rules engine
│   └── api/              # HTTP handlers
├── internal/             # Internal packages
│   ├── config/          # Configuration management
│   └── models/          # Data models
├── docker/              # Docker files
├── config.yaml          # Main configuration
├── alert_rules.json     # Alert rules
└── docker-compose.yaml  # Docker Compose setup
```

### Building

```bash
# Install dependencies
go mod tidy

# Build all services
make build

# Build individual service
go build -o bin/dashboard ./cmd/dashboard

# Run tests
make test
```

### Adding New Features

1. **New Metric Types**: Extend `pkg/prometheus/collector.go`
2. **New Alert Functions**: Add to `pkg/rules/engine.go`
3. **New API Endpoints**: Add to `pkg/api/handlers.go`
4. **New Data Sources**: Implement in respective `pkg/` directory

## 📊 Data Models

### Log Entry Structure
```go
type LogEntry struct {
    Timestamp  time.Time              `json:"timestamp"`  // When log was created
    Level      string                 `json:"level"`      // debug, info, warn, error
    Message    string                 `json:"message"`    // Log message content
    Service    string                 `json:"service"`    // Service identifier
    Host       string                 `json:"host"`       // Host/container name
    Tags       map[string]string      `json:"tags"`       // Key-value tags
    Fields     map[string]interface{} `json:"fields"`     // Additional structured data
}
```

### Metric Structure
```go
type Metric struct {
    Name      string            `json:"name"`      // Metric identifier
    Value     float64           `json:"value"`     // Numeric value
    Timestamp time.Time         `json:"timestamp"` // When metric was recorded
    Labels    map[string]string `json:"labels"`    // Metric labels/dimensions
    Type      string           `json:"type"`      // counter, gauge, histogram
}
```

### Alert Rule Structure
```go
type AlertRule struct {
    Name        string            `json:"name"`        // Rule identifier
    Query       string           `json:"query"`       // Metric query expression
    Threshold   float64          `json:"threshold"`   // Alert threshold value
    Operator    string           `json:"operator"`    // Comparison operator
    Duration    string           `json:"duration"`    // Evaluation window
    Labels      map[string]string `json:"labels"`      // Alert labels
    Annotations map[string]string `json:"annotations"` // Alert metadata
}
```

## 🔍 Monitoring & Operations

### Service Health Monitoring

Each service exposes health endpoints:
- **Dashboard**: `http://localhost:8080/api/v1/health`
- **Metrics**: `http://localhost:9090/api/health`

### Log Analysis

Access Kibana dashboard at `http://localhost:5601` for log visualization and analysis.

### Metrics Visualization

The metrics service exposes Prometheus-compatible endpoints at `/metrics` for integration with Grafana or other visualization tools.

### Alert Management

1. **View Active Alerts**: Monitor service logs for alert notifications
2. **Modify Rules**: Edit `alert_rules.json` and restart alerting service
3. **Test Rules**: Use the metrics query API to verify rule logic

## 🔧 Troubleshooting

### Common Issues

**Kafka Connection Failed**
```bash
# Check Kafka is running
docker-compose ps kafka

# Verify Kafka topics
docker-compose exec kafka kafka-topics --list --bootstrap-server localhost:9092
```

**ElasticSearch Connection Failed**
```bash
# Check ElasticSearch health
curl http://localhost:9200/_health

# View ElasticSearch logs
docker-compose logs elasticsearch
```

**High Memory Usage**
- Adjust batch sizes in ingestion service
- Reduce metrics retention period in config
- Scale services horizontally

### Performance Tuning

**Log Ingestion**:
- Increase batch size for higher throughput
- Reduce batch timeout for lower latency
- Scale ingestion service replicas

**Metrics Storage**:
- Adjust retention period based on memory constraints
- Implement external storage for long-term retention
- Use metric sampling for high-cardinality data

**Alert Evaluation**:
- Increase check interval to reduce CPU usage
- Optimize alert queries for better performance
- Use alert grouping to reduce notification spam

## 🧪 Testing

### Unit Tests
```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./pkg/prometheus
go test ./pkg/rules
```

### Integration Tests
```bash
# Start test environment
docker-compose -f docker-compose.test.yaml up -d

# Run integration tests
go test -tags=integration ./...
```

### Load Testing
```bash
# Test log ingestion throughput
go run scripts/load_test_logs.go

# Test metrics query performance
go run scripts/load_test_metrics.go
```

## 📈 Scaling

### Horizontal Scaling

**Log Ingestion**:
```bash
# Scale ingestion service
docker-compose up --scale log-ingestion=5
```

**Dashboard API**:
```bash
# Scale dashboard backend
docker-compose up --scale dashboard=3
```

### Vertical Scaling

**Memory Optimization**:
- Adjust batch sizes in `config.yaml`
- Tune ElasticSearch heap settings
- Configure metrics retention period

**CPU Optimization**:
- Increase alert check intervals
- Optimize ElasticSearch queries
- Use connection pooling

## 🔐 Security

### Authentication & Authorization

- Configure ElasticSearch authentication in `config.yaml`
- Implement API key authentication for dashboard endpoints
- Use TLS/SSL for production deployments

### Network Security

- Deploy services in private networks
- Use service mesh for inter-service communication
- Implement rate limiting on public endpoints

## 📝 Contributing

### Development Workflow

1. Fork the repository
2. Create feature branch: `git checkout -b feature/new-feature`
3. Make changes and add tests
4. Run tests: `make test`
5. Build services: `make build`
6. Submit pull request

### Code Style

- Follow Go conventions and `gofmt` formatting
- Add unit tests for new functionality
- Update documentation for API changes
- Use structured logging with contextual fields

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🤝 Support

For issues and questions:
- Check the troubleshooting section above
- Review service logs for error details
- Open an issue with reproduction steps
- Include configuration and environment details