package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/humfurie/pulpulitiko/api/internal/models"
	"github.com/humfurie/pulpulitiko/api/internal/repository"
	"github.com/humfurie/pulpulitiko/api/pkg/cache"
)

type PoliticalPartyService struct {
	repo  *repository.PoliticalPartyRepository
	cache *cache.RedisCache
}

func NewPoliticalPartyService(repo *repository.PoliticalPartyRepository, cache *cache.RedisCache) *PoliticalPartyService {
	return &PoliticalPartyService{repo: repo, cache: cache}
}

// Cache TTL
const partyTTL = 24 * time.Hour

// Political Party methods

func (s *PoliticalPartyService) Create(ctx context.Context, req *models.CreatePoliticalPartyRequest) (*models.PoliticalParty, error) {
	party, err := s.repo.Create(ctx, req)
	if err != nil {
		return nil, err
	}

	// Invalidate relevant caches
	s.cache.DeletePattern(ctx, "party:*")
	s.cache.DeletePattern(ctx, "parties:*")

	return party, nil
}

func (s *PoliticalPartyService) GetByID(ctx context.Context, id uuid.UUID) (*models.PoliticalParty, error) {
	cacheKey := "party:id:" + id.String()

	var party models.PoliticalParty
	if err := s.cache.Get(ctx, cacheKey, &party); err == nil {
		return &party, nil
	}

	result, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}

	s.cache.Set(ctx, cacheKey, result, partyTTL)
	return result, nil
}

func (s *PoliticalPartyService) GetBySlug(ctx context.Context, slug string) (*models.PoliticalParty, error) {
	cacheKey := "party:slug:" + slug

	var party models.PoliticalParty
	if err := s.cache.Get(ctx, cacheKey, &party); err == nil {
		return &party, nil
	}

	result, err := s.repo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}

	s.cache.Set(ctx, cacheKey, result, partyTTL)
	return result, nil
}

func (s *PoliticalPartyService) List(ctx context.Context, page, perPage int, majorOnly, activeOnly bool) (*models.PaginatedPoliticalParties, error) {
	return s.repo.List(ctx, page, perPage, majorOnly, activeOnly)
}

func (s *PoliticalPartyService) GetAll(ctx context.Context, activeOnly bool) ([]models.PoliticalPartyListItem, error) {
	cacheKey := "parties:all"
	if activeOnly {
		cacheKey += ":active"
	}

	var parties []models.PoliticalPartyListItem
	if err := s.cache.Get(ctx, cacheKey, &parties); err == nil {
		return parties, nil
	}

	result, err := s.repo.GetAll(ctx, activeOnly)
	if err != nil {
		return nil, err
	}

	s.cache.Set(ctx, cacheKey, result, partyTTL)
	return result, nil
}

func (s *PoliticalPartyService) Update(ctx context.Context, id uuid.UUID, req *models.UpdatePoliticalPartyRequest) (*models.PoliticalParty, error) {
	party, err := s.repo.Update(ctx, id, req)
	if err != nil {
		return nil, err
	}

	// Invalidate relevant caches
	s.cache.DeletePattern(ctx, "party:*")
	s.cache.DeletePattern(ctx, "parties:*")

	return party, nil
}

func (s *PoliticalPartyService) Delete(ctx context.Context, id uuid.UUID) error {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		return err
	}

	// Invalidate relevant caches
	s.cache.DeletePattern(ctx, "party:*")
	s.cache.DeletePattern(ctx, "parties:*")

	return nil
}

// Government Position methods

func (s *PoliticalPartyService) GetAllPositions(ctx context.Context) ([]models.GovernmentPositionListItem, error) {
	cacheKey := "positions:all"

	var positions []models.GovernmentPositionListItem
	if err := s.cache.Get(ctx, cacheKey, &positions); err == nil {
		return positions, nil
	}

	result, err := s.repo.GetAllPositions(ctx)
	if err != nil {
		return nil, err
	}

	s.cache.Set(ctx, cacheKey, result, partyTTL)
	return result, nil
}

func (s *PoliticalPartyService) GetPositionsByLevel(ctx context.Context, level string) ([]models.GovernmentPositionListItem, error) {
	cacheKey := "positions:level:" + level

	var positions []models.GovernmentPositionListItem
	if err := s.cache.Get(ctx, cacheKey, &positions); err == nil {
		return positions, nil
	}

	result, err := s.repo.GetPositionsByLevel(ctx, level)
	if err != nil {
		return nil, err
	}

	s.cache.Set(ctx, cacheKey, result, partyTTL)
	return result, nil
}

func (s *PoliticalPartyService) GetPositionByID(ctx context.Context, id uuid.UUID) (*models.GovernmentPosition, error) {
	return s.repo.GetPositionByID(ctx, id)
}

func (s *PoliticalPartyService) GetPositionBySlug(ctx context.Context, slug string) (*models.GovernmentPosition, error) {
	return s.repo.GetPositionBySlug(ctx, slug)
}

// Politician Jurisdiction methods

func (s *PoliticalPartyService) CreateJurisdiction(ctx context.Context, req *models.CreatePoliticianJurisdictionRequest) (*models.PoliticianJurisdiction, error) {
	return s.repo.CreateJurisdiction(ctx, req)
}

func (s *PoliticalPartyService) GetJurisdictionsByPolitician(ctx context.Context, politicianID uuid.UUID) ([]models.PoliticianJurisdiction, error) {
	return s.repo.GetJurisdictionsByPolitician(ctx, politicianID)
}

func (s *PoliticalPartyService) DeleteJurisdiction(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteJurisdiction(ctx, id)
}

func (s *PoliticalPartyService) DeleteAllJurisdictionsForPolitician(ctx context.Context, politicianID uuid.UUID) error {
	return s.repo.DeleteAllJurisdictionsForPolitician(ctx, politicianID)
}

// Find representatives by location
func (s *PoliticalPartyService) FindRepresentativesByBarangay(ctx context.Context, barangayID uuid.UUID) ([]models.PoliticianListItem, error) {
	return s.repo.FindRepresentativesByBarangay(ctx, barangayID)
}
