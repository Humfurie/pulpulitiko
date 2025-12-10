package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/humfurie/pulpulitiko/api/internal/models"
	"github.com/humfurie/pulpulitiko/api/internal/services"
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
		WriteInternalError(w, "Failed to get parties")
		return
	}

	WriteSuccess(w, parties)
}

// GetAllParties returns all political parties (for dropdowns)
func (h *PoliticalPartyHandler) GetAllParties(w http.ResponseWriter, r *http.Request) {
	activeOnly := r.URL.Query().Get("active_only") != "false"

	parties, err := h.partyService.GetAll(r.Context(), activeOnly)
	if err != nil {
		WriteInternalError(w, "Failed to get parties")
		return
	}

	WriteSuccess(w, parties)
}

// GetPartyBySlug returns a party by slug
func (h *PoliticalPartyHandler) GetPartyBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	party, err := h.partyService.GetBySlug(r.Context(), slug)
	if err != nil {
		WriteInternalError(w, "Failed to get party")
		return
	}
	if party == nil {
		WriteNotFound(w, "Party not found")
		return
	}

	WriteSuccess(w, party)
}

// GetAllPositions returns all government positions
func (h *PoliticalPartyHandler) GetAllPositions(w http.ResponseWriter, r *http.Request) {
	positions, err := h.partyService.GetAllPositions(r.Context())
	if err != nil {
		WriteInternalError(w, "Failed to get positions")
		return
	}

	WriteSuccess(w, positions)
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
		WriteBadRequest(w, "Invalid government level")
		return
	}

	positions, err := h.partyService.GetPositionsByLevel(r.Context(), level)
	if err != nil {
		WriteInternalError(w, "Failed to get positions")
		return
	}

	WriteSuccess(w, positions)
}

// GetPositionBySlug returns a position by slug
func (h *PoliticalPartyHandler) GetPositionBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	position, err := h.partyService.GetPositionBySlug(r.Context(), slug)
	if err != nil {
		WriteInternalError(w, "Failed to get position")
		return
	}
	if position == nil {
		WriteNotFound(w, "Position not found")
		return
	}

	WriteSuccess(w, position)
}

// FindMyRepresentatives returns representatives for a given barangay
func (h *PoliticalPartyHandler) FindMyRepresentatives(w http.ResponseWriter, r *http.Request) {
	barangayIDStr := r.URL.Query().Get("barangay_id")
	if barangayIDStr == "" {
		WriteBadRequest(w, "barangay_id is required")
		return
	}

	barangayID, err := uuid.Parse(barangayIDStr)
	if err != nil {
		WriteBadRequest(w, "Invalid barangay_id")
		return
	}

	representatives, err := h.partyService.FindRepresentativesByBarangay(r.Context(), barangayID)
	if err != nil {
		WriteInternalError(w, "Failed to find representatives")
		return
	}

	WriteSuccess(w, representatives)
}

// Admin endpoints

// CreateParty creates a new political party
func (h *PoliticalPartyHandler) CreateParty(w http.ResponseWriter, r *http.Request) {
	var req models.CreatePoliticalPartyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteBadRequest(w, "Invalid request body")
		return
	}

	party, err := h.partyService.Create(r.Context(), &req)
	if err != nil {
		WriteInternalError(w, "Failed to create party")
		return
	}

	WriteCreated(w, party)
}

// UpdateParty updates a political party
func (h *PoliticalPartyHandler) UpdateParty(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "Invalid party ID")
		return
	}

	var req models.UpdatePoliticalPartyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteBadRequest(w, "Invalid request body")
		return
	}

	party, err := h.partyService.Update(r.Context(), id, &req)
	if err != nil {
		WriteInternalError(w, "Failed to update party")
		return
	}
	if party == nil {
		WriteNotFound(w, "Party not found")
		return
	}

	WriteSuccess(w, party)
}

// DeleteParty deletes a political party
func (h *PoliticalPartyHandler) DeleteParty(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "Invalid party ID")
		return
	}

	err = h.partyService.Delete(r.Context(), id)
	if err != nil {
		WriteInternalError(w, "Failed to delete party")
		return
	}

	WriteSuccess(w, map[string]string{"message": "Party deleted successfully"})
}

// Jurisdiction endpoints

// CreateJurisdiction creates a jurisdiction for a politician
func (h *PoliticalPartyHandler) CreateJurisdiction(w http.ResponseWriter, r *http.Request) {
	var req models.CreatePoliticianJurisdictionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteBadRequest(w, "Invalid request body")
		return
	}

	jurisdiction, err := h.partyService.CreateJurisdiction(r.Context(), &req)
	if err != nil {
		WriteInternalError(w, "Failed to create jurisdiction")
		return
	}

	WriteCreated(w, jurisdiction)
}

// GetJurisdictionsByPolitician returns jurisdictions for a politician
func (h *PoliticalPartyHandler) GetJurisdictionsByPolitician(w http.ResponseWriter, r *http.Request) {
	politicianIDStr := chi.URLParam(r, "politicianId")
	politicianID, err := uuid.Parse(politicianIDStr)
	if err != nil {
		WriteBadRequest(w, "Invalid politician ID")
		return
	}

	jurisdictions, err := h.partyService.GetJurisdictionsByPolitician(r.Context(), politicianID)
	if err != nil {
		WriteInternalError(w, "Failed to get jurisdictions")
		return
	}

	WriteSuccess(w, jurisdictions)
}

// DeleteJurisdiction deletes a jurisdiction
func (h *PoliticalPartyHandler) DeleteJurisdiction(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "Invalid jurisdiction ID")
		return
	}

	err = h.partyService.DeleteJurisdiction(r.Context(), id)
	if err != nil {
		WriteInternalError(w, "Failed to delete jurisdiction")
		return
	}

	WriteSuccess(w, map[string]string{"message": "Jurisdiction deleted successfully"})
}
