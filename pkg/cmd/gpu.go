package cmd

import (
	"fmt"
	"sort"
	"time"
)

// Metric names for GPU metrics
const (
	MetricGPUTemp       = "DCGM_FI_DEV_GPU_TEMP"
	MetricGPUMemoryFree = "DCGM_FI_DEV_FB_FREE"
	MetricGPUMemoryUsed = "DCGM_FI_DEV_FB_USED"
	MetricGPUUtil       = "DCGM_FI_DEV_GPU_UTIL"
	MetricGPUMemoryUtil = "DCGM_FI_DEV_MEM_COPY_UTIL"
)

// GpuStatus represents the status of a GPU
type GpuStatus struct {
	Hostname  string    `json:"Hostname"`
	DeviceID  string    `json:"gpu"`
	UUID      string    `json:"uuid"`
	Timestamp time.Time `json:"timestamp"`
	Name      string    `json:"modelName"`
	MemFree   float64   `json:"memory_free"`
	MemUsed   float64   `json:"memory_used"`
	MemTotal  float64   `json:"memory_total"`
	GPUUtil   float64   `json:"gpu_utilization"`
	MemUtil   float64   `json:"gpu_memory_utilization"`
	GPUTemp   float64   `json:"gpu_temp"`
}

// ByHostnameAndDeviceID implements sort.Interface for []GpuStatus based on Hostname and DeviceID
type ByHostnameAndDeviceID []GpuStatus

func (a ByHostnameAndDeviceID) Len() int      { return len(a) }
func (a ByHostnameAndDeviceID) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByHostnameAndDeviceID) Less(i, j int) bool {
	if a[i].Hostname != a[j].Hostname {
		return a[i].Hostname < a[j].Hostname
	}
	return a[i].DeviceID < a[j].DeviceID
}

// ConvertUTCToJST converts UTC time to JST
func ConvertUTCToJST(utcTime time.Time) time.Time {
	jst := time.FixedZone("Asia/Tokyo", 9*60*60)
	return utcTime.In(jst)
}

// MergeGpuMetrics merges Prometheus metrics into GPU status
func MergeGpuMetrics(results []Result) ([]GpuStatus, error) {
	if len(results) == 0 {
		return nil, fmt.Errorf("no metrics provided")
	}

	gpuMap := make(map[string]*GpuStatus)

	for _, result := range results {
		uuid := result.Metric["UUID"]
		if uuid == "" {
			continue
		}

		status, exists := gpuMap[uuid]
		if !exists {
			status = &GpuStatus{
				Hostname: result.Metric["Hostname"],
				DeviceID: result.Metric["gpu"],
				Name:     result.Metric["modelName"],
				UUID:     uuid,
			}
			gpuMap[uuid] = status
		}

		timestamp, err := result.GetTimestamp()
		if err == nil {
			status.Timestamp = ConvertUTCToJST(timestamp)
		}

		metricName := result.Metric["__name__"]
		val, err := result.GetValue()
		if err != nil {
			return nil, fmt.Errorf("invalid value for metric %s: %v", metricName, err)
		}

		switch metricName {
		case MetricGPUMemoryFree:
			status.MemFree = val
		case MetricGPUMemoryUsed:
			status.MemUsed = val
		case MetricGPUUtil:
			status.GPUUtil = val
		case MetricGPUMemoryUtil:
			status.MemUtil = val
		case MetricGPUTemp:
			status.GPUTemp = val
		default:
			return nil, fmt.Errorf("invalid metric name: %s", metricName)
		}
		if status.MemFree != 0 && status.MemUsed != 0 {
			status.MemTotal = status.MemFree + status.MemUsed
		}
	}

	if len(gpuMap) == 0 {
		return nil, fmt.Errorf("no valid GPU metrics found")
	}

	statuses := make([]GpuStatus, 0, len(gpuMap))
	for _, s := range gpuMap {
		statuses = append(statuses, *s)
	}

	// Sort the statuses by Hostname and DeviceID
	sort.Sort(ByHostnameAndDeviceID(statuses))

	return statuses, nil
}
