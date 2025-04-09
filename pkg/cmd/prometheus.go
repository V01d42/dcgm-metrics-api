package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

// PrometheusResponse represents the response from Prometheus API
type PrometheusResponse struct {
	Status string         `json:"status"`
	Data   PrometheusData `json:"data"`
}

// PrometheusData represents the data part of Prometheus response
type PrometheusData struct {
	ResultType string   `json:"resultType"`
	Result     []Result `json:"result"`
}

// Result represents a single metric result
type Result struct {
	Metric map[string]string `json:"metric"`
	Value  []interface{}     `json:"value"`
}

// GetTimestamp returns the timestamp as time.Time
func (r *Result) GetTimestamp() (time.Time, error) {
	if len(r.Value) < 1 {
		return time.Time{}, fmt.Errorf("no timestamp in value")
	}

	timestamp, ok := r.Value[0].(float64)
	if !ok {
		return time.Time{}, fmt.Errorf("invalid timestamp format: expected float64")
	}

	// Convert Unix timestamp (seconds) to time.Time
	sec := int64(timestamp)
	nsec := int64((timestamp - float64(sec)) * 1e9)
	return time.Unix(sec, nsec), nil
}

// GetValue returns the metric value as float64
func (r *Result) GetValue() (float64, error) {
	if len(r.Value) < 2 {
		return 0, fmt.Errorf("no value in result")
	}

	valStr, ok := r.Value[1].(string)
	if !ok {
		return 0, fmt.Errorf("invalid value format: expected string")
	}

	val, err := strconv.ParseFloat(valStr, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse value: %v", err)
	}

	return val, nil
}

// FetchPrometheusMetrics fetches metrics from Prometheus API
func FetchPrometheusMetrics(promURL string, metricNamesStr string) ([]Result, error) {
	if promURL == "" {
		return nil, fmt.Errorf("prometheus URL is empty")
	}

	if metricNamesStr == "" {
		return nil, fmt.Errorf("metric names string is empty")
	}

	var metricNames []string
	if err := yaml.Unmarshal([]byte(metricNamesStr), &metricNames); err != nil {
		return nil, fmt.Errorf("failed to parse metric names: %v", err)
	}

	if len(metricNames) == 0 {
		return nil, fmt.Errorf("no metric names provided")
	}

	var allResults []Result

	for _, metric := range metricNames {
		url := fmt.Sprintf("%s/api/v1/query?query=%s", promURL, metric)

		resp, err := http.Get(url)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch metric %s: %v", metric, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("failed to fetch metric %s: status %d, body: %s",
				metric, resp.StatusCode, string(body))
		}

		var pResp PrometheusResponse
		if err := json.NewDecoder(resp.Body).Decode(&pResp); err != nil {
			return nil, fmt.Errorf("failed to decode response for metric %s: %v", metric, err)
		}

		if pResp.Status != "success" {
			return nil, fmt.Errorf("prometheus returned non-success status for metric %s: %s",
				metric, pResp.Status)
		}

		allResults = append(allResults, pResp.Data.Result...)
	}

	if len(allResults) == 0 {
		return nil, fmt.Errorf("no results returned from Prometheus")
	}

	return allResults, nil
}
