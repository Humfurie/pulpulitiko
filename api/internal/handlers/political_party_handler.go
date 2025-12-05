package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"pulpulitiko/internal/models"
	"pulpulitiko/internal/services"
)

type PoliticalPartyHandler struct {
	partyService *services.PoliticalPartyService
}

func NewPoliticalPartyHandler(partyService *services.PoliticalPartyService) *PoliticalPartyHandler {
	return &PoliticalPartyHandler{partyService: partyService}
}

// Public endpoints

// GetParties returns a list of political parties
func (h *PoliticalPartyHandler) GetParties(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	majorOnly := r.URL.Query().Get("major_only") == "true"
	activeOnly := r.URL.Query().Get("active_only") != "false" // Default to true

	parties, err := h.partyService.List(r.Context(), page, perPage, majorOnly, activeOnly)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get parties", err)
		return
	}

	respondWithJSON(w, http.StatusOK, parties)
}

// GetAllParties returns all political parties (for dropdowns)
func (h *PoliticalPartyHandler) GetAllParties(w http.ResponseWriter, r *http.Request) {
	activeOnly := r.URL.Query().Get("active_only") != "false"

	parties, err := h.partyService.GetAll(r.Context(), activeOnly)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get parties", err)
		return
	}

	respondWithJSON(w, http.StatusOK, parties)
}

// GetPartyBySlug returns a party by slug
func (h *PoliticalPartyHandler) GetPartyBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	party, err := h.partyService.GetBySlug(r.Context(), slug)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get party", err)
		return
	}
	if party == nil {
		respondWithError(w, http.StatusNotFound, "Party not found", nil)
		return
	}

	respondWithJSON(w, http.StatusOK, party)
}

// GetAllPositions returns all government positions
func (h *PoliticalPartyHandler) GetAllPositions(w http.ResponseWriter, r *http.Request) {
	positions, err := h.partyService.GetAllPositions(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get positions", err)
		return
	}

	respondWithJSON(w, http.StatusOK, positions)
}

// GetPositionsByLevel returns positions for a specific government level
func (h *PoliticalPartyHandler) GetPositionsByLevel(w http.ResponseWriter, r *http.Request) {
	level := chi.URLParam(r, "level")

	// Validate level
	validLevels := []string{"national", "regional", "provincial", "city", "municipal", "barangay"}
	isValid := false
	for _, v := range validLevels {
		if level == v {
			isValid = true
			break
		}
	}
	if !isValid {
		respondWithError(w, http.StatusBadRequest, "Invalid government level", nil)
		return
	}

	positions, err := h.partyService.GetPositionsByLevel(r.Context(), level)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get positions", err)
		return
	}

	respondWithJSON(w, http.StatusOK, positions)
}

// GetPositionBySlug returns a position by slug
func (h *PoliticalPartyHandler) GetPositionBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	position, err := h.partyService.GetPositionBySlug(r.Context(), slug)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get position", err)
		return
	}
	if position == nil {
		respondWithError(w, http.StatusNotFound, "Position not found", nil)
		return
	}

	respondWithJSON(w, http.StatusOK, position)
}

// FindMyRepresentatives returns representatives for a given barangay
func (h *PoliticalPartyHandler) FindMyRepresentatives(w http.ResponseWriter, r *http.Request) {
	barangayIDStr := r.URL.Query().Get("barangay_id")
	if barangayIDStr == "" {
		respondWithError(w, http.StatusBadRequest, "barangay_id is required", nil)
		return
	}

	barangayID, err := uuid.Parse(barangayIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid barangay_id", err)
		return
	}

	representatives, err := h.partyService.FindRepresentativesByBarangay(r.Context(), barangayID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to find representatives", err)
		return
	}

	respondWithJSON(w, http.StatusOK, representatives)
}

// Admin endpoints

// CreateParty creates a new political party
func (h *PoliticalPartyHandler) CreateParty(w http.ResponseWriter, r *http.Request) {
	var req models.CreatePoliticalPartyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	party, err := h.partyService.Create(r.Context(), &req)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create party", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, party)
}

// UpdateParty updates a political party
func (h *PoliticalPartyHandler) UpdateParty(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid party ID", err)
		return
	}

	var req models.UpdatePoliticalPartyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	party, err := h.partyService.Update(r.Context(), id, &req)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update party", err)
		return
	}
	if party == nil {
		respondWithError(w, http.StatusNotFound, "Party not found", nil)
		return
	}

	respondWithJSON(w, http.StatusOK, party)
}

// DeleteParty deletes a political party
func (h *PoliticalPartyHandler) DeleteParty(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid party ID", err)
		return
	}

	err = h.partyService.Delete(r.Context(), id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to delete party", err)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Party deleted successfully"})
}

// Jurisdiction endpoints

// CreateJurisdiction creates a jurisdiction for a politician
func (h *PoliticalPartyHandler) CreateJurisdiction(w http.ResponseWriter, r *http.Request) {
	var req models.CreatePoliticianJurisdictionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	jurisdiction, err := h.partyService.CreateJurisdiction(r.Context(), &req)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create jurisdiction", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, jurisdiction)
}

// GetJurisdictionsByPolitician returns jurisdictions for a politician
func (h *PoliticalPartyHandler) GetJurisdictionsByPolitician(w http.ResponseWriter, r *http.Request) {
	politicianIDStr := chi.URLParam(r, "politicianId")
	politicianID, err := uuid.Parse(politicianIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid politician ID", err)
		return
	}

	jurisdictions, err := h.partyService.GetJurisdictionsByPolitician(r.Context(), politicianID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get jurisdictions", err)
		return
	}

	respondWithJSON(w, http.StatusOK, jurisdictions)
}

// DeleteJurisdiction deletes a jurisdiction
func (h *PoliticalPartyHandler) DeleteJurisdiction(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid jurisdiction ID", err)
		return
	}

	err = h.partyService.DeleteJurisdiction(r.Context(), id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to delete jurisdiction", err)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Jurisdiction deleted successfully"})
}
