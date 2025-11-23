package models

type SystemStats struct {
	Goroutines int    `json:"goroutines"`
	CPUCores   int    `json:"cpu_cores"`
	GoVersion  string `json:"go_version"`
	OS         string `json:"os"`
	Arch       string `json:"arch"`
}

type MemoryStats struct {
	Alloc      uint64 `json:"alloc"`
	TotalAlloc uint64 `json:"total_alloc"`
	Sys        uint64 `json:"sys"`
	NumGC      uint32 `json:"num_gc"`
	HeapAlloc  uint64 `json:"heap_alloc"`
	HeapSys    uint64 `json:"heap_sys"`
}

type ServerStats struct {
	UptimeSeconds float64 `json:"uptime_seconds"`
	StartTime     string  `json:"start_time"`
	TotalRequests int64   `json:"total_requests"`
}

type StatsResponse struct {
	System    SystemStats `json:"system"`
	Memory    MemoryStats `json:"memory"`
	Server    ServerStats `json:"server"`
	Timestamp string      `json:"timestamp"`
}
