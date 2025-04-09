package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/V01d42/dcgm-metrics-api/pkg/cmd"
)

func TestFetchPrometheusMetrics(t *testing.T) {
	tests := []struct {
		name          string
		metricNames   string
		mockResponse  string
		expectedError bool
	}{
		{
			name:        "Success: Single metric",
			metricNames: "- DCGM_FI_DEV_GPU_TEMP",
			mockResponse: `{
				"status": "success",
				"data": {
					"resultType": "vector",
					"result": [
						{
							"metric": {
								"__name__": "DCGM_FI_DEV_GPU_TEMP",
								"Hostname": "test-host",
								"gpu": "0",
								"UUID": "test-uuid",
								"modelName": "Test GPU"
							},
							"value": [1743982065.253, "14"]
						}
					]
				}
			}`,
			expectedError: false,
		},
		{
			name:        "Success: Multiple metrics",
			metricNames: "- DCGM_FI_DEV_GPU_TEMP\n- DCGM_FI_DEV_POWER_USAGE",
			mockResponse: `{
				"status": "success",
				"data": {
					"resultType": "vector",
					"result": [
						{
							"metric": {
								"__name__": "DCGM_FI_DEV_GPU_TEMP",
								"Hostname": "test-host",
								"gpu": "0",
								"UUID": "test-uuid",
								"modelName": "Test GPU"
							},
							"value": [1743982065.253, "14"]
						},
						{
							"metric": {
								"__name__": "DCGM_FI_DEV_POWER_USAGE",
								"Hostname": "test-host",
								"gpu": "0",
								"UUID": "test-uuid",
								"modelName": "Test GPU"
							},
							"value": [1743982065.253, "100"]
						}
					]
				}
			}`,
			expectedError: false,
		},
		{
			name:          "Error: Invalid YAML",
			metricNames:   "invalid yaml",
			mockResponse:  `{}`,
			expectedError: true,
		},
		{
			name:          "Error: Invalid response",
			metricNames:   "- DCGM_FI_DEV_GPU_TEMP",
			mockResponse:  `invalid json`,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(tt.mockResponse))
			}))
			defer server.Close()

			// Set environment variables
			os.Setenv("PROMETHEUS_URL", server.URL)
			os.Setenv("METRIC_NAMES", tt.metricNames)

			// Execute test
			results, err := cmd.FetchPrometheusMetrics(server.URL, tt.metricNames)

			// Check error
			if tt.expectedError {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Validate results
			if len(results) == 0 {
				t.Error("expected results but got empty")
			}

			// Validate timestamp
			for _, result := range results {
				timestamp, err := result.GetTimestamp()
				if err != nil {
					t.Errorf("failed to get timestamp: %v", err)
				}
				if timestamp.IsZero() {
					t.Error("timestamp is zero")
				}

				value, err := result.GetValue()
				if err != nil {
					t.Errorf("failed to get value: %v", err)
				}
				if value == 0 {
					t.Error("value is zero")
				}
			}
		})
	}
}

func TestMergeGpuMetrics(t *testing.T) {
	// Define test cases
	tests := []struct {
		name          string
		results       []cmd.Result
		expectedCount int
		expectedError bool
	}{
		{
			name: "Success: Single GPU",
			results: []cmd.Result{
				{
					Metric: map[string]string{
						"__name__":  "DCGM_FI_DEV_GPU_TEMP",
						"Hostname":  "test-host",
						"gpu":       "0",
						"UUID":      "test-uuid",
						"modelName": "Test GPU",
					},
					Value: []interface{}{1743982065.253, "14"},
				},
			},
			expectedCount: 1,
			expectedError: false,
		},
		{
			name: "Success: Multiple GPUs",
			results: []cmd.Result{
				{
					Metric: map[string]string{
						"__name__":  "DCGM_FI_DEV_GPU_TEMP",
						"Hostname":  "test-host",
						"gpu":       "0",
						"UUID":      "test-uuid-1",
						"modelName": "Test GPU 1",
					},
					Value: []interface{}{1743982065.253, "14"},
				},
				{
					Metric: map[string]string{
						"__name__":  "DCGM_FI_DEV_GPU_TEMP",
						"Hostname":  "test-host",
						"gpu":       "1",
						"UUID":      "test-uuid-2",
						"modelName": "Test GPU 2",
					},
					Value: []interface{}{1743982065.253, "15"},
				},
			},
			expectedCount: 2,
			expectedError: false,
		},
		{
			name: "Error: Invalid metric",
			results: []cmd.Result{
				{
					Metric: map[string]string{
						"__name__":  "INVALID_METRIC",
						"Hostname":  "test-host",
						"gpu":       "0",
						"UUID":      "test-uuid",
						"modelName": "Test GPU",
					},
					Value: []interface{}{1743982065.253, "invalid"},
				},
			},
			expectedCount: 0,
			expectedError: true,
		},
		{
			name: "Success: Multiple GPUs with different hostnames",
			results: []cmd.Result{
				{
					Metric: map[string]string{
						"__name__":  "DCGM_FI_DEV_GPU_TEMP",
						"Hostname":  "gpu14",
						"gpu":       "9",
						"UUID":      "test-uuid-2",
						"modelName": "Test GPU 2",
					},
					Value: []interface{}{1743982065.253, "15"},
				},
				{
					Metric: map[string]string{
						"__name__":  "DCGM_FI_DEV_GPU_TEMP",
						"Hostname":  "gpu14",
						"gpu":       "0",
						"UUID":      "test-uuid-1",
						"modelName": "Test GPU 1",
					},
					Value: []interface{}{1743982065.253, "16"},
				},
				{
					Metric: map[string]string{
						"__name__":  "DCGM_FI_DEV_GPU_TEMP",
						"Hostname":  "gpu14",
						"gpu":       "3",
						"UUID":      "test-uuid-3",
						"modelName": "Test GPU 3",
					},
					Value: []interface{}{1743982065.253, "14"},
				},
				{
					Metric: map[string]string{
						"__name__":  "DCGM_FI_DEV_GPU_TEMP",
						"Hostname":  "gpu15",
						"gpu":       "5",
						"UUID":      "test-uuid-4",
						"modelName": "Test GPU 4",
					},
					Value: []interface{}{1743982065.253, "17"},
				},
			},
			expectedCount: 4,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute test
			statuses, err := cmd.MergeGpuMetrics(tt.results)

			// Check error
			if tt.expectedError {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Validate results
			if len(statuses) != tt.expectedCount {
				t.Errorf("expected %d GPUs, got %d", tt.expectedCount, len(statuses))
			}

			// Validate GPU status
			for _, status := range statuses {
				if status.Hostname == "" {
					t.Error("Hostname is empty")
				}
				if status.DeviceID == "" {
					t.Error("DeviceID is empty")
				}
				if status.UUID == "" {
					t.Error("UUID is empty")
				}
				if status.Name == "" {
					t.Error("Name is empty")
				}
				if status.Timestamp.IsZero() {
					t.Error("Timestamp is zero")
				}
			}

			// For the sorting test case, verify the order
			if tt.name == "Success: Multiple GPUs with different hostnames" {
				// Expected order:
				// 1. gpu14, gpu0
				// 2. gpu14, gpu3
				// 3. gpu14, gpu9
				// 4. gpu15, gpu5
				if statuses[0].Hostname != "gpu14" || statuses[0].DeviceID != "0" {
					t.Errorf("first GPU should be gpu14:gpu0, got %s:%s", statuses[0].Hostname, statuses[0].DeviceID)
				}
				if statuses[1].Hostname != "gpu14" || statuses[1].DeviceID != "3" {
					t.Errorf("second GPU should be gpu14:gpu3, got %s:%s", statuses[1].Hostname, statuses[1].DeviceID)
				}
				if statuses[2].Hostname != "gpu14" || statuses[2].DeviceID != "9" {
					t.Errorf("third GPU should be gpu14:gpu9, got %s:%s", statuses[2].Hostname, statuses[2].DeviceID)
				}
				if statuses[3].Hostname != "gpu15" || statuses[3].DeviceID != "5" {
					t.Errorf("fourth GPU should be gpu15:gpu5, got %s:%s", statuses[3].Hostname, statuses[3].DeviceID)
				}
			}
		})
	}
}

func TestMetricsHandler(t *testing.T) {
	// Test cases
	tests := []struct {
		name           string
		endpoint       string
		metricNames    string
		mockResponse   string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Success with default endpoint",
			endpoint:       "/metrics",
			metricNames:    "- DCGM_FI_DEV_GPU_TEMP",
			mockResponse:   `{"status":"success","data":{"resultType":"vector","result":[{"metric":{"__name__":"DCGM_FI_DEV_GPU_TEMP","Hostname":"test-host","gpu":"0","UUID":"test-uuid","modelName":"Test GPU"},"value":[1743982065.253,"14"]}]}}`,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Success with custom endpoint",
			endpoint:       "/custom-metrics",
			metricNames:    "- DCGM_FI_DEV_GPU_TEMP",
			mockResponse:   `{"status":"success","data":{"resultType":"vector","result":[{"metric":{"__name__":"DCGM_FI_DEV_GPU_TEMP","Hostname":"test-host","gpu":"0","UUID":"test-uuid","modelName":"Test GPU"},"value":[1743982065.253,"14"]}]}}`,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Missing PROMETHEUS_URL",
			endpoint:       "/metrics",
			metricNames:    "- DCGM_FI_DEV_GPU_TEMP",
			mockResponse:   "",
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "PROMETHEUS_URL environment variable is not set",
		},
		{
			name:           "Missing METRIC_NAMES",
			endpoint:       "/metrics",
			metricNames:    "",
			mockResponse:   "",
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "METRIC_NAMES environment variable is not set",
		},
		{
			name:           "Invalid Prometheus response",
			endpoint:       "/metrics",
			metricNames:    "- DCGM_FI_DEV_GPU_TEMP",
			mockResponse:   "invalid json",
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "failed to decode response for metric DCGM_FI_DEV_GPU_TEMP: invalid character 'i' looking for beginning of value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(tt.mockResponse))
			}))
			defer server.Close()

			// Set environment variables
			if tt.name != "Missing PROMETHEUS_URL" {
				os.Setenv("PROMETHEUS_URL", server.URL)
			} else {
				os.Unsetenv("PROMETHEUS_URL")
			}

			if tt.name != "Missing METRIC_NAMES" {
				os.Setenv("METRIC_NAMES", tt.metricNames)
			} else {
				os.Unsetenv("METRIC_NAMES")
			}

			os.Setenv("METRICS_ENDPOINT", tt.endpoint)

			// Create request
			req := httptest.NewRequest("GET", tt.endpoint, nil)
			w := httptest.NewRecorder()

			// Call handler
			cmd.MetricsHandler(w, req)

			// Check response
			resp := w.Result()
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if tt.expectedError != "" {
				var errorResponse struct {
					Error string `json:"error"`
				}
				if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err != nil {
					t.Errorf("failed to decode error response: %v", err)
				}
				if errorResponse.Error != tt.expectedError {
					t.Errorf("expected error %q, got %q", tt.expectedError, errorResponse.Error)
				}
			} else {
				var statuses []cmd.GpuStatus
				if err := json.NewDecoder(resp.Body).Decode(&statuses); err != nil {
					t.Errorf("failed to decode response: %v", err)
				}
				if len(statuses) == 0 {
					t.Error("expected GPU statuses but got empty")
				}
			}
		})
	}
}

func TestRun(t *testing.T) {
	// Test cases
	tests := []struct {
		name          string
		endpoint      string
		expectedError bool
	}{
		{
			name:          "Default endpoint",
			endpoint:      "",
			expectedError: false,
		},
		{
			name:          "Custom endpoint",
			endpoint:      "/custom-metrics",
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			if tt.endpoint != "" {
				os.Setenv("METRICS_ENDPOINT", tt.endpoint)
			} else {
				os.Unsetenv("METRICS_ENDPOINT")
			}

			// Start server in a goroutine
			go func() {
				if err := cmd.Run(); err != nil {
					t.Errorf("Run() error = %v", err)
				}
			}()

			// Cleanup
			defer func() {
				os.Unsetenv("METRICS_ENDPOINT")
			}()
		})
	}
}
