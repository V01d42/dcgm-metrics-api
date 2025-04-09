package cmd

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"gopkg.in/yaml.v3"
)

const (
	defaultEndpoint = "/metrics"
)

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// MetricsHandler handles HTTP requests for metrics
// It fetches metrics from Prometheus and returns them in a formatted JSON response
func MetricsHandler(w http.ResponseWriter, r *http.Request) {
	promURL := os.Getenv("PROMETHEUS_URL")
	if promURL == "" {
		sendError(w, "PROMETHEUS_URL environment variable is not set", http.StatusInternalServerError)
		return
	}

	metricNamesStr := os.Getenv("METRIC_NAMES")
	if metricNamesStr == "" {
		sendError(w, "METRIC_NAMES environment variable is not set", http.StatusInternalServerError)
		return
	}

	results, err := FetchPrometheusMetrics(promURL, metricNamesStr)
	if err != nil {
		sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := MergeGpuMetrics(results)
	if err != nil {
		sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		sendError(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// sendError sends an error response in JSON format
func sendError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(ErrorResponse{Error: message}); err != nil {
		// If JSON encoding fails, write a simple error message
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to encode error response"}`))
	}
}

// ReadinessProbeHandler handles readiness probe requests
func ReadinessProbeHandler(w http.ResponseWriter, r *http.Request) {
	// Check if Prometheus is accessible
	promURL := os.Getenv("PROMETHEUS_URL")
	if promURL == "" {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("PROMETHEUS_URL not set"))
		return
	}

	// Try to connect to Prometheus
	resp, err := http.Get(promURL + "/api/v1/query?query=up")
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("Cannot connect to Prometheus"))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("Prometheus returned non-200 status"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// LivenessProbeHandler handles liveness probe requests
func LivenessProbeHandler(w http.ResponseWriter, r *http.Request) {
	// Check if required environment variables are set
	promURL := os.Getenv("PROMETHEUS_URL")
	metricNamesStr := os.Getenv("METRIC_NAMES")

	if promURL == "" || metricNamesStr == "" {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("Required environment variables not set"))
		return
	}

	// Check if we can parse the metric names
	var metricNames []string
	if err := yaml.Unmarshal([]byte(metricNamesStr), &metricNames); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("Invalid METRIC_NAMES format"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// Run starts the HTTP server and sets up the metrics endpoint
func Run() error {
	// Get endpoint from environment variable or use default
	endpoint := os.Getenv("METRICS_ENDPOINT")
	if endpoint == "" {
		endpoint = defaultEndpoint
	}

	// Register handlers
	http.HandleFunc(endpoint, MetricsHandler)
	http.HandleFunc("/ready", ReadinessProbeHandler)
	http.HandleFunc("/health", LivenessProbeHandler)

	// Start server
	port := ":8080"
	log.Printf("Starting server on %s with endpoint %s", port, endpoint)
	return http.ListenAndServe(port, nil)
}
