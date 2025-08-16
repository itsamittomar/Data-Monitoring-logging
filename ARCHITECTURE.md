# Distributed Logging & Monitoring System

## Architecture Overview

This system provides a complete distributed logging and monitoring solution with four main components:

```
┌─────────────┐    ┌──────────────┐    ┌─────────────────┐
│   Applications  │    │     Kafka        │    │   Log Ingestion   │
│   (Log Sources) │───▶│    Message       │───▶│    Service        │
│                 │    │    Queue         │    │                   │
└─────────────────┘    └──────────────────┘    └─────────┬─────────┘
                                                         │
                                                         ▼
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Dashboard     │    │    Alerting      │    │  ElasticSearch  │
│   Backend       │    │    System        │    │    Storage      │
│   (REST API)    │    │  (Rules Engine)  │    │                 │
└─────────┬───────┘    └─────────┬────────┘    └─────────────────┘
          │                      │
          ▼                      ▼
┌─────────────────┐    ┌──────────────────┐
│   Metrics       │    │    Alert         │
│   Aggregation   │    │    Notifications │
│   System        │    │                  │
└─────────────────┘    └──────────────────┘
```

## Data Flow

### 1. Log Ingestion Flow
```
Application Logs → Kafka Topic → Log Ingestion Service → ElasticSearch
```

1. **Applications** send structured log entries to Kafka topics
2. **Kafka** acts as a message buffer, ensuring no log loss during high load
3. **Log Ingestion Service** consumes from Kafka in batches
4. **ElasticSearch** stores logs with daily indices for efficient querying

### 2. Metrics Collection Flow
```
Metrics Sources → Metrics Service → In-Memory Time Series → Aggregation Functions
```

1. **Metrics** are collected from various sources (applications, infrastructure)
2. **Metrics Service** stores time series data in memory with automatic pruning
3. **Aggregation Functions** calculate sum, avg, max, percentiles, and rates

### 3. Alerting Flow
```
Alert Rules → Rules Engine → Metric Queries → Alert Evaluation → Notifications
```

1. **Alert Rules** define conditions using query language
2. **Rules Engine** evaluates rules at regular intervals
3. **Metric Queries** fetch aggregated data from metrics system
4. **Alert Evaluation** compares values against thresholds
5. **Notifications** are triggered when conditions are met

### 4. Dashboard Flow
```
Frontend → REST API → ElasticSearch/Metrics → Response Data
```

1. **Frontend** makes HTTP requests to dashboard API
2. **REST API** handles queries for logs and metrics
3. **Data Sources** (ElasticSearch/Metrics) provide requested data
4. **Response** returns formatted JSON data

## Component Details

### Log Ingestion Service (`cmd/ingestion`)
- **Purpose**: Consume logs from Kafka and store in ElasticSearch
- **Key Features**:
  - Batch processing for high throughput (100 logs per batch or 5-second intervals)
  - Graceful shutdown handling
  - Automatic ElasticSearch index creation
  - Daily index rotation (logs-2024.01.01 format)

### Metrics Service (`cmd/metrics`)
- **Purpose**: Collect and aggregate metrics data
- **Key Features**:
  - Prometheus-compatible metrics collection
  - In-memory time series storage
  - Multiple aggregation functions (sum, avg, max, percentiles, rate)
  - Automatic data pruning (24-hour retention)

### Alerting Service (`cmd/alerting`)
- **Purpose**: Monitor metrics and trigger alerts based on rules
- **Key Features**:
  - JSON-based rule configuration
  - Flexible query language for metric evaluation
  - Configurable check intervals
  - Alert status tracking

### Dashboard Service (`cmd/dashboard`)
- **Purpose**: Provide REST API for frontend dashboards
- **Key Features**:
  - RESTful endpoints for logs and metrics
  - CORS support for web frontends
  - Request logging middleware
  - Health check endpoints

## Configuration

All services use a shared `config.yaml` file:

```yaml
kafka:
  brokers: ["localhost:9092"]
  topic: "logs"

elasticsearch:
  urls: ["http://localhost:9200"]
  index: "logs"

metrics:
  port: 9090
  retention_days: 15

alerting:
  rules_path: "alert_rules.json"
  check_interval: "30s"

dashboard:
  port: 8080
```

## Alert Rules Format

Alert rules are defined in JSON format (`alert_rules.json`):

```json
{
  "name": "HighErrorRate",
  "query": "rate error_count service=api",
  "threshold": 10.0,
  "operator": ">",
  "duration": "5m",
  "labels": {"severity": "critical"},
  "annotations": {"summary": "High error rate detected"}
}
```

### Query Language
- `sum metric_name [labels]` - Sum of values
- `avg metric_name [labels]` - Average of values  
- `max metric_name [labels]` - Maximum value
- `rate metric_name [labels]` - Rate per second
- `p95 metric_name [labels]` - 95th percentile
- `p99 metric_name [labels]` - 99th percentile

## API Endpoints

### Dashboard API (`localhost:8080`)
- `GET /api/v1/logs/search` - Search logs with filters
- `GET /api/v1/metrics/query` - Query single metric value
- `GET /api/v1/metrics/range` - Query metric time series
- `GET /api/v1/health` - Health check

### Metrics API (`localhost:9090`)
- `GET /metrics` - Prometheus-compatible metrics endpoint
- `GET /api/metrics` - Custom metrics API
- `GET /api/query` - Query interface

## Deployment

### Local Development
```bash
# Build all services
make build

# Run individual services
make run-ingestion
make run-metrics
make run-alerting
make run-dashboard
```

### Docker Deployment
```bash
# Start complete stack
docker-compose up

# Scale individual services
docker-compose up --scale log-ingestion=3
```

## Data Models

### Log Entry
```go
type LogEntry struct {
    Timestamp  time.Time
    Level      string
    Message    string
    Service    string
    Host       string
    Tags       map[string]string
    Fields     map[string]interface{}
}
```

### Metric
```go
type Metric struct {
    Name      string
    Value     float64
    Timestamp time.Time
    Labels    map[string]string
    Type      string
}
```

### Alert
```go
type Alert struct {
    Rule      AlertRule
    Value     float64
    Timestamp time.Time
    Status    string
    Labels    map[string]string
}
```

## Scalability Features

1. **Horizontal Scaling**: Each service can be scaled independently
2. **Batch Processing**: Optimized for high-throughput log ingestion
3. **Memory Management**: Automatic pruning of old metrics data
4. **Load Balancing**: Kafka consumer groups distribute load
5. **Index Rotation**: Daily ElasticSearch indices for efficient queries

## Monitoring & Observability

- **Health Checks**: Each service exposes health endpoints
- **Structured Logging**: All services use structured logging with logrus
- **Metrics Exposure**: Services expose their own metrics for monitoring
- **Graceful Shutdown**: Proper cleanup on service termination