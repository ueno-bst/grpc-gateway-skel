package runtime

import (
	"encoding/json"
	"net/http"
	"runtime"
	"time"
)

var startTime time.Time

// HealthCheckResponse は死活状況のレスポンス構造体
type HealthCheckResponse struct {
	Status string  `json:"status"`
	Time   string  `json:"time"`
	Uptime float64 `json:"uptime"`
}

type StatusResponse struct {
	Alloc        uint64 `json:"alloc"`
	TotalAlloc   uint64 `json:"totalAlloc"`
	Sys          uint64 `json:"sys"`
	NumGC        uint32 `json:"numGC"`
	NumGoroutine int    `json:"numProcess"`
	RequestCount int    `json:"requestCount"`
}

func WithHealthCheckPathHandle(endpoint string) GatewayOptionFunc {
	return func(opt *GatewayOption) {
		startTime = time.Now()
		opt.paths = append(opt.paths, PathHandler{"GET", endpoint, HealthCheckPathHandle})
	}
}

func WithStatusPathHandle(endpoint string) GatewayOptionFunc {
	return func(opt *GatewayOption) {
		opt.paths = append(opt.paths, PathHandler{"GET", endpoint, StatusCheckPathHandle})
	}
}

func HealthCheckPathHandle(w http.ResponseWriter, _ *http.Request, _ map[string]string) {
	uptime := time.Since(startTime)

	response := HealthCheckResponse{
		Status: "ok",
		Time:   time.Now().Format("2006-01-02T15:04:05.999Z07:00"),
		Uptime: uptime.Seconds(),
	}

	w.Header().Set("Content-Type", "application/json")

	err := json.NewEncoder(w).Encode(response)

	if err != nil {
		return
	}
}

func StatusCheckPathHandle(w http.ResponseWriter, _ *http.Request, _ map[string]string) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	response := StatusResponse{
		Alloc:        memStats.Alloc,
		TotalAlloc:   memStats.TotalAlloc,
		Sys:          memStats.Sys,
		NumGC:        memStats.NumGC,
		NumGoroutine: runtime.NumGoroutine(),
		RequestCount: 0,
	}

	err := json.NewEncoder(w).Encode(response)

	if err != nil {
		return
	}
}
