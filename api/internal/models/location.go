package models

import (
	"time"

	"github.com/google/uuid"
)

// Region represents a Philippine region (e.g., NCR, Region I)
type Region struct {
	ID        uuid.UUID  `json:"id"`
	Code      string     `json:"code"`
	Name      string     `json:"name"`
	Slug      string     `json:"slug"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`

	// Relations (populated when needed)
	Provinces     []Province `json:"provinces,omitempty"`
	ProvinceCount int        `json:"province_count,omitempty"`
}

// Province represents a Philippine province
type Province struct {
	ID        uuid.UUID  `json:"id"`
	RegionID  uuid.UUID  `json:"region_id"`
	Code      string     `json:"code"`
	Name      string     `json:"name"`
	Slug      string     `json:"slug"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`

	// Relations
	Region               *Region            `json:"region,omitempty"`
	CitiesMunicipalities []CityMunicipality `json:"cities_municipalities,omitempty"`
	CityCount            int                `json:"city_count,omitempty"`
}

// CityMunicipality represents a city or municipality
type CityMunicipality struct {
	ID         uuid.UUID  `json:"id"`
	ProvinceID uuid.UUID  `json:"province_id"`
	Code       string     `json:"code"`
	Name       string     `json:"name"`
	Slug       string     `json:"slug"`
	IsCity     bool       `json:"is_city"`
	IsCapital  bool       `json:"is_capital"`
	IsHUC      bool       `json:"is_huc"` // Highly Urbanized City
	IsICC      bool       `json:"is_icc"` // Independent Component City
	Population *int       `json:"population,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty"`

	// Relations
	Province      *Province  `json:"province,omitempty"`
	Barangays     []Barangay `json:"barangays,omitempty"`
	BarangayCount int        `json:"barangay_count,omitempty"`
}

// Barangay represents a barangay (smallest administrative division)
type Barangay struct {
	ID                 uuid.UUID  `json:"id"`
	CityMunicipalityID uuid.UUID  `json:"city_municipality_id"`
	Code               string     `json:"code"`
	Name               string     `json:"name"`
	Slug               string     `json:"slug"`
	Population         *int       `json:"population,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	DeletedAt          *time.Time `json:"deleted_at,omitempty"`

	// Relations
	CityMunicipality *CityMunicipality `json:"city_municipality,omitempty"`
}

// CongressionalDistrict represents a congressional district
type CongressionalDistrict struct {
	ID                 uuid.UUID  `json:"id"`
	ProvinceID         *uuid.UUID `json:"province_id,omitempty"`
	CityMunicipalityID *uuid.UUID `json:"city_municipality_id,omitempty"` // For lone/HUC districts
	DistrictNumber     int        `json:"district_number"`
	Name               string     `json:"name"`
	Slug               string     `json:"slug"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	DeletedAt          *time.Time `json:"deleted_at,omitempty"`

	// Relations
	Province         *Province          `json:"province,omitempty"`
	CityMunicipality *CityMunicipality  `json:"city_municipality,omitempty"`
	Coverage         []CityMunicipality `json:"coverage,omitempty"` // Cities/municipalities in this district
}

// =====================================================
// LIST ITEMS (Lightweight for dropdowns/lists)
// =====================================================

type RegionListItem struct {
	ID            uuid.UUID `json:"id"`
	Code          string    `json:"code"`
	Name          string    `json:"name"`
	Slug          string    `json:"slug"`
	ProvinceCount int       `json:"province_count"`
}

type ProvinceListItem struct {
	ID         uuid.UUID `json:"id"`
	RegionID   uuid.UUID `json:"region_id"`
	Code       string    `json:"code"`
	Name       string    `json:"name"`
	Slug       string    `json:"slug"`
	RegionName string    `json:"region_name,omitempty"`
	CityCount  int       `json:"city_count"`
}

type CityMunicipalityListItem struct {
	ID            uuid.UUID `json:"id"`
	ProvinceID    uuid.UUID `json:"province_id"`
	Code          string    `json:"code"`
	Name          string    `json:"name"`
	Slug          string    `json:"slug"`
	IsCity        bool      `json:"is_city"`
	IsCapital     bool      `json:"is_capital"`
	IsHUC         bool      `json:"is_huc"`
	ProvinceName  string    `json:"province_name,omitempty"`
	BarangayCount int       `json:"barangay_count"`
}

type BarangayListItem struct {
	ID                   uuid.UUID `json:"id"`
	CityMunicipalityID   uuid.UUID `json:"city_municipality_id"`
	Code                 string    `json:"code"`
	Name                 string    `json:"name"`
	Slug                 string    `json:"slug"`
	CityMunicipalityName string    `json:"city_municipality_name,omitempty"`
}

type DistrictListItem struct {
	ID             uuid.UUID `json:"id"`
	DistrictNumber int       `json:"district_number"`
	Name           string    `json:"name"`
	Slug           string    `json:"slug"`
	ProvinceName   string    `json:"province_name,omitempty"`
	CityName       string    `json:"city_name,omitempty"` // For lone districts
}

// =====================================================
// HIERARCHICAL VIEW (Full location path)
// =====================================================

type LocationHierarchy struct {
	Region           *RegionListItem           `json:"region,omitempty"`
	Province         *ProvinceListItem         `json:"province,omitempty"`
	CityMunicipality *CityMunicipalityListItem `json:"city_municipality,omitempty"`
	Barangay         *BarangayListItem         `json:"barangay,omitempty"`
	District         *DistrictListItem         `json:"district,omitempty"`
}

// =====================================================
// REQUEST/RESPONSE TYPES
// =====================================================

// Create Requests
type CreateRegionRequest struct {
	Code string `json:"code" validate:"required,max=20"`
	Name string `json:"name" validate:"required,max=200"`
	Slug string `json:"slug" validate:"required,max=200"`
}

type CreateProvinceRequest struct {
	RegionID string `json:"region_id" validate:"required,uuid"`
	Code     string `json:"code" validate:"required,max=20"`
	Name     string `json:"name" validate:"required,max=200"`
	Slug     string `json:"slug" validate:"required,max=200"`
}

type CreateCityMunicipalityRequest struct {
	ProvinceID string `json:"province_id" validate:"required,uuid"`
	Code       string `json:"code" validate:"required,max=20"`
	Name       string `json:"name" validate:"required,max=200"`
	Slug       string `json:"slug" validate:"required,max=200"`
	IsCity     bool   `json:"is_city"`
	IsCapital  bool   `json:"is_capital"`
	IsHUC      bool   `json:"is_huc"`
	IsICC      bool   `json:"is_icc"`
	Population *int   `json:"population,omitempty"`
}

type CreateBarangayRequest struct {
	CityMunicipalityID string `json:"city_municipality_id" validate:"required,uuid"`
	Code               string `json:"code" validate:"required,max=20"`
	Name               string `json:"name" validate:"required,max=200"`
	Slug               string `json:"slug" validate:"required,max=200"`
	Population         *int   `json:"population,omitempty"`
}

type CreateDistrictRequest struct {
	ProvinceID         *string `json:"province_id,omitempty" validate:"omitempty,uuid"`
	CityMunicipalityID *string `json:"city_municipality_id,omitempty" validate:"omitempty,uuid"`
	DistrictNumber     int     `json:"district_number" validate:"required,min=1"`
	Name               string  `json:"name" validate:"required,max=200"`
	Slug               string  `json:"slug" validate:"required,max=200"`
}

// Update Requests
type UpdateRegionRequest struct {
	Code *string `json:"code,omitempty" validate:"omitempty,max=20"`
	Name *string `json:"name,omitempty" validate:"omitempty,max=200"`
	Slug *string `json:"slug,omitempty" validate:"omitempty,max=200"`
}

type UpdateProvinceRequest struct {
	RegionID *string `json:"region_id,omitempty" validate:"omitempty,uuid"`
	Code     *string `json:"code,omitempty" validate:"omitempty,max=20"`
	Name     *string `json:"name,omitempty" validate:"omitempty,max=200"`
	Slug     *string `json:"slug,omitempty" validate:"omitempty,max=200"`
}

type UpdateCityMunicipalityRequest struct {
	ProvinceID *string `json:"province_id,omitempty" validate:"omitempty,uuid"`
	Code       *string `json:"code,omitempty" validate:"omitempty,max=20"`
	Name       *string `json:"name,omitempty" validate:"omitempty,max=200"`
	Slug       *string `json:"slug,omitempty" validate:"omitempty,max=200"`
	IsCity     *bool   `json:"is_city,omitempty"`
	IsCapital  *bool   `json:"is_capital,omitempty"`
	IsHUC      *bool   `json:"is_huc,omitempty"`
	IsICC      *bool   `json:"is_icc,omitempty"`
	Population *int    `json:"population,omitempty"`
}

type UpdateBarangayRequest struct {
	CityMunicipalityID *string `json:"city_municipality_id,omitempty" validate:"omitempty,uuid"`
	Code               *string `json:"code,omitempty" validate:"omitempty,max=20"`
	Name               *string `json:"name,omitempty" validate:"omitempty,max=200"`
	Slug               *string `json:"slug,omitempty" validate:"omitempty,max=200"`
	Population         *int    `json:"population,omitempty"`
}

// Filters
type LocationFilter struct {
	Search         *string
	RegionID       *uuid.UUID
	ProvinceID     *uuid.UUID
	CityID         *uuid.UUID
	IncludeDeleted bool
}

// Paginated Responses
type PaginatedRegions struct {
	Regions    []RegionListItem `json:"regions"`
	Total      int              `json:"total"`
	Page       int              `json:"page"`
	PerPage    int              `json:"per_page"`
	TotalPages int              `json:"total_pages"`
}

type PaginatedProvinces struct {
	Provinces  []ProvinceListItem `json:"provinces"`
	Total      int                `json:"total"`
	Page       int                `json:"page"`
	PerPage    int                `json:"per_page"`
	TotalPages int                `json:"total_pages"`
}

type PaginatedCitiesMunicipalities struct {
	CitiesMunicipalities []CityMunicipalityListItem `json:"cities_municipalities"`
	Total                int                        `json:"total"`
	Page                 int                        `json:"page"`
	PerPage              int                        `json:"per_page"`
	TotalPages           int                        `json:"total_pages"`
}

type PaginatedBarangays struct {
	Barangays  []BarangayListItem `json:"barangays"`
	Total      int                `json:"total"`
	Page       int                `json:"page"`
	PerPage    int                `json:"per_page"`
	TotalPages int                `json:"total_pages"`
}

type PaginatedDistricts struct {
	Districts  []DistrictListItem `json:"districts"`
	Total      int                `json:"total"`
	Page       int                `json:"page"`
	PerPage    int                `json:"per_page"`
	TotalPages int                `json:"total_pages"`
}

// Bulk Import Request (for PSGC data import)
type BulkLocationImportRequest struct {
	Regions             []CreateRegionRequest           `json:"regions,omitempty"`
	Provinces           []CreateProvinceRequest         `json:"provinces,omitempty"`
	CitiesMunicpalities []CreateCityMunicipalityRequest `json:"cities_municipalities,omitempty"`
	Barangays           []CreateBarangayRequest         `json:"barangays,omitempty"`
}

type BulkImportResult struct {
	RegionsCreated   int      `json:"regions_created"`
	ProvincesCreated int      `json:"provinces_created"`
	CitiesCreated    int      `json:"cities_created"`
	BarangaysCreated int      `json:"barangays_created"`
	Errors           []string `json:"errors,omitempty"`
}

// Search Result (unified search across all location types)
type LocationSearchResult struct {
	Type       string    `json:"type"` // "region", "province", "city", "barangay"
	ID         uuid.UUID `json:"id"`
	Code       string    `json:"code"`
	Name       string    `json:"name"`
	Slug       string    `json:"slug"`
	ParentName string    `json:"parent_name,omitempty"` // For display context
	FullPath   string    `json:"full_path"`             // e.g., "Barangay 1, Quezon City, NCR"
}
