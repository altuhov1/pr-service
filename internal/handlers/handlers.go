package handlers

import (
	"encoding/json"
	"net/http"
	"test-task/internal/services"
)

type Handler struct {
	TeamManag        services.TeamManager
	UserManag        services.UserManager
	PullRequestManag services.PullRequestManager
}

func NewHandler(
	TeamManag services.TeamManager,
	UserManag services.UserManager,
	PullRequestManag services.PullRequestManager,
) (*Handler, error) {

	return &Handler{
		TeamManag:        TeamManag,
		UserManag:        UserManag,
		PullRequestManag: PullRequestManag,
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
