package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/humfurie/pulpulitiko/api/internal/models"
	"github.com/humfurie/pulpulitiko/api/internal/services"
)

type LocationHandler struct {
	locationService *services.LocationService
}

func NewLocationHandler(locationService *services.LocationService) *LocationHandler {
	return &LocationHandler{
		locationService: locationService,
	}
}

// =====================================================
// PUBLIC ENDPOINTS
// =====================================================

// GET /api/locations/regions - List all regions
func (h *LocationHandler) ListRegions(w http.ResponseWriter, r *http.Request) {
	regions, err := h.locationService.ListRegions(r.Context())
	if err != nil {
		WriteInternalError(w, "failed to fetch regions")
		return
	}

	WriteSuccess(w, regions)
}

// GET /api/locations/regions/{slug} - Get region by slug with provinces
func (h *LocationHandler) GetRegionBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		WriteBadRequest(w, "slug is required")
		return
	}

	region, err := h.locationService.GetRegionBySlug(r.Context(), slug)
	if err != nil {
		WriteInternalError(w, "failed to fetch region")
		return
	}

	if region == nil {
		WriteNotFound(w, "region not found")
		return
	}

	// Get provinces for this region
	provinces, err := h.locationService.ListProvincesByRegion(r.Context(), region.ID)
	if err != nil {
		WriteInternalError(w, "failed to fetch provinces")
		return
	}

	WriteSuccess(w, map[string]interface{}{
		"region":    region,
		"provinces": provinces,
	})
}

// GET /api/locations/provinces - List all provinces
func (h *LocationHandler) ListAllProvinces(w http.ResponseWriter, r *http.Request) {
	provinces, err := h.locationService.ListAllProvinces(r.Context())
	if err != nil {
		WriteInternalError(w, "failed to fetch provinces")
		return
	}

	WriteSuccess(w, provinces)
}

// GET /api/locations/provinces/{slug} - Get province by slug with cities
func (h *LocationHandler) GetProvinceBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		WriteBadRequest(w, "slug is required")
		return
	}

	province, err := h.locationService.GetProvinceBySlug(r.Context(), slug)
	if err != nil {
		WriteInternalError(w, "failed to fetch province")
		return
	}

	if province == nil {
		WriteNotFound(w, "province not found")
		return
	}

	// Get cities for this province
	cities, err := h.locationService.ListCitiesByProvince(r.Context(), province.ID)
	if err != nil {
		WriteInternalError(w, "failed to fetch cities")
		return
	}

	WriteSuccess(w, map[string]interface{}{
		"province": province,
		"cities":   cities,
	})
}

// GET /api/locations/cities/{slug} - Get city by slug with barangays
func (h *LocationHandler) GetCityBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		WriteBadRequest(w, "slug is required")
		return
	}

	city, err := h.locationService.GetCityMunicipalityBySlug(r.Context(), slug)
	if err != nil {
		WriteInternalError(w, "failed to fetch city")
		return
	}

	if city == nil {
		WriteNotFound(w, "city/municipality not found")
		return
	}

	// Get barangays for this city (paginated)
	page, perPage := GetPaginationParams(r)
	barangays, err := h.locationService.ListBarangaysByCity(r.Context(), city.ID, page, perPage)
	if err != nil {
		WriteInternalError(w, "failed to fetch barangays")
		return
	}

	WriteSuccess(w, map[string]interface{}{
		"city":      city,
		"barangays": barangays,
	})
}

// GET /api/locations/barangays/{slug} - Get barangay by slug
func (h *LocationHandler) GetBarangayBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		WriteBadRequest(w, "slug is required")
		return
	}

	barangay, err := h.locationService.GetBarangayBySlug(r.Context(), slug)
	if err != nil {
		WriteInternalError(w, "failed to fetch barangay")
		return
	}

	if barangay == nil {
		WriteNotFound(w, "barangay not found")
		return
	}

	WriteSuccess(w, barangay)
}

// GET /api/locations/districts/{slug} - Get district by slug
func (h *LocationHandler) GetDistrictBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		WriteBadRequest(w, "slug is required")
		return
	}

	district, err := h.locationService.GetDistrictBySlug(r.Context(), slug)
	if err != nil {
		WriteInternalError(w, "failed to fetch district")
		return
	}

	if district == nil {
		WriteNotFound(w, "district not found")
		return
	}

	WriteSuccess(w, district)
}

// GET /api/locations/search?q= - Search locations
func (h *LocationHandler) SearchLocations(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		WriteSuccess(w, []models.LocationSearchResult{})
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 20
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 50 {
			limit = l
		}
	}

	results, err := h.locationService.SearchLocations(r.Context(), query, limit)
	if err != nil {
		WriteInternalError(w, "failed to search locations")
		return
	}

	WriteSuccess(w, results)
}

// GET /api/locations/hierarchy/{barangay_id} - Get full location hierarchy
func (h *LocationHandler) GetHierarchy(w http.ResponseWriter, r *http.Request) {
	barangayIDStr := chi.URLParam(r, "barangay_id")
	barangayID, err := uuid.Parse(barangayIDStr)
	if err != nil {
		WriteBadRequest(w, "invalid barangay ID")
		return
	}

	hierarchy, err := h.locationService.GetLocationHierarchy(r.Context(), barangayID)
	if err != nil {
		WriteInternalError(w, "failed to fetch location hierarchy")
		return
	}

	if hierarchy == nil {
		WriteNotFound(w, "barangay not found")
		return
	}

	WriteSuccess(w, hierarchy)
}

// =====================================================
// CASCADING ENDPOINTS (for LocationPicker component)
// =====================================================

// GET /api/locations/provinces/by-region/{region_id} - Get provinces by region ID
func (h *LocationHandler) GetProvincesByRegion(w http.ResponseWriter, r *http.Request) {
	regionIDStr := chi.URLParam(r, "region_id")
	regionID, err := uuid.Parse(regionIDStr)
	if err != nil {
		WriteBadRequest(w, "invalid region ID")
		return
	}

	provinces, err := h.locationService.ListProvincesByRegion(r.Context(), regionID)
	if err != nil {
		WriteInternalError(w, "failed to fetch provinces")
		return
	}

	WriteSuccess(w, provinces)
}

// GET /api/locations/cities/by-province/{province_id} - Get cities by province ID
func (h *LocationHandler) GetCitiesByProvince(w http.ResponseWriter, r *http.Request) {
	provinceIDStr := chi.URLParam(r, "province_id")
	provinceID, err := uuid.Parse(provinceIDStr)
	if err != nil {
		WriteBadRequest(w, "invalid province ID")
		return
	}

	cities, err := h.locationService.ListCitiesByProvince(r.Context(), provinceID)
	if err != nil {
		WriteInternalError(w, "failed to fetch cities")
		return
	}

	WriteSuccess(w, cities)
}

// GET /api/locations/barangays/by-city/{city_id} - Get barangays by city ID (paginated)
func (h *LocationHandler) GetBarangaysByCity(w http.ResponseWriter, r *http.Request) {
	cityIDStr := chi.URLParam(r, "city_id")
	cityID, err := uuid.Parse(cityIDStr)
	if err != nil {
		WriteBadRequest(w, "invalid city ID")
		return
	}

	page, perPage := GetPaginationParams(r)
	barangays, err := h.locationService.ListBarangaysByCity(r.Context(), cityID, page, perPage)
	if err != nil {
		WriteInternalError(w, "failed to fetch barangays")
		return
	}

	WriteSuccess(w, barangays)
}

// GET /api/locations/districts/by-province/{province_id} - Get districts by province ID
func (h *LocationHandler) GetDistrictsByProvince(w http.ResponseWriter, r *http.Request) {
	provinceIDStr := chi.URLParam(r, "province_id")
	provinceID, err := uuid.Parse(provinceIDStr)
	if err != nil {
		WriteBadRequest(w, "invalid province ID")
		return
	}

	districts, err := h.locationService.ListDistrictsByProvince(r.Context(), provinceID)
	if err != nil {
		WriteInternalError(w, "failed to fetch districts")
		return
	}

	WriteSuccess(w, districts)
}

// =====================================================
// ADMIN ENDPOINTS
// =====================================================

// GET /api/admin/locations/regions/{id} - Get region by ID
func (h *LocationHandler) AdminGetRegionByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "invalid region ID")
		return
	}

	region, err := h.locationService.GetRegionByID(r.Context(), id)
	if err != nil {
		WriteInternalError(w, "failed to fetch region")
		return
	}

	if region == nil {
		WriteNotFound(w, "region not found")
		return
	}

	WriteSuccess(w, region)
}

// POST /api/admin/locations/regions - Create region
func (h *LocationHandler) CreateRegion(w http.ResponseWriter, r *http.Request) {
	var req models.CreateRegionRequest
	if err := DecodeAndValidate(r, &req); err != nil {
		WriteValidationError(w, err)
		return
	}

	region, err := h.locationService.CreateRegion(r.Context(), &req)
	if err != nil {
		WriteInternalError(w, err.Error())
		return
	}

	WriteCreated(w, region)
}

// PUT /api/admin/locations/regions/{id} - Update region
func (h *LocationHandler) UpdateRegion(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "invalid region ID")
		return
	}

	var req models.UpdateRegionRequest
	if err := DecodeAndValidate(r, &req); err != nil {
		WriteValidationError(w, err)
		return
	}

	region, err := h.locationService.UpdateRegion(r.Context(), id, &req)
	if err != nil {
		WriteInternalError(w, err.Error())
		return
	}

	WriteSuccess(w, region)
}

// DELETE /api/admin/locations/regions/{id} - Delete region
func (h *LocationHandler) DeleteRegion(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "invalid region ID")
		return
	}

	if err := h.locationService.DeleteRegion(r.Context(), id); err != nil {
		WriteInternalError(w, err.Error())
		return
	}

	WriteSuccess(w, map[string]string{"message": "region deleted"})
}

// POST /api/admin/locations/provinces - Create province
func (h *LocationHandler) CreateProvince(w http.ResponseWriter, r *http.Request) {
	var req models.CreateProvinceRequest
	if err := DecodeAndValidate(r, &req); err != nil {
		WriteValidationError(w, err)
		return
	}

	province, err := h.locationService.CreateProvince(r.Context(), &req)
	if err != nil {
		WriteInternalError(w, err.Error())
		return
	}

	WriteCreated(w, province)
}

// GET /api/admin/locations/provinces/{id} - Get province by ID
func (h *LocationHandler) AdminGetProvinceByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "invalid province ID")
		return
	}

	province, err := h.locationService.GetProvinceByID(r.Context(), id)
	if err != nil {
		WriteInternalError(w, "failed to fetch province")
		return
	}

	if province == nil {
		WriteNotFound(w, "province not found")
		return
	}

	WriteSuccess(w, province)
}

// PUT /api/admin/locations/provinces/{id} - Update province
func (h *LocationHandler) UpdateProvince(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "invalid province ID")
		return
	}

	var req models.UpdateProvinceRequest
	if err := DecodeAndValidate(r, &req); err != nil {
		WriteValidationError(w, err)
		return
	}

	province, err := h.locationService.UpdateProvince(r.Context(), id, &req)
	if err != nil {
		WriteInternalError(w, err.Error())
		return
	}

	WriteSuccess(w, province)
}

// DELETE /api/admin/locations/provinces/{id} - Delete province
func (h *LocationHandler) DeleteProvince(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "invalid province ID")
		return
	}

	if err := h.locationService.DeleteProvince(r.Context(), id); err != nil {
		WriteInternalError(w, err.Error())
		return
	}

	WriteSuccess(w, map[string]string{"message": "province deleted"})
}

// POST /api/admin/locations/cities - Create city/municipality
func (h *LocationHandler) CreateCity(w http.ResponseWriter, r *http.Request) {
	var req models.CreateCityMunicipalityRequest
	if err := DecodeAndValidate(r, &req); err != nil {
		WriteValidationError(w, err)
		return
	}

	city, err := h.locationService.CreateCityMunicipality(r.Context(), &req)
	if err != nil {
		WriteInternalError(w, err.Error())
		return
	}

	WriteCreated(w, city)
}

// GET /api/admin/locations/cities/{id} - Get city by ID
func (h *LocationHandler) AdminGetCityByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "invalid city ID")
		return
	}

	city, err := h.locationService.GetCityMunicipalityByID(r.Context(), id)
	if err != nil {
		WriteInternalError(w, "failed to fetch city")
		return
	}

	if city == nil {
		WriteNotFound(w, "city not found")
		return
	}

	WriteSuccess(w, city)
}

// PUT /api/admin/locations/cities/{id} - Update city
func (h *LocationHandler) UpdateCity(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "invalid city ID")
		return
	}

	var req models.UpdateCityMunicipalityRequest
	if err := DecodeAndValidate(r, &req); err != nil {
		WriteValidationError(w, err)
		return
	}

	city, err := h.locationService.UpdateCityMunicipality(r.Context(), id, &req)
	if err != nil {
		WriteInternalError(w, err.Error())
		return
	}

	WriteSuccess(w, city)
}

// DELETE /api/admin/locations/cities/{id} - Delete city
func (h *LocationHandler) DeleteCity(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "invalid city ID")
		return
	}

	if err := h.locationService.DeleteCityMunicipality(r.Context(), id); err != nil {
		WriteInternalError(w, err.Error())
		return
	}

	WriteSuccess(w, map[string]string{"message": "city deleted"})
}

// POST /api/admin/locations/barangays - Create barangay
func (h *LocationHandler) CreateBarangay(w http.ResponseWriter, r *http.Request) {
	var req models.CreateBarangayRequest
	if err := DecodeAndValidate(r, &req); err != nil {
		WriteValidationError(w, err)
		return
	}

	barangay, err := h.locationService.CreateBarangay(r.Context(), &req)
	if err != nil {
		WriteInternalError(w, err.Error())
		return
	}

	WriteCreated(w, barangay)
}

// GET /api/admin/locations/barangays/{id} - Get barangay by ID
func (h *LocationHandler) AdminGetBarangayByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "invalid barangay ID")
		return
	}

	barangay, err := h.locationService.GetBarangayByID(r.Context(), id)
	if err != nil {
		WriteInternalError(w, "failed to fetch barangay")
		return
	}

	if barangay == nil {
		WriteNotFound(w, "barangay not found")
		return
	}

	WriteSuccess(w, barangay)
}

// PUT /api/admin/locations/barangays/{id} - Update barangay
func (h *LocationHandler) UpdateBarangay(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "invalid barangay ID")
		return
	}

	var req models.UpdateBarangayRequest
	if err := DecodeAndValidate(r, &req); err != nil {
		WriteValidationError(w, err)
		return
	}

	barangay, err := h.locationService.UpdateBarangay(r.Context(), id, &req)
	if err != nil {
		WriteInternalError(w, err.Error())
		return
	}

	WriteSuccess(w, barangay)
}

// DELETE /api/admin/locations/barangays/{id} - Delete barangay
func (h *LocationHandler) DeleteBarangay(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "invalid barangay ID")
		return
	}

	if err := h.locationService.DeleteBarangay(r.Context(), id); err != nil {
		WriteInternalError(w, err.Error())
		return
	}

	WriteSuccess(w, map[string]string{"message": "barangay deleted"})
}

// POST /api/admin/locations/districts - Create congressional district
func (h *LocationHandler) CreateDistrict(w http.ResponseWriter, r *http.Request) {
	var req models.CreateDistrictRequest
	if err := DecodeAndValidate(r, &req); err != nil {
		WriteValidationError(w, err)
		return
	}

	district, err := h.locationService.CreateDistrict(r.Context(), &req)
	if err != nil {
		WriteInternalError(w, err.Error())
		return
	}

	WriteCreated(w, district)
}

// GET /api/admin/locations/districts/{id} - Get district by ID
func (h *LocationHandler) AdminGetDistrictByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "invalid district ID")
		return
	}

	district, err := h.locationService.GetDistrictByID(r.Context(), id)
	if err != nil {
		WriteInternalError(w, "failed to fetch district")
		return
	}

	if district == nil {
		WriteNotFound(w, "district not found")
		return
	}

	WriteSuccess(w, district)
}
