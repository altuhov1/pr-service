package handlers

/*
	// POST /team/add
	// GET /team/get
*/
import (
	"encoding/json"
	"net/http"
	"test-task/internal/models"
)

// POST /team/add
func (h *Handler) AddTeam(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		TeamName string        `json:"team_name"`
		Members  []models.User `json:"members"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	for i := range request.Members {
		request.Members[i].TeamName = request.TeamName
	}

	team := models.Team{
		TeamName: request.TeamName,
		Members:  request.Members,
	}

	createdTeam, err := h.TeamManag.CreateTeam(r.Context(), team)
	if err != nil {
		switch err {
		case models.ErrTeamExists:
			writeErrorResponse(w, http.StatusBadRequest, "TEAM_EXISTS", "team_name already exists")
		default:
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	response := map[string]interface{}{
		"team": createdTeam,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GET /team/get
func (h *Handler) GetTeam(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		writeError(w, http.StatusBadRequest, "team_name parameter is required")
		return
	}

	team, err := h.TeamManag.GetTeam(r.Context(), teamName)
	if err != nil {
		switch err {
		case models.ErrNotFound:
			writeErrorResponse(w, http.StatusNotFound, "NOT_FOUND", "resource not found")
		default:
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(team)
}
