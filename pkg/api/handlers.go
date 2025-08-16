package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"awesomeProject6/pkg/elasticsearch"
	"awesomeProject6/pkg/prometheus"
)

type Handlers struct {
	esClient   *elasticsearch.Client
	aggregator *prometheus.Aggregator
	logger     *logrus.Logger
}

func NewHandlers(esClient *elasticsearch.Client, aggregator *prometheus.Aggregator) *Handlers {
	return &Handlers{
		esClient:   esClient,
		aggregator: aggregator,
		logger:     logrus.New(),
	}
}

func (h *Handlers) SetupRoutes(router *mux.Router) {
	api := router.PathPrefix("/api/v1").Subrouter()
	
	api.HandleFunc("/logs/search", h.searchLogs).Methods("GET")
	api.HandleFunc("/metrics/query", h.queryMetrics).Methods("GET")
	api.HandleFunc("/metrics/range", h.queryMetricsRange).Methods("GET")
	api.HandleFunc("/health", h.healthCheck).Methods("GET")
}

func (h *Handlers) searchLogs(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	service := query.Get("service")
	level := query.Get("level")
	from := query.Get("from")
	to := query.Get("to")
	limit := query.Get("limit")

	response := map[string]interface{}{
		"logs": []interface{}{},
		"total": 0,
		"query": map[string]string{
			"service": service,
			"level":   level,
			"from":    from,
			"to":      to,
			"limit":   limit,
		},
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

func (h *Handlers) queryMetrics(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	metric := query.Get("metric")
	function := query.Get("function")
	durationStr := query.Get("duration")

	if metric == "" || function == "" || durationStr == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "Missing required parameters")
		return
	}

	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid duration format")
		return
	}

	labels := make(map[string]string)
	for key, values := range query {
		if key != "metric" && key != "function" && key != "duration" && len(values) > 0 {
			labels[key] = values[0]
		}
	}

	var value float64
	switch function {
	case "sum":
		value = h.aggregator.Sum(metric, labels, duration)
	case "avg":
		value = h.aggregator.Average(metric, labels, duration)
	case "max":
		value = h.aggregator.Max(metric, labels, duration)
	case "rate":
		value = h.aggregator.Rate(metric, labels, duration)
	case "p95":
		value = h.aggregator.Percentile(metric, labels, duration, 95)
	case "p99":
		value = h.aggregator.Percentile(metric, labels, duration, 99)
	default:
		h.writeErrorResponse(w, http.StatusBadRequest, "Unknown function")
		return
	}

	response := map[string]interface{}{
		"metric":   metric,
		"function": function,
		"value":    value,
		"duration": durationStr,
		"labels":   labels,
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

func (h *Handlers) queryMetricsRange(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	metric := query.Get("metric")
	fromStr := query.Get("from")
	toStr := query.Get("to")
	stepStr := query.Get("step")

	if metric == "" || fromStr == "" || toStr == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "Missing required parameters")
		return
	}

	from, err := strconv.ParseInt(fromStr, 10, 64)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid from timestamp")
		return
	}

	to, err := strconv.ParseInt(toStr, 10, 64)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid to timestamp")
		return
	}

	step := int64(60)
	if stepStr != "" {
		step, err = strconv.ParseInt(stepStr, 10, 64)
		if err != nil {
			h.writeErrorResponse(w, http.StatusBadRequest, "Invalid step")
			return
		}
	}

	response := map[string]interface{}{
		"metric": metric,
		"values": []map[string]interface{}{},
		"from":   from,
		"to":     to,
		"step":   step,
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

func (h *Handlers) healthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"version":   "1.0.0",
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

func (h *Handlers) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Errorf("Failed to encode JSON response: %v", err)
	}
}

func (h *Handlers) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	response := map[string]interface{}{
		"error":   message,
		"status":  statusCode,
		"timestamp": time.Now().Unix(),
	}

	h.writeJSONResponse(w, statusCode, response)
}