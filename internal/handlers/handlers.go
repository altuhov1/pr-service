package handlers

import (
	"encoding/json"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"test-task/internal/services"
)

type Handler struct {
	TeamManag        services.TeamManager
	UserManag        services.UserManager
	PullRequestManag services.PullRequestManager
	statService      *services.StatService
	tmpl             *template.Template
}

func NewHandler(
	TeamManag services.TeamManager,
	UserManag services.UserManager,
	PullRequestManag services.PullRequestManager,
	statService *services.StatService,
) (*Handler, error) {
	tmpl := template.New("stats.html").Funcs(template.FuncMap{
		"formatBytes": formatBytes,
	})

	tmpl, err := tmpl.ParseFiles("static/stats.html")
	if err != nil {
		slog.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	return &Handler{
		TeamManag:        TeamManag,
		UserManag:        UserManag,
		PullRequestManag: PullRequestManag,
		statService:      statService,
		tmpl:             tmpl,
	}, nil
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": message,
	})
}

func writeErrorResponse(w http.ResponseWriter, status int, code, message string) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]interface{}{
			"code":    code,
			"message": message,
		},
	})
}
