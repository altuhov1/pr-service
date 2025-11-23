package handlers
/*
	// POST /users/setIsActive
	// GET /users/getReview
*/
import (
	"encoding/json"
	"net/http"
	"test-task/internal/models"
)

// POST /users/setIsActive
func (h *Handler) SetIsActive(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		UserID   string `json:"user_id"`
		IsActive bool   `json:"is_active"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.UserID == "" {
		writeError(w, http.StatusBadRequest, "user_id is required")
		return
	}

	user, err := h.UserManag.SetUserActive(r.Context(), req.UserID, req.IsActive)
	if err != nil {
		switch err {
		case models.ErrNotFound:
			writeErrorResponse(w, http.StatusNotFound, "NOT_FOUND", "resource not found")
		default:
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	response := map[string]interface{}{
		"user": map[string]interface{}{
			"user_id":   user.UserID,
			"username":  user.Username,
			"team_name": user.TeamName,
			"is_active": user.IsActive,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GET /users/getReview
func (h *Handler) GetUserReviews(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "user_id parameter is required")
		return
	}

	prs, err := h.PullRequestManag.GetUserReviews(r.Context(), userID)
	if err != nil {
		switch err {
		case models.ErrNotFound:
			writeErrorResponse(w, http.StatusNotFound, "NOT_FOUND", "resource not found")
		default:
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	response := map[string]interface{}{
		"user_id":       userID,
		"pull_requests": prs,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
