package handlers

/*
	// POST /pullRequest/create
	// POST /pullRequest/merge
	// POST /pullRequest/reassign

*/
import (
	"encoding/json"
	"net/http"
	"subscription-budget/internal/models"
)

// POST /pullRequest/create
func (h *Handler) CreatePR(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.CreatePRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.PullRequestID == "" || req.PullRequestName == "" || req.AuthorID == "" {
		writeError(w, http.StatusBadRequest, "pull_request_id, pull_request_name and author_id are required")
		return
	}

	pr, err := h.PullRequestManag.CreatePR(r.Context(), req)
	if err != nil {
		switch err {
		case models.ErrPRExists:
			writeErrorResponse(w, http.StatusConflict, "PR_EXISTS", "PR id already exists")
		case models.ErrNotFound:
			writeErrorResponse(w, http.StatusNotFound, "NOT_FOUND", "resource not found")
		default:
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	response := map[string]interface{}{
		"pr": pr,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// POST /pullRequest/merge
func (h *Handler) MergePR(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		PullRequestID string `json:"pull_request_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.PullRequestID == "" {
		writeError(w, http.StatusBadRequest, "pull_request_id is required")
		return
	}

	pr, err := h.PullRequestManag.MergePR(r.Context(), req.PullRequestID)
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
		"pr": pr,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// POST /pullRequest/reassign
func (h *Handler) ReassignReviewer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.ReassignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.PullRequestID == "" || req.OldUserID == "" {
		writeError(w, http.StatusBadRequest, "pull_request_id and old_user_id are required")
		return
	}

	pr, newReviewer, err := h.PullRequestManag.ReassignReviewer(r.Context(), req)
	if err != nil {
		switch err {
		case models.ErrNotFound:
			writeErrorResponse(w, http.StatusNotFound, "NOT_FOUND", "resource not found")
		case models.ErrPRMerged:
			writeErrorResponse(w, http.StatusConflict, "PR_MERGED", "cannot reassign on merged PR")
		case models.ErrNotAssigned:
			writeErrorResponse(w, http.StatusConflict, "NOT_ASSIGNED", "reviewer is not assigned to this PR")
		case models.ErrNoCandidate:
			writeErrorResponse(w, http.StatusConflict, "NO_CANDIDATE", "no active replacement candidate in team")
		default:
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	response := map[string]interface{}{
		"pr":          pr,
		"replaced_by": newReviewer,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
