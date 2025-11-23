package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"test-task/internal/models"
	"time"
)

func (h *Handler) JSONHandler(w http.ResponseWriter, r *http.Request) {
	go h.statService.RegisterRequest()

	stats := h.statService.GetStats()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (h *Handler) HTMLHandler(w http.ResponseWriter, r *http.Request) {
	go h.statService.RegisterRequest()

	stats := h.statService.GetStats()

	statsWithUptime := struct {
		models.StatsResponse
		FormattedUptime string
	}{
		StatsResponse:   stats,
		FormattedUptime: formatUptime(time.Since(h.statService.GetStartTime())),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	h.tmpl.Execute(w, statsWithUptime)
}

func formatUptime(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if hours > 0 {
		return formatDurationPart(hours, "h") + formatDurationPart(minutes, "m") + formatDurationPart(seconds, "s")
	}
	if minutes > 0 {
		return formatDurationPart(minutes, "m") + formatDurationPart(seconds, "s")
	}
	return formatDurationPart(seconds, "s")
}

func formatDurationPart(value int, unit string) string {
	if value == 0 {
		return ""
	}
	return strconv.Itoa(value) + unit + " "
}
