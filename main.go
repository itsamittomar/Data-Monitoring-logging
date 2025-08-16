package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Distributed Logging & Monitoring System")
		fmt.Println("Usage:")
		fmt.Println("  go run cmd/ingestion/main.go    - Start log ingestion service")
		fmt.Println("  go run cmd/metrics/main.go      - Start metrics service")
		fmt.Println("  go run cmd/alerting/main.go     - Start alerting service")
		fmt.Println("  go run cmd/dashboard/main.go    - Start dashboard service")
		fmt.Println("\nOr use the Makefile:")
		fmt.Println("  make build                      - Build all services")
		fmt.Println("  make run-ingestion             - Run ingestion service")
		fmt.Println("  make run-dashboard             - Run dashboard service")
		fmt.Println("\nOr use Docker Compose:")
		fmt.Println("  docker-compose up              - Start all services")
		return
	}

	fmt.Printf("Distributed Logging & Monitoring System - %s\n", os.Args[1])
}