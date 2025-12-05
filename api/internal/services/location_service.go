package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/humfurie/pulpulitiko/api/internal/models"
	"github.com/humfurie/pulpulitiko/api/internal/repository"
	"github.com/humfurie/pulpulitiko/api/pkg/cache"
)

type LocationService struct {
	repo  *repository.LocationRepository
	cache *cache.RedisCache
}

func NewLocationService(repo *repository.LocationRepository, cache *cache.RedisCache) *LocationService {
	return &LocationService{
		repo:  repo,
		cache: cache,
	}
}

// =====================================================
// REGIONS
// =====================================================

func (s *LocationService) CreateRegion(ctx context.Context, req *models.CreateRegionRequest) (*models.Region, error) {
	region := &models.Region{
		Code: req.Code,
		Name: req.Name,
		Slug: req.Slug,
	}

	if err := s.repo.CreateRegion(ctx, region); err != nil {
		return nil, err
	}

	s.invalidateRegionsCache(ctx)
	return region, nil
}

func (s *LocationService) GetRegionByID(ctx context.Context, id uuid.UUID) (*models.Region, error) {
	cacheKey := cache.RegionKey(id.String())
	var region models.Region
	if err := s.cache.Get(ctx, cacheKey, &region); err == nil {
		return &region, nil
	}

	result, err := s.repo.GetRegionByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}

	_ = s.cache.Set(ctx, cacheKey, result, time.Hour)
	return result, nil
}

func (s *LocationService) GetRegionBySlug(ctx context.Context, slug string) (*models.Region, error) {
	cacheKey := cache.RegionSlugKey(slug)
	var region models.Region
	if err := s.cache.Get(ctx, cacheKey, &region); err == nil {
		return &region, nil
	}

	result, err := s.repo.GetRegionBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}

	_ = s.cache.Set(ctx, cacheKey, result, time.Hour)
	return result, nil
}

func (s *LocationService) ListRegions(ctx context.Context) ([]models.RegionListItem, error) {
	cacheKey := cache.RegionsKey()
	var regions []models.RegionListItem
	if err := s.cache.Get(ctx, cacheKey, &regions); err == nil {
		return regions, nil
	}

	result, err := s.repo.ListRegions(ctx)
	if err != nil {
		return nil, err
	}

	// Cache for 24 hours (regions rarely change)
	_ = s.cache.Set(ctx, cacheKey, result, 24*time.Hour)
	return result, nil
}

func (s *LocationService) UpdateRegion(ctx context.Context, id uuid.UUID, req *models.UpdateRegionRequest) (*models.Region, error) {
	if err := s.repo.UpdateRegion(ctx, id, req); err != nil {
		return nil, err
	}

	s.invalidateRegionsCache(ctx)
	_ = s.cache.Delete(ctx, cache.RegionKey(id.String()))

	return s.repo.GetRegionByID(ctx, id)
}

func (s *LocationService) DeleteRegion(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.DeleteRegion(ctx, id); err != nil {
		return err
	}

	s.invalidateRegionsCache(ctx)
	_ = s.cache.Delete(ctx, cache.RegionKey(id.String()))
	return nil
}

// =====================================================
// PROVINCES
// =====================================================

func (s *LocationService) CreateProvince(ctx context.Context, req *models.CreateProvinceRequest) (*models.Province, error) {
	regionID, err := uuid.Parse(req.RegionID)
	if err != nil {
		return nil, err
	}

	province := &models.Province{
		RegionID: regionID,
		Code:     req.Code,
		Name:     req.Name,
		Slug:     req.Slug,
	}

	if err := s.repo.CreateProvince(ctx, province); err != nil {
		return nil, err
	}

	s.invalidateProvincesCache(ctx, regionID)
	return province, nil
}

func (s *LocationService) GetProvinceByID(ctx context.Context, id uuid.UUID) (*models.Province, error) {
	cacheKey := cache.ProvinceKey(id.String())
	var province models.Province
	if err := s.cache.Get(ctx, cacheKey, &province); err == nil {
		return &province, nil
	}

	result, err := s.repo.GetProvinceByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}

	_ = s.cache.Set(ctx, cacheKey, result, time.Hour)
	return result, nil
}

func (s *LocationService) GetProvinceBySlug(ctx context.Context, slug string) (*models.Province, error) {
	cacheKey := cache.ProvinceSlugKey(slug)
	var province models.Province
	if err := s.cache.Get(ctx, cacheKey, &province); err == nil {
		return &province, nil
	}

	result, err := s.repo.GetProvinceBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}

	_ = s.cache.Set(ctx, cacheKey, result, time.Hour)
	return result, nil
}

func (s *LocationService) ListProvincesByRegion(ctx context.Context, regionID uuid.UUID) ([]models.ProvinceListItem, error) {
	cacheKey := cache.ProvincesKey(regionID.String())
	var provinces []models.ProvinceListItem
	if err := s.cache.Get(ctx, cacheKey, &provinces); err == nil {
		return provinces, nil
	}

	result, err := s.repo.ListProvincesByRegion(ctx, regionID)
	if err != nil {
		return nil, err
	}

	_ = s.cache.Set(ctx, cacheKey, result, 24*time.Hour)
	return result, nil
}

func (s *LocationService) ListAllProvinces(ctx context.Context) ([]models.ProvinceListItem, error) {
	cacheKey := cache.AllProvincesKey()
	var provinces []models.ProvinceListItem
	if err := s.cache.Get(ctx, cacheKey, &provinces); err == nil {
		return provinces, nil
	}

	result, err := s.repo.ListAllProvinces(ctx)
	if err != nil {
		return nil, err
	}

	_ = s.cache.Set(ctx, cacheKey, result, 24*time.Hour)
	return result, nil
}

func (s *LocationService) UpdateProvince(ctx context.Context, id uuid.UUID, req *models.UpdateProvinceRequest) (*models.Province, error) {
	// Get current province to know which region's cache to invalidate
	current, _ := s.repo.GetProvinceByID(ctx, id)

	if err := s.repo.UpdateProvince(ctx, id, req); err != nil {
		return nil, err
	}

	if current != nil {
		s.invalidateProvincesCache(ctx, current.RegionID)
	}
	_ = s.cache.Delete(ctx, cache.ProvinceKey(id.String()))

	return s.repo.GetProvinceByID(ctx, id)
}

func (s *LocationService) DeleteProvince(ctx context.Context, id uuid.UUID) error {
	current, _ := s.repo.GetProvinceByID(ctx, id)

	if err := s.repo.DeleteProvince(ctx, id); err != nil {
		return err
	}

	if current != nil {
		s.invalidateProvincesCache(ctx, current.RegionID)
	}
	_ = s.cache.Delete(ctx, cache.ProvinceKey(id.String()))
	return nil
}

// =====================================================
// CITIES/MUNICIPALITIES
// =====================================================

func (s *LocationService) CreateCityMunicipality(ctx context.Context, req *models.CreateCityMunicipalityRequest) (*models.CityMunicipality, error) {
	provinceID, err := uuid.Parse(req.ProvinceID)
	if err != nil {
		return nil, err
	}

	city := &models.CityMunicipality{
		ProvinceID: provinceID,
		Code:       req.Code,
		Name:       req.Name,
		Slug:       req.Slug,
		IsCity:     req.IsCity,
		IsCapital:  req.IsCapital,
		IsHUC:      req.IsHUC,
		IsICC:      req.IsICC,
		Population: req.Population,
	}

	if err := s.repo.CreateCityMunicipality(ctx, city); err != nil {
		return nil, err
	}

	s.invalidateCitiesCache(ctx, provinceID)
	return city, nil
}

func (s *LocationService) GetCityMunicipalityByID(ctx context.Context, id uuid.UUID) (*models.CityMunicipality, error) {
	cacheKey := cache.CityKey(id.String())
	var city models.CityMunicipality
	if err := s.cache.Get(ctx, cacheKey, &city); err == nil {
		return &city, nil
	}

	result, err := s.repo.GetCityMunicipalityByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}

	_ = s.cache.Set(ctx, cacheKey, result, time.Hour)
	return result, nil
}

func (s *LocationService) GetCityMunicipalityBySlug(ctx context.Context, slug string) (*models.CityMunicipality, error) {
	cacheKey := cache.CitySlugKey(slug)
	var city models.CityMunicipality
	if err := s.cache.Get(ctx, cacheKey, &city); err == nil {
		return &city, nil
	}

	result, err := s.repo.GetCityMunicipalityBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}

	_ = s.cache.Set(ctx, cacheKey, result, time.Hour)
	return result, nil
}

func (s *LocationService) ListCitiesByProvince(ctx context.Context, provinceID uuid.UUID) ([]models.CityMunicipalityListItem, error) {
	cacheKey := cache.CitiesKey(provinceID.String())
	var cities []models.CityMunicipalityListItem
	if err := s.cache.Get(ctx, cacheKey, &cities); err == nil {
		return cities, nil
	}

	result, err := s.repo.ListCitiesByProvince(ctx, provinceID)
	if err != nil {
		return nil, err
	}

	_ = s.cache.Set(ctx, cacheKey, result, 24*time.Hour)
	return result, nil
}

func (s *LocationService) UpdateCityMunicipality(ctx context.Context, id uuid.UUID, req *models.UpdateCityMunicipalityRequest) (*models.CityMunicipality, error) {
	current, _ := s.repo.GetCityMunicipalityByID(ctx, id)

	if err := s.repo.UpdateCityMunicipality(ctx, id, req); err != nil {
		return nil, err
	}

	if current != nil {
		s.invalidateCitiesCache(ctx, current.ProvinceID)
	}
	_ = s.cache.Delete(ctx, cache.CityKey(id.String()))

	return s.repo.GetCityMunicipalityByID(ctx, id)
}

func (s *LocationService) DeleteCityMunicipality(ctx context.Context, id uuid.UUID) error {
	current, _ := s.repo.GetCityMunicipalityByID(ctx, id)

	if err := s.repo.DeleteCityMunicipality(ctx, id); err != nil {
		return err
	}

	if current != nil {
		s.invalidateCitiesCache(ctx, current.ProvinceID)
	}
	_ = s.cache.Delete(ctx, cache.CityKey(id.String()))
	return nil
}

// =====================================================
// BARANGAYS
// =====================================================

func (s *LocationService) CreateBarangay(ctx context.Context, req *models.CreateBarangayRequest) (*models.Barangay, error) {
	cityID, err := uuid.Parse(req.CityMunicipalityID)
	if err != nil {
		return nil, err
	}

	barangay := &models.Barangay{
		CityMunicipalityID: cityID,
		Code:               req.Code,
		Name:               req.Name,
		Slug:               req.Slug,
		Population:         req.Population,
	}

	if err := s.repo.CreateBarangay(ctx, barangay); err != nil {
		return nil, err
	}

	s.invalidateBarangaysCache(ctx, cityID)
	return barangay, nil
}

func (s *LocationService) GetBarangayByID(ctx context.Context, id uuid.UUID) (*models.Barangay, error) {
	cacheKey := cache.BarangayKey(id.String())
	var barangay models.Barangay
	if err := s.cache.Get(ctx, cacheKey, &barangay); err == nil {
		return &barangay, nil
	}

	result, err := s.repo.GetBarangayByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}

	_ = s.cache.Set(ctx, cacheKey, result, time.Hour)
	return result, nil
}

func (s *LocationService) GetBarangayBySlug(ctx context.Context, slug string) (*models.Barangay, error) {
	cacheKey := cache.BarangaySlugKey(slug)
	var barangay models.Barangay
	if err := s.cache.Get(ctx, cacheKey, &barangay); err == nil {
		return &barangay, nil
	}

	result, err := s.repo.GetBarangayBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}

	_ = s.cache.Set(ctx, cacheKey, result, time.Hour)
	return result, nil
}

func (s *LocationService) ListBarangaysByCity(ctx context.Context, cityID uuid.UUID, page, perPage int) (*models.PaginatedBarangays, error) {
	// Don't cache paginated results
	return s.repo.ListBarangaysByCity(ctx, cityID, page, perPage)
}

func (s *LocationService) UpdateBarangay(ctx context.Context, id uuid.UUID, req *models.UpdateBarangayRequest) (*models.Barangay, error) {
	current, _ := s.repo.GetBarangayByID(ctx, id)

	if err := s.repo.UpdateBarangay(ctx, id, req); err != nil {
		return nil, err
	}

	if current != nil {
		s.invalidateBarangaysCache(ctx, current.CityMunicipalityID)
	}
	_ = s.cache.Delete(ctx, cache.BarangayKey(id.String()))

	return s.repo.GetBarangayByID(ctx, id)
}

func (s *LocationService) DeleteBarangay(ctx context.Context, id uuid.UUID) error {
	current, _ := s.repo.GetBarangayByID(ctx, id)

	if err := s.repo.DeleteBarangay(ctx, id); err != nil {
		return err
	}

	if current != nil {
		s.invalidateBarangaysCache(ctx, current.CityMunicipalityID)
	}
	_ = s.cache.Delete(ctx, cache.BarangayKey(id.String()))
	return nil
}

// =====================================================
// CONGRESSIONAL DISTRICTS
// =====================================================

func (s *LocationService) CreateDistrict(ctx context.Context, req *models.CreateDistrictRequest) (*models.CongressionalDistrict, error) {
	district := &models.CongressionalDistrict{
		DistrictNumber: req.DistrictNumber,
		Name:           req.Name,
		Slug:           req.Slug,
	}

	if req.ProvinceID != nil {
		provinceID, err := uuid.Parse(*req.ProvinceID)
		if err != nil {
			return nil, err
		}
		district.ProvinceID = &provinceID
	}

	if req.CityMunicipalityID != nil {
		cityID, err := uuid.Parse(*req.CityMunicipalityID)
		if err != nil {
			return nil, err
		}
		district.CityMunicipalityID = &cityID
	}

	if err := s.repo.CreateDistrict(ctx, district); err != nil {
		return nil, err
	}

	return district, nil
}

func (s *LocationService) GetDistrictByID(ctx context.Context, id uuid.UUID) (*models.CongressionalDistrict, error) {
	return s.repo.GetDistrictByID(ctx, id)
}

func (s *LocationService) GetDistrictBySlug(ctx context.Context, slug string) (*models.CongressionalDistrict, error) {
	return s.repo.GetDistrictBySlug(ctx, slug)
}

func (s *LocationService) ListDistrictsByProvince(ctx context.Context, provinceID uuid.UUID) ([]models.DistrictListItem, error) {
	return s.repo.ListDistrictsByProvince(ctx, provinceID)
}

// =====================================================
// SEARCH & HIERARCHY
// =====================================================

func (s *LocationService) SearchLocations(ctx context.Context, query string, limit int) ([]models.LocationSearchResult, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.repo.SearchLocations(ctx, query, limit)
}

func (s *LocationService) GetLocationHierarchy(ctx context.Context, barangayID uuid.UUID) (*models.LocationHierarchy, error) {
	cacheKey := cache.LocationHierarchyKey(barangayID.String())
	var hierarchy models.LocationHierarchy
	if err := s.cache.Get(ctx, cacheKey, &hierarchy); err == nil {
		return &hierarchy, nil
	}

	result, err := s.repo.GetLocationHierarchy(ctx, barangayID)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}

	_ = s.cache.Set(ctx, cacheKey, result, 24*time.Hour)
	return result, nil
}

// =====================================================
// CACHE INVALIDATION
// =====================================================

func (s *LocationService) invalidateRegionsCache(ctx context.Context) {
	_ = s.cache.Delete(ctx, cache.RegionsKey())
	_ = s.cache.DeletePattern(ctx, cache.KeyPrefixRegion+"*")
}

func (s *LocationService) invalidateProvincesCache(ctx context.Context, regionID uuid.UUID) {
	_ = s.cache.Delete(ctx, cache.ProvincesKey(regionID.String()))
	_ = s.cache.Delete(ctx, cache.AllProvincesKey())
	_ = s.cache.DeletePattern(ctx, cache.KeyPrefixProvince+"*")
}

func (s *LocationService) invalidateCitiesCache(ctx context.Context, provinceID uuid.UUID) {
	_ = s.cache.Delete(ctx, cache.CitiesKey(provinceID.String()))
	_ = s.cache.DeletePattern(ctx, cache.KeyPrefixCity+"*")
}

func (s *LocationService) invalidateBarangaysCache(ctx context.Context, cityID uuid.UUID) {
	_ = s.cache.Delete(ctx, cache.BarangaysKey(cityID.String()))
	_ = s.cache.DeletePattern(ctx, cache.KeyPrefixBarangay+"*")
}
