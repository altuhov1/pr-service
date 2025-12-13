package services

import (
	"runtime"
	"sync"
	"subscription-budget/internal/models"
	"time"
)

type StatService struct {
	mu            sync.RWMutex
	startTime     time.Time
	totalRequests int64
}

func NewStatService() *StatService {
	return &StatService{
		startTime: time.Now(),
	}
}

func (s *StatService) RegisterRequest() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.totalRequests++
}

func (s *StatService) GetStats() models.StatsResponse {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	s.mu.RLock()
	defer s.mu.RUnlock()

	return models.StatsResponse{
		System: models.SystemStats{
			Goroutines: runtime.NumGoroutine(),
			CPUCores:   runtime.NumCPU(),
			GoVersion:  runtime.Version(),
			OS:         runtime.GOOS,
			Arch:       runtime.GOARCH,
		},
		Memory: models.MemoryStats{
			Alloc:      memStats.Alloc,
			TotalAlloc: memStats.TotalAlloc,
			Sys:        memStats.Sys,
			NumGC:      memStats.NumGC,
			HeapAlloc:  memStats.HeapAlloc,
			HeapSys:    memStats.HeapSys,
		},
		Server: models.ServerStats{
			UptimeSeconds: time.Since(s.startTime).Seconds(),
			StartTime:     s.startTime.Format(time.RFC3339),
			TotalRequests: s.totalRequests,
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}
}

func (s *StatService) GetStartTime() time.Time {
	return s.startTime
}
