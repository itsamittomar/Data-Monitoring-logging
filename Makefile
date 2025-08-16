.PHONY: build clean run-ingestion run-metrics run-alerting run-dashboard deps

BINARY_DIR=bin

build: deps
	@mkdir -p $(BINARY_DIR)
	go build -o $(BINARY_DIR)/ingestion ./cmd/ingestion
	go build -o $(BINARY_DIR)/metrics ./cmd/metrics
	go build -o $(BINARY_DIR)/alerting ./cmd/alerting
	go build -o $(BINARY_DIR)/dashboard ./cmd/dashboard

deps:
	go mod tidy
	go mod download

clean:
	rm -rf $(BINARY_DIR)

run-ingestion: build
	./$(BINARY_DIR)/ingestion

run-metrics: build
	./$(BINARY_DIR)/metrics

run-alerting: build
	./$(BINARY_DIR)/alerting

run-dashboard: build
	./$(BINARY_DIR)/dashboard

docker-build:
	docker build -t logging-system-ingestion -f docker/Dockerfile.ingestion .
	docker build -t logging-system-metrics -f docker/Dockerfile.metrics .
	docker build -t logging-system-alerting -f docker/Dockerfile.alerting .
	docker build -t logging-system-dashboard -f docker/Dockerfile.dashboard .

test:
	go test ./...