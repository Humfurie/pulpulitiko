package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/humfurie/pulpulitiko/api/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LocationRepository struct {
	db *pgxpool.Pool
}

func NewLocationRepository(db *pgxpool.Pool) *LocationRepository {
	return &LocationRepository{db: db}
}

// =====================================================
// REGIONS
// =====================================================

func (r *LocationRepository) CreateRegion(ctx context.Context, region *models.Region) error {
	query := `
		INSERT INTO regions (code, name, slug)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(ctx, query, region.Code, region.Name, region.Slug).
		Scan(&region.ID, &region.CreatedAt, &region.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create region: %w", err)
	}

	return nil
}

func (r *LocationRepository) GetRegionByID(ctx context.Context, id uuid.UUID) (*models.Region, error) {
	query := `
		SELECT id, code, name, slug, created_at, updated_at, deleted_at
		FROM regions
		WHERE id = $1 AND deleted_at IS NULL
	`

	region := &models.Region{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&region.ID, &region.Code, &region.Name, &region.Slug,
		&region.CreatedAt, &region.UpdatedAt, &region.DeletedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get region: %w", err)
	}

	return region, nil
}

func (r *LocationRepository) GetRegionBySlug(ctx context.Context, slug string) (*models.Region, error) {
	query := `
		SELECT id, code, name, slug, created_at, updated_at, deleted_at
		FROM regions
		WHERE slug = $1 AND deleted_at IS NULL
	`

	region := &models.Region{}
	err := r.db.QueryRow(ctx, query, slug).Scan(
		&region.ID, &region.Code, &region.Name, &region.Slug,
		&region.CreatedAt, &region.UpdatedAt, &region.DeletedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get region by slug: %w", err)
	}

	return region, nil
}

func (r *LocationRepository) ListRegions(ctx context.Context) ([]models.RegionListItem, error) {
	query := `
		SELECT r.id, r.code, r.name, r.slug,
			(SELECT COUNT(*) FROM provinces p WHERE p.region_id = r.id AND p.deleted_at IS NULL) as province_count
		FROM regions r
		WHERE r.deleted_at IS NULL
		ORDER BY r.name ASC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list regions: %w", err)
	}
	defer rows.Close()

	regions := []models.RegionListItem{}
	for rows.Next() {
		var region models.RegionListItem
		err := rows.Scan(&region.ID, &region.Code, &region.Name, &region.Slug, &region.ProvinceCount)
		if err != nil {
			return nil, fmt.Errorf("failed to scan region: %w", err)
		}
		regions = append(regions, region)
	}

	return regions, nil
}

func (r *LocationRepository) UpdateRegion(ctx context.Context, id uuid.UUID, req *models.UpdateRegionRequest) error {
	query := `
		UPDATE regions
		SET code = COALESCE($1, code),
			name = COALESCE($2, name),
			slug = COALESCE($3, slug),
			updated_at = NOW()
		WHERE id = $4 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(ctx, query, req.Code, req.Name, req.Slug, id)
	if err != nil {
		return fmt.Errorf("failed to update region: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("region not found")
	}

	return nil
}

func (r *LocationRepository) DeleteRegion(ctx context.Context, id uuid.UUID) error {
	query := "UPDATE regions SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL"

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete region: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("region not found")
	}

	return nil
}

// =====================================================
// PROVINCES
// =====================================================

func (r *LocationRepository) CreateProvince(ctx context.Context, province *models.Province) error {
	query := `
		INSERT INTO provinces (region_id, code, name, slug)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(ctx, query, province.RegionID, province.Code, province.Name, province.Slug).
		Scan(&province.ID, &province.CreatedAt, &province.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create province: %w", err)
	}

	return nil
}

func (r *LocationRepository) GetProvinceByID(ctx context.Context, id uuid.UUID) (*models.Province, error) {
	query := `
		SELECT p.id, p.region_id, p.code, p.name, p.slug, p.created_at, p.updated_at, p.deleted_at,
			r.id, r.code, r.name, r.slug
		FROM provinces p
		LEFT JOIN regions r ON p.region_id = r.id
		WHERE p.id = $1 AND p.deleted_at IS NULL
	`

	province := &models.Province{Region: &models.Region{}}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&province.ID, &province.RegionID, &province.Code, &province.Name, &province.Slug,
		&province.CreatedAt, &province.UpdatedAt, &province.DeletedAt,
		&province.Region.ID, &province.Region.Code, &province.Region.Name, &province.Region.Slug,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get province: %w", err)
	}

	return province, nil
}

func (r *LocationRepository) GetProvinceBySlug(ctx context.Context, slug string) (*models.Province, error) {
	query := `
		SELECT p.id, p.region_id, p.code, p.name, p.slug, p.created_at, p.updated_at, p.deleted_at,
			r.id, r.code, r.name, r.slug
		FROM provinces p
		LEFT JOIN regions r ON p.region_id = r.id
		WHERE p.slug = $1 AND p.deleted_at IS NULL
	`

	province := &models.Province{Region: &models.Region{}}
	err := r.db.QueryRow(ctx, query, slug).Scan(
		&province.ID, &province.RegionID, &province.Code, &province.Name, &province.Slug,
		&province.CreatedAt, &province.UpdatedAt, &province.DeletedAt,
		&province.Region.ID, &province.Region.Code, &province.Region.Name, &province.Region.Slug,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get province by slug: %w", err)
	}

	return province, nil
}

func (r *LocationRepository) ListProvincesByRegion(ctx context.Context, regionID uuid.UUID) ([]models.ProvinceListItem, error) {
	query := `
		SELECT p.id, p.region_id, p.code, p.name, p.slug, r.name as region_name,
			(SELECT COUNT(*) FROM cities_municipalities c WHERE c.province_id = p.id AND c.deleted_at IS NULL) as city_count
		FROM provinces p
		LEFT JOIN regions r ON p.region_id = r.id
		WHERE p.region_id = $1 AND p.deleted_at IS NULL
		ORDER BY p.name ASC
	`

	rows, err := r.db.Query(ctx, query, regionID)
	if err != nil {
		return nil, fmt.Errorf("failed to list provinces: %w", err)
	}
	defer rows.Close()

	provinces := []models.ProvinceListItem{}
	for rows.Next() {
		var province models.ProvinceListItem
		err := rows.Scan(&province.ID, &province.RegionID, &province.Code, &province.Name, &province.Slug, &province.RegionName, &province.CityCount)
		if err != nil {
			return nil, fmt.Errorf("failed to scan province: %w", err)
		}
		provinces = append(provinces, province)
	}

	return provinces, nil
}

func (r *LocationRepository) ListAllProvinces(ctx context.Context) ([]models.ProvinceListItem, error) {
	query := `
		SELECT p.id, p.region_id, p.code, p.name, p.slug, r.name as region_name,
			(SELECT COUNT(*) FROM cities_municipalities c WHERE c.province_id = p.id AND c.deleted_at IS NULL) as city_count
		FROM provinces p
		LEFT JOIN regions r ON p.region_id = r.id
		WHERE p.deleted_at IS NULL
		ORDER BY r.name ASC, p.name ASC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list all provinces: %w", err)
	}
	defer rows.Close()

	provinces := []models.ProvinceListItem{}
	for rows.Next() {
		var province models.ProvinceListItem
		err := rows.Scan(&province.ID, &province.RegionID, &province.Code, &province.Name, &province.Slug, &province.RegionName, &province.CityCount)
		if err != nil {
			return nil, fmt.Errorf("failed to scan province: %w", err)
		}
		provinces = append(provinces, province)
	}

	return provinces, nil
}

func (r *LocationRepository) UpdateProvince(ctx context.Context, id uuid.UUID, req *models.UpdateProvinceRequest) error {
	var regionID *uuid.UUID
	if req.RegionID != nil {
		parsed, err := uuid.Parse(*req.RegionID)
		if err != nil {
			return fmt.Errorf("invalid region_id: %w", err)
		}
		regionID = &parsed
	}

	query := `
		UPDATE provinces
		SET region_id = COALESCE($1, region_id),
			code = COALESCE($2, code),
			name = COALESCE($3, name),
			slug = COALESCE($4, slug),
			updated_at = NOW()
		WHERE id = $5 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(ctx, query, regionID, req.Code, req.Name, req.Slug, id)
	if err != nil {
		return fmt.Errorf("failed to update province: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("province not found")
	}

	return nil
}

func (r *LocationRepository) DeleteProvince(ctx context.Context, id uuid.UUID) error {
	query := "UPDATE provinces SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL"

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete province: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("province not found")
	}

	return nil
}

// =====================================================
// CITIES/MUNICIPALITIES
// =====================================================

func (r *LocationRepository) CreateCityMunicipality(ctx context.Context, city *models.CityMunicipality) error {
	query := `
		INSERT INTO cities_municipalities (province_id, code, name, slug, is_city, is_capital, is_huc, is_icc, population)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(ctx, query,
		city.ProvinceID, city.Code, city.Name, city.Slug,
		city.IsCity, city.IsCapital, city.IsHUC, city.IsICC, city.Population,
	).Scan(&city.ID, &city.CreatedAt, &city.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create city/municipality: %w", err)
	}

	return nil
}

func (r *LocationRepository) GetCityMunicipalityByID(ctx context.Context, id uuid.UUID) (*models.CityMunicipality, error) {
	query := `
		SELECT c.id, c.province_id, c.code, c.name, c.slug, c.is_city, c.is_capital, c.is_huc, c.is_icc, c.population,
			c.created_at, c.updated_at, c.deleted_at,
			p.id, p.code, p.name, p.slug, p.region_id
		FROM cities_municipalities c
		LEFT JOIN provinces p ON c.province_id = p.id
		WHERE c.id = $1 AND c.deleted_at IS NULL
	`

	city := &models.CityMunicipality{Province: &models.Province{}}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&city.ID, &city.ProvinceID, &city.Code, &city.Name, &city.Slug,
		&city.IsCity, &city.IsCapital, &city.IsHUC, &city.IsICC, &city.Population,
		&city.CreatedAt, &city.UpdatedAt, &city.DeletedAt,
		&city.Province.ID, &city.Province.Code, &city.Province.Name, &city.Province.Slug, &city.Province.RegionID,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get city/municipality: %w", err)
	}

	return city, nil
}

func (r *LocationRepository) GetCityMunicipalityBySlug(ctx context.Context, slug string) (*models.CityMunicipality, error) {
	query := `
		SELECT c.id, c.province_id, c.code, c.name, c.slug, c.is_city, c.is_capital, c.is_huc, c.is_icc, c.population,
			c.created_at, c.updated_at, c.deleted_at,
			p.id, p.code, p.name, p.slug, p.region_id
		FROM cities_municipalities c
		LEFT JOIN provinces p ON c.province_id = p.id
		WHERE c.slug = $1 AND c.deleted_at IS NULL
	`

	city := &models.CityMunicipality{Province: &models.Province{}}
	err := r.db.QueryRow(ctx, query, slug).Scan(
		&city.ID, &city.ProvinceID, &city.Code, &city.Name, &city.Slug,
		&city.IsCity, &city.IsCapital, &city.IsHUC, &city.IsICC, &city.Population,
		&city.CreatedAt, &city.UpdatedAt, &city.DeletedAt,
		&city.Province.ID, &city.Province.Code, &city.Province.Name, &city.Province.Slug, &city.Province.RegionID,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get city/municipality by slug: %w", err)
	}

	return city, nil
}

func (r *LocationRepository) ListCitiesByProvince(ctx context.Context, provinceID uuid.UUID) ([]models.CityMunicipalityListItem, error) {
	query := `
		SELECT c.id, c.province_id, c.code, c.name, c.slug, c.is_city, c.is_capital, c.is_huc, p.name as province_name,
			(SELECT COUNT(*) FROM barangays b WHERE b.city_municipality_id = c.id AND b.deleted_at IS NULL) as barangay_count
		FROM cities_municipalities c
		LEFT JOIN provinces p ON c.province_id = p.id
		WHERE c.province_id = $1 AND c.deleted_at IS NULL
		ORDER BY c.is_capital DESC, c.is_city DESC, c.name ASC
	`

	rows, err := r.db.Query(ctx, query, provinceID)
	if err != nil {
		return nil, fmt.Errorf("failed to list cities: %w", err)
	}
	defer rows.Close()

	cities := []models.CityMunicipalityListItem{}
	for rows.Next() {
		var city models.CityMunicipalityListItem
		err := rows.Scan(&city.ID, &city.ProvinceID, &city.Code, &city.Name, &city.Slug, &city.IsCity, &city.IsCapital, &city.IsHUC, &city.ProvinceName, &city.BarangayCount)
		if err != nil {
			return nil, fmt.Errorf("failed to scan city: %w", err)
		}
		cities = append(cities, city)
	}

	return cities, nil
}

func (r *LocationRepository) UpdateCityMunicipality(ctx context.Context, id uuid.UUID, req *models.UpdateCityMunicipalityRequest) error {
	var provinceID *uuid.UUID
	if req.ProvinceID != nil {
		parsed, err := uuid.Parse(*req.ProvinceID)
		if err != nil {
			return fmt.Errorf("invalid province_id: %w", err)
		}
		provinceID = &parsed
	}

	query := `
		UPDATE cities_municipalities
		SET province_id = COALESCE($1, province_id),
			code = COALESCE($2, code),
			name = COALESCE($3, name),
			slug = COALESCE($4, slug),
			is_city = COALESCE($5, is_city),
			is_capital = COALESCE($6, is_capital),
			is_huc = COALESCE($7, is_huc),
			is_icc = COALESCE($8, is_icc),
			population = COALESCE($9, population),
			updated_at = NOW()
		WHERE id = $10 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(ctx, query,
		provinceID, req.Code, req.Name, req.Slug,
		req.IsCity, req.IsCapital, req.IsHUC, req.IsICC, req.Population, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update city/municipality: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("city/municipality not found")
	}

	return nil
}

func (r *LocationRepository) DeleteCityMunicipality(ctx context.Context, id uuid.UUID) error {
	query := "UPDATE cities_municipalities SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL"

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete city/municipality: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("city/municipality not found")
	}

	return nil
}

// =====================================================
// BARANGAYS
// =====================================================

func (r *LocationRepository) CreateBarangay(ctx context.Context, barangay *models.Barangay) error {
	query := `
		INSERT INTO barangays (city_municipality_id, code, name, slug, population)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(ctx, query,
		barangay.CityMunicipalityID, barangay.Code, barangay.Name, barangay.Slug, barangay.Population,
	).Scan(&barangay.ID, &barangay.CreatedAt, &barangay.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create barangay: %w", err)
	}

	return nil
}

func (r *LocationRepository) GetBarangayByID(ctx context.Context, id uuid.UUID) (*models.Barangay, error) {
	query := `
		SELECT b.id, b.city_municipality_id, b.code, b.name, b.slug, b.population,
			b.created_at, b.updated_at, b.deleted_at,
			c.id, c.code, c.name, c.slug, c.is_city, c.province_id
		FROM barangays b
		LEFT JOIN cities_municipalities c ON b.city_municipality_id = c.id
		WHERE b.id = $1 AND b.deleted_at IS NULL
	`

	barangay := &models.Barangay{CityMunicipality: &models.CityMunicipality{}}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&barangay.ID, &barangay.CityMunicipalityID, &barangay.Code, &barangay.Name, &barangay.Slug, &barangay.Population,
		&barangay.CreatedAt, &barangay.UpdatedAt, &barangay.DeletedAt,
		&barangay.CityMunicipality.ID, &barangay.CityMunicipality.Code, &barangay.CityMunicipality.Name,
		&barangay.CityMunicipality.Slug, &barangay.CityMunicipality.IsCity, &barangay.CityMunicipality.ProvinceID,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get barangay: %w", err)
	}

	return barangay, nil
}

func (r *LocationRepository) GetBarangayBySlug(ctx context.Context, slug string) (*models.Barangay, error) {
	query := `
		SELECT b.id, b.city_municipality_id, b.code, b.name, b.slug, b.population,
			b.created_at, b.updated_at, b.deleted_at,
			c.id, c.code, c.name, c.slug, c.is_city, c.province_id
		FROM barangays b
		LEFT JOIN cities_municipalities c ON b.city_municipality_id = c.id
		WHERE b.slug = $1 AND b.deleted_at IS NULL
	`

	barangay := &models.Barangay{CityMunicipality: &models.CityMunicipality{}}
	err := r.db.QueryRow(ctx, query, slug).Scan(
		&barangay.ID, &barangay.CityMunicipalityID, &barangay.Code, &barangay.Name, &barangay.Slug, &barangay.Population,
		&barangay.CreatedAt, &barangay.UpdatedAt, &barangay.DeletedAt,
		&barangay.CityMunicipality.ID, &barangay.CityMunicipality.Code, &barangay.CityMunicipality.Name,
		&barangay.CityMunicipality.Slug, &barangay.CityMunicipality.IsCity, &barangay.CityMunicipality.ProvinceID,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get barangay by slug: %w", err)
	}

	return barangay, nil
}

func (r *LocationRepository) ListBarangaysByCity(ctx context.Context, cityID uuid.UUID, page, perPage int) (*models.PaginatedBarangays, error) {
	// Get total count
	var total int
	countQuery := "SELECT COUNT(*) FROM barangays WHERE city_municipality_id = $1 AND deleted_at IS NULL"
	err := r.db.QueryRow(ctx, countQuery, cityID).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count barangays: %w", err)
	}

	offset := (page - 1) * perPage
	query := `
		SELECT b.id, b.city_municipality_id, b.code, b.name, b.slug, c.name as city_name
		FROM barangays b
		LEFT JOIN cities_municipalities c ON b.city_municipality_id = c.id
		WHERE b.city_municipality_id = $1 AND b.deleted_at IS NULL
		ORDER BY b.name ASC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, cityID, perPage, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list barangays: %w", err)
	}
	defer rows.Close()

	barangays := []models.BarangayListItem{}
	for rows.Next() {
		var barangay models.BarangayListItem
		err := rows.Scan(&barangay.ID, &barangay.CityMunicipalityID, &barangay.Code, &barangay.Name, &barangay.Slug, &barangay.CityMunicipalityName)
		if err != nil {
			return nil, fmt.Errorf("failed to scan barangay: %w", err)
		}
		barangays = append(barangays, barangay)
	}

	totalPages := (total + perPage - 1) / perPage

	return &models.PaginatedBarangays{
		Barangays:  barangays,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}, nil
}

func (r *LocationRepository) UpdateBarangay(ctx context.Context, id uuid.UUID, req *models.UpdateBarangayRequest) error {
	var cityID *uuid.UUID
	if req.CityMunicipalityID != nil {
		parsed, err := uuid.Parse(*req.CityMunicipalityID)
		if err != nil {
			return fmt.Errorf("invalid city_municipality_id: %w", err)
		}
		cityID = &parsed
	}

	query := `
		UPDATE barangays
		SET city_municipality_id = COALESCE($1, city_municipality_id),
			code = COALESCE($2, code),
			name = COALESCE($3, name),
			slug = COALESCE($4, slug),
			population = COALESCE($5, population),
			updated_at = NOW()
		WHERE id = $6 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(ctx, query, cityID, req.Code, req.Name, req.Slug, req.Population, id)
	if err != nil {
		return fmt.Errorf("failed to update barangay: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("barangay not found")
	}

	return nil
}

func (r *LocationRepository) DeleteBarangay(ctx context.Context, id uuid.UUID) error {
	query := "UPDATE barangays SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL"

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete barangay: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("barangay not found")
	}

	return nil
}

// =====================================================
// CONGRESSIONAL DISTRICTS
// =====================================================

func (r *LocationRepository) CreateDistrict(ctx context.Context, district *models.CongressionalDistrict) error {
	query := `
		INSERT INTO congressional_districts (province_id, city_municipality_id, district_number, name, slug)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(ctx, query,
		district.ProvinceID, district.CityMunicipalityID,
		district.DistrictNumber, district.Name, district.Slug,
	).Scan(&district.ID, &district.CreatedAt, &district.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create district: %w", err)
	}

	return nil
}

func (r *LocationRepository) GetDistrictByID(ctx context.Context, id uuid.UUID) (*models.CongressionalDistrict, error) {
	query := `
		SELECT d.id, d.province_id, d.city_municipality_id, d.district_number, d.name, d.slug,
			d.created_at, d.updated_at, d.deleted_at
		FROM congressional_districts d
		WHERE d.id = $1 AND d.deleted_at IS NULL
	`

	district := &models.CongressionalDistrict{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&district.ID, &district.ProvinceID, &district.CityMunicipalityID,
		&district.DistrictNumber, &district.Name, &district.Slug,
		&district.CreatedAt, &district.UpdatedAt, &district.DeletedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get district: %w", err)
	}

	return district, nil
}

func (r *LocationRepository) GetDistrictBySlug(ctx context.Context, slug string) (*models.CongressionalDistrict, error) {
	query := `
		SELECT d.id, d.province_id, d.city_municipality_id, d.district_number, d.name, d.slug,
			d.created_at, d.updated_at, d.deleted_at
		FROM congressional_districts d
		WHERE d.slug = $1 AND d.deleted_at IS NULL
	`

	district := &models.CongressionalDistrict{}
	err := r.db.QueryRow(ctx, query, slug).Scan(
		&district.ID, &district.ProvinceID, &district.CityMunicipalityID,
		&district.DistrictNumber, &district.Name, &district.Slug,
		&district.CreatedAt, &district.UpdatedAt, &district.DeletedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get district by slug: %w", err)
	}

	return district, nil
}

func (r *LocationRepository) ListDistrictsByProvince(ctx context.Context, provinceID uuid.UUID) ([]models.DistrictListItem, error) {
	query := `
		SELECT d.id, d.district_number, d.name, d.slug, p.name as province_name
		FROM congressional_districts d
		LEFT JOIN provinces p ON d.province_id = p.id
		WHERE d.province_id = $1 AND d.deleted_at IS NULL
		ORDER BY d.district_number ASC
	`

	rows, err := r.db.Query(ctx, query, provinceID)
	if err != nil {
		return nil, fmt.Errorf("failed to list districts: %w", err)
	}
	defer rows.Close()

	districts := []models.DistrictListItem{}
	for rows.Next() {
		var district models.DistrictListItem
		err := rows.Scan(&district.ID, &district.DistrictNumber, &district.Name, &district.Slug, &district.ProvinceName)
		if err != nil {
			return nil, fmt.Errorf("failed to scan district: %w", err)
		}
		districts = append(districts, district)
	}

	return districts, nil
}

// =====================================================
// SEARCH & HIERARCHY
// =====================================================

func (r *LocationRepository) SearchLocations(ctx context.Context, query string, limit int) ([]models.LocationSearchResult, error) {
	if limit <= 0 {
		limit = 20
	}

	searchPattern := "%" + strings.ToLower(query) + "%"

	sqlQuery := `
		(SELECT 'region' as type, id, code, name, slug, '' as parent_name, name as full_path
		 FROM regions WHERE LOWER(name) LIKE $1 AND deleted_at IS NULL LIMIT $2)
		UNION ALL
		(SELECT 'province' as type, p.id, p.code, p.name, p.slug, r.name as parent_name,
		        p.name || ', ' || r.name as full_path
		 FROM provinces p
		 LEFT JOIN regions r ON p.region_id = r.id
		 WHERE LOWER(p.name) LIKE $1 AND p.deleted_at IS NULL LIMIT $2)
		UNION ALL
		(SELECT 'city' as type, c.id, c.code, c.name, c.slug, p.name as parent_name,
		        c.name || ', ' || p.name as full_path
		 FROM cities_municipalities c
		 LEFT JOIN provinces p ON c.province_id = p.id
		 WHERE LOWER(c.name) LIKE $1 AND c.deleted_at IS NULL LIMIT $2)
		UNION ALL
		(SELECT 'barangay' as type, b.id, b.code, b.name, b.slug, c.name as parent_name,
		        b.name || ', ' || c.name as full_path
		 FROM barangays b
		 LEFT JOIN cities_municipalities c ON b.city_municipality_id = c.id
		 WHERE LOWER(b.name) LIKE $1 AND b.deleted_at IS NULL LIMIT $2)
		ORDER BY type, name
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, sqlQuery, searchPattern, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search locations: %w", err)
	}
	defer rows.Close()

	results := []models.LocationSearchResult{}
	for rows.Next() {
		var result models.LocationSearchResult
		err := rows.Scan(&result.Type, &result.ID, &result.Code, &result.Name, &result.Slug, &result.ParentName, &result.FullPath)
		if err != nil {
			return nil, fmt.Errorf("failed to scan search result: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

// GetLocationHierarchy returns the full hierarchy for a given barangay
func (r *LocationRepository) GetLocationHierarchy(ctx context.Context, barangayID uuid.UUID) (*models.LocationHierarchy, error) {
	query := `
		SELECT
			r.id, r.code, r.name, r.slug,
			p.id, p.region_id, p.code, p.name, p.slug,
			c.id, c.province_id, c.code, c.name, c.slug, c.is_city, c.is_capital, c.is_huc,
			b.id, b.city_municipality_id, b.code, b.name, b.slug
		FROM barangays b
		JOIN cities_municipalities c ON b.city_municipality_id = c.id
		JOIN provinces p ON c.province_id = p.id
		JOIN regions r ON p.region_id = r.id
		WHERE b.id = $1 AND b.deleted_at IS NULL
	`

	hierarchy := &models.LocationHierarchy{
		Region:           &models.RegionListItem{},
		Province:         &models.ProvinceListItem{},
		CityMunicipality: &models.CityMunicipalityListItem{},
		Barangay:         &models.BarangayListItem{},
	}

	err := r.db.QueryRow(ctx, query, barangayID).Scan(
		&hierarchy.Region.ID, &hierarchy.Region.Code, &hierarchy.Region.Name, &hierarchy.Region.Slug,
		&hierarchy.Province.ID, &hierarchy.Province.RegionID, &hierarchy.Province.Code, &hierarchy.Province.Name, &hierarchy.Province.Slug,
		&hierarchy.CityMunicipality.ID, &hierarchy.CityMunicipality.ProvinceID, &hierarchy.CityMunicipality.Code,
		&hierarchy.CityMunicipality.Name, &hierarchy.CityMunicipality.Slug, &hierarchy.CityMunicipality.IsCity,
		&hierarchy.CityMunicipality.IsCapital, &hierarchy.CityMunicipality.IsHUC,
		&hierarchy.Barangay.ID, &hierarchy.Barangay.CityMunicipalityID, &hierarchy.Barangay.Code, &hierarchy.Barangay.Name, &hierarchy.Barangay.Slug,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get location hierarchy: %w", err)
	}

	return hierarchy, nil
}

// GetRegionByCode gets a region by its PSGC code
func (r *LocationRepository) GetRegionByCode(ctx context.Context, code string) (*models.Region, error) {
	query := `
		SELECT id, code, name, slug, created_at, updated_at, deleted_at
		FROM regions
		WHERE code = $1 AND deleted_at IS NULL
	`

	region := &models.Region{}
	err := r.db.QueryRow(ctx, query, code).Scan(
		&region.ID, &region.Code, &region.Name, &region.Slug,
		&region.CreatedAt, &region.UpdatedAt, &region.DeletedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get region by code: %w", err)
	}

	return region, nil
}

// GetProvinceByCode gets a province by its PSGC code
func (r *LocationRepository) GetProvinceByCode(ctx context.Context, code string) (*models.Province, error) {
	query := `
		SELECT id, region_id, code, name, slug, created_at, updated_at, deleted_at
		FROM provinces
		WHERE code = $1 AND deleted_at IS NULL
	`

	province := &models.Province{}
	err := r.db.QueryRow(ctx, query, code).Scan(
		&province.ID, &province.RegionID, &province.Code, &province.Name, &province.Slug,
		&province.CreatedAt, &province.UpdatedAt, &province.DeletedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get province by code: %w", err)
	}

	return province, nil
}

// GetCityMunicipalityByCode gets a city/municipality by its PSGC code
func (r *LocationRepository) GetCityMunicipalityByCode(ctx context.Context, code string) (*models.CityMunicipality, error) {
	query := `
		SELECT id, province_id, code, name, slug, is_city, is_capital, is_huc, is_icc, population,
			created_at, updated_at, deleted_at
		FROM cities_municipalities
		WHERE code = $1 AND deleted_at IS NULL
	`

	city := &models.CityMunicipality{}
	err := r.db.QueryRow(ctx, query, code).Scan(
		&city.ID, &city.ProvinceID, &city.Code, &city.Name, &city.Slug,
		&city.IsCity, &city.IsCapital, &city.IsHUC, &city.IsICC, &city.Population,
		&city.CreatedAt, &city.UpdatedAt, &city.DeletedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get city by code: %w", err)
	}

	return city, nil
}

// GetBarangayByCode gets a barangay by its PSGC code
func (r *LocationRepository) GetBarangayByCode(ctx context.Context, code string) (*models.Barangay, error) {
	query := `
		SELECT id, city_municipality_id, code, name, slug, population, created_at, updated_at, deleted_at
		FROM barangays
		WHERE code = $1 AND deleted_at IS NULL
	`

	barangay := &models.Barangay{}
	err := r.db.QueryRow(ctx, query, code).Scan(
		&barangay.ID, &barangay.CityMunicipalityID, &barangay.Code, &barangay.Name, &barangay.Slug,
		&barangay.Population, &barangay.CreatedAt, &barangay.UpdatedAt, &barangay.DeletedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get barangay by code: %w", err)
	}

	return barangay, nil
}
