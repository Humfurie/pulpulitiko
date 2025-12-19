package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/humfurie/pulpulitiko/api/internal/models"
	"github.com/humfurie/pulpulitiko/api/internal/repository"
	"github.com/humfurie/pulpulitiko/api/pkg/cache"
)

type PositionHistoryService struct {
	repo           *repository.PositionHistoryRepository
	politicianRepo *repository.PoliticianRepository
	cache          *cache.RedisCache
}

func NewPositionHistoryService(
	repo *repository.PositionHistoryRepository,
	politicianRepo *repository.PoliticianRepository,
	cache *cache.RedisCache,
) *PositionHistoryService {
	return &PositionHistoryService{
		repo:           repo,
		politicianRepo: politicianRepo,
		cache:          cache,
	}
}

// Cache TTL
const positionHistoryTTL = 1 * time.Hour

// AssignPosition assigns a position to a politician with smart logic
// Smart Logic:
// 1. If politician already holds this exact position in this exact jurisdiction → update without history (if withHistory=false)
// 2. If different politician currently holds the position → end current holder's term, create history entry
// 3. If same person but withHistory=true → create new history record (for elections, term changes)
func (s *PositionHistoryService) AssignPosition(ctx context.Context, req *models.CreatePositionHistoryRequest, createdBy uuid.UUID) (*models.PoliticianPositionHistory, error) {
	// Check if there's a current holder for this position in this jurisdiction
	currentHolder, err := s.repo.GetCurrentHolder(ctx, &models.GetCurrentHolderRequest{
		PositionID: req.PositionID,
		RegionID:   req.RegionID,
		ProvinceID: req.ProvinceID,
		CityID:     req.CityID,
		BarangayID: req.BarangayID,
		DistrictID: req.DistrictID,
		IsNational: req.IsNational,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to check current holder: %w", err)
	}

	// Determine action based on current holder
	if currentHolder != nil {
		// Case 1: Same politician already holds this position
		if currentHolder.PoliticianID == req.PoliticianID {
			if !req.WithHistory {
				// Update without creating new history record
				return s.UpdatePositionWithoutHistory(ctx, currentHolder.ID, &models.UpdatePositionHistoryRequest{
					PartyID:   req.PartyID,
					TermStart: &req.TermStart,
					TermEnd:   req.TermEnd,
				})
			}
			// WithHistory=true: end current term and create new one (for elections, new terms)
			if err := s.repo.EndTermByID(ctx, currentHolder.ID, req.TermStart.Format("2006-01-02"), "term_change"); err != nil {
				return nil, fmt.Errorf("failed to end previous term: %w", err)
			}
		} else {
			// Case 2: Different politician holds the position → replace them
			if err := s.repo.EndTermByID(ctx, currentHolder.ID, req.TermStart.Format("2006-01-02"), "replaced"); err != nil {
				return nil, fmt.Errorf("failed to end current holder's term: %w", err)
			}
		}
	}

	// Create new position history entry
	history := &models.PoliticianPositionHistory{
		PoliticianID: req.PoliticianID,
		PositionID:   req.PositionID,
		PartyID:      req.PartyID,
		RegionID:     req.RegionID,
		ProvinceID:   req.ProvinceID,
		CityID:       req.CityID,
		BarangayID:   req.BarangayID,
		DistrictID:   req.DistrictID,
		IsNational:   req.IsNational,
		TermStart:    req.TermStart,
		TermEnd:      req.TermEnd,
		IsCurrent:    true,
		ElectionID:   req.ElectionID,
		CreatedBy:    &createdBy,
	}

	if err := s.repo.Create(ctx, history); err != nil {
		return nil, fmt.Errorf("failed to create position history: %w", err)
	}

	// Update politician's current position_id and party_id
	// Note: This would require adding a method to PoliticianRepository
	// For now, we'll skip this as the history table is the source of truth

	// Invalidate caches
	s.invalidateCaches(ctx, req.PoliticianID, req.PositionID)

	// Fetch the full history entry with joined data
	return s.repo.GetByID(ctx, history.ID)
}

// UpdatePositionWithHistory creates a new history record even if same politician
// Used for elections and term changes where we want to preserve the history
func (s *PositionHistoryService) UpdatePositionWithHistory(ctx context.Context, req *models.CreatePositionHistoryRequest, createdBy uuid.UUID) (*models.PoliticianPositionHistory, error) {
	req.WithHistory = true
	return s.AssignPosition(ctx, req, createdBy)
}

// UpdatePositionWithoutHistory updates the current history record in-place
// Used for minor updates like party changes, date corrections, etc.
func (s *PositionHistoryService) UpdatePositionWithoutHistory(ctx context.Context, historyID uuid.UUID, req *models.UpdatePositionHistoryRequest) (*models.PoliticianPositionHistory, error) {
	if err := s.repo.Update(ctx, historyID, req); err != nil {
		return nil, err
	}

	// Fetch updated history
	history, err := s.repo.GetByID(ctx, historyID)
	if err != nil {
		return nil, err
	}

	// Invalidate caches
	if history != nil {
		s.invalidateCaches(ctx, history.PoliticianID, history.PositionID)
	}

	return history, nil
}

// EndTerm ends a politician's current term
func (s *PositionHistoryService) EndTerm(ctx context.Context, politicianID uuid.UUID, req *models.EndTermRequest) error {
	if err := s.repo.EndTerm(ctx, politicianID, req.EndDate.Format("2006-01-02"), req.EndedReason); err != nil {
		return err
	}

	// Invalidate caches
	s.invalidateCaches(ctx, politicianID, uuid.Nil)

	return nil
}

// EndTermByID ends a specific position history entry
func (s *PositionHistoryService) EndTermByID(ctx context.Context, historyID uuid.UUID, req *models.EndTermRequest) error {
	// Get the history entry first to get politician and position IDs for cache invalidation
	history, err := s.repo.GetByID(ctx, historyID)
	if err != nil {
		return err
	}
	if history == nil {
		return fmt.Errorf("position history not found")
	}

	if err := s.repo.EndTermByID(ctx, historyID, req.EndDate.Format("2006-01-02"), req.EndedReason); err != nil {
		return err
	}

	// Invalidate caches
	s.invalidateCaches(ctx, history.PoliticianID, history.PositionID)

	return nil
}

// GetHistory retrieves a politician's position history timeline
func (s *PositionHistoryService) GetHistory(ctx context.Context, politicianID uuid.UUID) (*models.PoliticianPositionTimeline, error) {
	cacheKey := fmt.Sprintf("position_history:politician:%s", politicianID.String())

	var timeline models.PoliticianPositionTimeline
	if err := s.cache.Get(ctx, cacheKey, &timeline); err == nil {
		return &timeline, nil
	}

	// Get all history entries for politician
	history, err := s.repo.GetPoliticianHistory(ctx, politicianID)
	if err != nil {
		return nil, err
	}

	if len(history) == 0 {
		return nil, fmt.Errorf("no position history found for politician")
	}

	// Build timeline
	timeline = models.PoliticianPositionTimeline{
		PoliticianID:   politicianID,
		PoliticianName: history[0].PoliticianName,
		PoliticianSlug: history[0].PoliticianSlug,
		TotalPositions: len(history),
	}

	for _, h := range history {
		if h.IsCurrent {
			timeline.CurrentPosition = &h
		} else {
			timeline.PastPositions = append(timeline.PastPositions, h)
		}
	}

	// Cache the result
	_ = s.cache.Set(ctx, cacheKey, timeline, positionHistoryTTL)

	return &timeline, nil
}

// GetCurrentHolder finds who currently holds a position in a jurisdiction
func (s *PositionHistoryService) GetCurrentHolder(ctx context.Context, req *models.GetCurrentHolderRequest) (*models.PoliticianPositionHistory, error) {
	cacheKey := fmt.Sprintf("position_holder:position:%s:jurisdiction:%v",
		req.PositionID.String(),
		getJurisdictionCacheKey(req.RegionID, req.ProvinceID, req.CityID, req.BarangayID, req.DistrictID, req.IsNational))

	var holder models.PoliticianPositionHistory
	if err := s.cache.Get(ctx, cacheKey, &holder); err == nil {
		return &holder, nil
	}

	result, err := s.repo.GetCurrentHolder(ctx, req)
	if err != nil {
		return nil, err
	}

	if result != nil {
		// Cache the result
		_ = s.cache.Set(ctx, cacheKey, result, positionHistoryTTL)
	}

	return result, nil
}

// GetPositionHolders retrieves all politicians who held a specific position
func (s *PositionHistoryService) GetPositionHolders(ctx context.Context, positionID uuid.UUID) ([]models.PositionHistoryListItem, error) {
	cacheKey := fmt.Sprintf("position_holders:position:%s", positionID.String())

	var holders []models.PositionHistoryListItem
	if err := s.cache.Get(ctx, cacheKey, &holders); err == nil {
		return holders, nil
	}

	result, err := s.repo.GetPositionHolders(ctx, positionID)
	if err != nil {
		return nil, err
	}

	// Cache the result
	_ = s.cache.Set(ctx, cacheKey, result, positionHistoryTTL)

	return result, nil
}

// GetByID retrieves a position history entry by ID
func (s *PositionHistoryService) GetByID(ctx context.Context, id uuid.UUID) (*models.PoliticianPositionHistory, error) {
	cacheKey := fmt.Sprintf("position_history:id:%s", id.String())

	var history models.PoliticianPositionHistory
	if err := s.cache.Get(ctx, cacheKey, &history); err == nil {
		return &history, nil
	}

	result, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if result != nil {
		// Cache the result
		_ = s.cache.Set(ctx, cacheKey, result, positionHistoryTTL)
	}

	return result, nil
}

// Delete removes a position history entry
func (s *PositionHistoryService) Delete(ctx context.Context, id uuid.UUID) error {
	// Get the history entry first for cache invalidation
	history, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if history == nil {
		return fmt.Errorf("position history not found")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	// Invalidate caches
	s.invalidateCaches(ctx, history.PoliticianID, history.PositionID)

	return nil
}

// Helper functions

func (s *PositionHistoryService) invalidateCaches(ctx context.Context, politicianID, positionID uuid.UUID) {
	// Invalidate politician history cache
	_ = s.cache.Delete(ctx, fmt.Sprintf("position_history:politician:%s", politicianID.String()))

	// Invalidate position holders cache
	if positionID != uuid.Nil {
		_ = s.cache.DeletePattern(ctx, fmt.Sprintf("position_holder:position:%s:*", positionID.String()))
		_ = s.cache.Delete(ctx, fmt.Sprintf("position_holders:position:%s", positionID.String()))
	}

	// Invalidate politician cache (if needed for badge updates)
	_ = s.cache.DeletePattern(ctx, fmt.Sprintf("politician:*:%s", politicianID.String()))
}

func getJurisdictionCacheKey(regionID, provinceID, cityID, barangayID, districtID *uuid.UUID, isNational bool) string {
	if isNational {
		return "national"
	}
	if regionID != nil {
		return fmt.Sprintf("region:%s", regionID.String())
	}
	if provinceID != nil {
		return fmt.Sprintf("province:%s", provinceID.String())
	}
	if cityID != nil {
		return fmt.Sprintf("city:%s", cityID.String())
	}
	if barangayID != nil {
		return fmt.Sprintf("barangay:%s", barangayID.String())
	}
	if districtID != nil {
		return fmt.Sprintf("district:%s", districtID.String())
	}
	return "unknown"
}
