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

type ElectionEventService struct {
	repo        *repository.ElectionEventRepository
	historyRepo *repository.PositionHistoryRepository
	cache       *cache.RedisCache
}

func NewElectionEventService(
	repo *repository.ElectionEventRepository,
	historyRepo *repository.PositionHistoryRepository,
	cache *cache.RedisCache,
) *ElectionEventService {
	return &ElectionEventService{
		repo:        repo,
		historyRepo: historyRepo,
		cache:       cache,
	}
}

// Cache TTL
const electionTTL = 24 * time.Hour

// Create creates a new election event
func (s *ElectionEventService) Create(ctx context.Context, req *models.CreateElectionEventRequest, createdBy uuid.UUID) (*models.ElectionEvent, error) {
	election, err := s.repo.Create(ctx, req, createdBy)
	if err != nil {
		return nil, err
	}

	// Invalidate caches
	_ = s.cache.DeletePattern(ctx, "election:*")
	_ = s.cache.DeletePattern(ctx, "elections:*")

	return election, nil
}

// GetByID retrieves an election event by ID
func (s *ElectionEventService) GetByID(ctx context.Context, id uuid.UUID) (*models.ElectionEvent, error) {
	cacheKey := fmt.Sprintf("election:id:%s", id.String())

	var election models.ElectionEvent
	if err := s.cache.Get(ctx, cacheKey, &election); err == nil {
		return &election, nil
	}

	result, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if result != nil {
		// Cache the result
		_ = s.cache.Set(ctx, cacheKey, result, electionTTL)
	}

	return result, nil
}

// List retrieves election events with pagination
func (s *ElectionEventService) List(ctx context.Context, page, perPage int, status *string) (*models.PaginatedElectionEvents, error) {
	cacheKey := fmt.Sprintf("elections:page:%d:perPage:%d:status:%v", page, perPage, status)

	var elections models.PaginatedElectionEvents
	if err := s.cache.Get(ctx, cacheKey, &elections); err == nil {
		return &elections, nil
	}

	result, err := s.repo.List(ctx, page, perPage, status)
	if err != nil {
		return nil, err
	}

	// Cache the result
	_ = s.cache.Set(ctx, cacheKey, result, electionTTL)

	return result, nil
}

// Update updates an election event
func (s *ElectionEventService) Update(ctx context.Context, id uuid.UUID, req *models.UpdateElectionEventRequest) error {
	if err := s.repo.Update(ctx, id, req); err != nil {
		return err
	}

	// Invalidate caches
	_ = s.cache.Delete(ctx, fmt.Sprintf("election:id:%s", id.String()))
	_ = s.cache.DeletePattern(ctx, "elections:*")

	return nil
}

// Delete deletes an election event
func (s *ElectionEventService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	// Invalidate caches
	_ = s.cache.Delete(ctx, fmt.Sprintf("election:id:%s", id.String()))
	_ = s.cache.DeletePattern(ctx, "elections:*")

	return nil
}

// ArchiveCurrentHolders archives all current holders of specified positions for an election
// This is called before importing new election results
func (s *ElectionEventService) ArchiveCurrentHolders(ctx context.Context, electionID uuid.UUID, positionIDs []uuid.UUID) (int, error) {
	// Get election details
	election, err := s.repo.GetByID(ctx, electionID)
	if err != nil {
		return 0, fmt.Errorf("failed to get election: %w", err)
	}
	if election == nil {
		return 0, fmt.Errorf("election not found")
	}

	// Archive all current holders
	if err := s.historyRepo.BulkArchiveForElection(ctx, electionID, positionIDs, election.ElectionDate.Format("2006-01-02")); err != nil {
		return 0, fmt.Errorf("failed to bulk archive positions: %w", err)
	}

	// Update election status to in_progress
	if err := s.repo.UpdateStatus(ctx, electionID, "in_progress"); err != nil {
		return 0, fmt.Errorf("failed to update election status: %w", err)
	}

	// Invalidate caches
	_ = s.cache.DeletePattern(ctx, "position_holder:*")
	_ = s.cache.DeletePattern(ctx, "position_history:*")

	return len(positionIDs), nil
}

// CompleteElection marks an election as completed
func (s *ElectionEventService) CompleteElection(ctx context.Context, electionID uuid.UUID) error {
	if err := s.repo.UpdateStatus(ctx, electionID, "completed"); err != nil {
		return err
	}

	// Invalidate caches
	_ = s.cache.Delete(ctx, fmt.Sprintf("election:id:%s", electionID.String()))
	_ = s.cache.DeletePattern(ctx, "elections:*")

	return nil
}

// FailElection marks an election as failed (e.g., import errors)
func (s *ElectionEventService) FailElection(ctx context.Context, electionID uuid.UUID) error {
	if err := s.repo.UpdateStatus(ctx, electionID, "failed"); err != nil {
		return err
	}

	// Invalidate caches
	_ = s.cache.Delete(ctx, fmt.Sprintf("election:id:%s", electionID.String()))
	_ = s.cache.DeletePattern(ctx, "elections:*")

	return nil
}

// GetStatistics retrieves detailed statistics for an election
func (s *ElectionEventService) GetStatistics(ctx context.Context, electionID uuid.UUID) (*models.ElectionStatistics, error) {
	cacheKey := fmt.Sprintf("election:stats:%s", electionID.String())

	var stats models.ElectionStatistics
	if err := s.cache.Get(ctx, cacheKey, &stats); err == nil {
		return &stats, nil
	}

	result, err := s.repo.GetStatistics(ctx, electionID)
	if err != nil {
		return nil, err
	}

	// Cache the result with shorter TTL (stats change during imports)
	_ = s.cache.Set(ctx, cacheKey, result, 5*time.Minute)

	return result, nil
}
