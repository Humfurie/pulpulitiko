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

const (
	electionCachePrefix       = "election:"
	electionsCachePrefix      = "elections:"
	candidatesCachePrefix     = "candidates:"
	voterEducationCachePrefix = "voter_ed:"
	electionCacheTTL          = 1 * time.Hour
	calendarCacheTTL          = 24 * time.Hour
)

type ElectionService struct {
	repo  *repository.ElectionRepository
	cache *cache.RedisCache
}

func NewElectionService(repo *repository.ElectionRepository, cache *cache.RedisCache) *ElectionService {
	return &ElectionService{
		repo:  repo,
		cache: cache,
	}
}

// Elections

func (s *ElectionService) CreateElection(ctx context.Context, req *models.CreateElectionRequest) (*models.Election, error) {
	election, err := s.repo.CreateElection(ctx, req)
	if err != nil {
		return nil, err
	}

	s.cache.DeletePattern(ctx, electionsCachePrefix+"*")

	return election, nil
}

func (s *ElectionService) GetElectionByID(ctx context.Context, id uuid.UUID) (*models.Election, error) {
	cacheKey := electionCachePrefix + "id:" + id.String()

	var election models.Election
	if err := s.cache.Get(ctx, cacheKey, &election); err == nil {
		return &election, nil
	}

	electionPtr, err := s.repo.GetElectionByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if electionPtr != nil {
		s.cache.Set(ctx, cacheKey, electionPtr, electionCacheTTL)
	}

	return electionPtr, nil
}

func (s *ElectionService) GetElectionBySlug(ctx context.Context, slug string) (*models.Election, error) {
	cacheKey := electionCachePrefix + "slug:" + slug

	var election models.Election
	if err := s.cache.Get(ctx, cacheKey, &election); err == nil {
		return &election, nil
	}

	electionPtr, err := s.repo.GetElectionBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	if electionPtr != nil {
		s.cache.Set(ctx, cacheKey, electionPtr, electionCacheTTL)
	}

	return electionPtr, nil
}

func (s *ElectionService) ListElections(ctx context.Context, filter *models.ElectionFilter, page, perPage int) (*models.PaginatedElections, error) {
	return s.repo.ListElections(ctx, filter, page, perPage)
}

func (s *ElectionService) GetUpcomingElections(ctx context.Context, limit int) ([]models.ElectionListItem, error) {
	cacheKey := fmt.Sprintf("%supcoming:%d", electionsCachePrefix, limit)

	var elections []models.ElectionListItem
	if err := s.cache.Get(ctx, cacheKey, &elections); err == nil {
		return elections, nil
	}

	elections, err := s.repo.GetUpcomingElections(ctx, limit)
	if err != nil {
		return nil, err
	}

	s.cache.Set(ctx, cacheKey, elections, electionCacheTTL)

	return elections, nil
}

func (s *ElectionService) GetFeaturedElections(ctx context.Context) ([]models.ElectionListItem, error) {
	cacheKey := electionsCachePrefix + "featured"

	var elections []models.ElectionListItem
	if err := s.cache.Get(ctx, cacheKey, &elections); err == nil {
		return elections, nil
	}

	elections, err := s.repo.GetFeaturedElections(ctx)
	if err != nil {
		return nil, err
	}

	s.cache.Set(ctx, cacheKey, elections, electionCacheTTL)

	return elections, nil
}

func (s *ElectionService) GetElectionCalendar(ctx context.Context, year int) ([]models.ElectionCalendarItem, error) {
	cacheKey := fmt.Sprintf("%scalendar:%d", electionsCachePrefix, year)

	var items []models.ElectionCalendarItem
	if err := s.cache.Get(ctx, cacheKey, &items); err == nil {
		return items, nil
	}

	items, err := s.repo.GetElectionCalendar(ctx, year)
	if err != nil {
		return nil, err
	}

	s.cache.Set(ctx, cacheKey, items, calendarCacheTTL)

	return items, nil
}

func (s *ElectionService) UpdateElection(ctx context.Context, id uuid.UUID, req *models.UpdateElectionRequest) (*models.Election, error) {
	election, err := s.repo.UpdateElection(ctx, id, req)
	if err != nil {
		return nil, err
	}

	if election != nil {
		s.invalidateElectionCache(ctx, id, election.Slug)
	}

	return election, nil
}

func (s *ElectionService) DeleteElection(ctx context.Context, id uuid.UUID) error {
	election, _ := s.repo.GetElectionByID(ctx, id)

	err := s.repo.DeleteElection(ctx, id)
	if err != nil {
		return err
	}

	if election != nil {
		s.invalidateElectionCache(ctx, id, election.Slug)
	}

	return nil
}

// Election Positions

func (s *ElectionService) CreateElectionPosition(ctx context.Context, req *models.CreateElectionPositionRequest) (*models.ElectionPosition, error) {
	position, err := s.repo.CreateElectionPosition(ctx, req)
	if err != nil {
		return nil, err
	}

	s.cache.Delete(ctx, electionCachePrefix+"id:"+req.ElectionID.String())

	return position, nil
}

func (s *ElectionService) GetElectionPositions(ctx context.Context, electionID uuid.UUID) ([]models.ElectionPositionListItem, error) {
	return s.repo.GetElectionPositions(ctx, electionID)
}

// Candidates

func (s *ElectionService) CreateCandidate(ctx context.Context, req *models.CreateCandidateRequest) (*models.Candidate, error) {
	candidate, err := s.repo.CreateCandidate(ctx, req)
	if err != nil {
		return nil, err
	}

	s.cache.DeletePattern(ctx, candidatesCachePrefix+"*")

	return candidate, nil
}

func (s *ElectionService) GetCandidateByID(ctx context.Context, id uuid.UUID) (*models.Candidate, error) {
	return s.repo.GetCandidateByID(ctx, id)
}

func (s *ElectionService) GetCandidatesForPosition(ctx context.Context, positionID uuid.UUID) ([]models.CandidateListItem, error) {
	cacheKey := candidatesCachePrefix + "position:" + positionID.String()

	var candidates []models.CandidateListItem
	if err := s.cache.Get(ctx, cacheKey, &candidates); err == nil {
		return candidates, nil
	}

	candidates, err := s.repo.GetCandidatesForPosition(ctx, positionID)
	if err != nil {
		return nil, err
	}

	s.cache.Set(ctx, cacheKey, candidates, electionCacheTTL)

	return candidates, nil
}

func (s *ElectionService) ListCandidates(ctx context.Context, filter *models.CandidateFilter, page, perPage int) (*models.PaginatedCandidates, error) {
	return s.repo.ListCandidates(ctx, filter, page, perPage)
}

func (s *ElectionService) UpdateCandidate(ctx context.Context, id uuid.UUID, req *models.UpdateCandidateRequest) (*models.Candidate, error) {
	candidate, err := s.repo.UpdateCandidate(ctx, id, req)
	if err != nil {
		return nil, err
	}

	s.cache.DeletePattern(ctx, candidatesCachePrefix+"*")

	return candidate, nil
}

// Voter Education

func (s *ElectionService) CreateVoterEducation(ctx context.Context, req *models.CreateVoterEducationRequest) (*models.VoterEducation, error) {
	ve, err := s.repo.CreateVoterEducation(ctx, req)
	if err != nil {
		return nil, err
	}

	s.cache.DeletePattern(ctx, voterEducationCachePrefix+"*")

	return ve, nil
}

func (s *ElectionService) GetVoterEducationBySlug(ctx context.Context, slug string) (*models.VoterEducation, error) {
	cacheKey := voterEducationCachePrefix + "slug:" + slug

	var ve models.VoterEducation
	if err := s.cache.Get(ctx, cacheKey, &ve); err == nil {
		return &ve, nil
	}

	vePtr, err := s.repo.GetVoterEducationBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	if vePtr != nil {
		s.cache.Set(ctx, cacheKey, vePtr, electionCacheTTL)
	}

	return vePtr, nil
}

func (s *ElectionService) ListVoterEducation(ctx context.Context, electionID *uuid.UUID, category *string, page, perPage int) (*models.PaginatedVoterEducation, error) {
	return s.repo.ListVoterEducation(ctx, electionID, category, page, perPage)
}

func (s *ElectionService) IncrementVoterEducationViewCount(ctx context.Context, id uuid.UUID) error {
	return s.repo.IncrementVoterEducationViewCount(ctx, id)
}

// Helper methods

func (s *ElectionService) invalidateElectionCache(ctx context.Context, id uuid.UUID, slug string) {
	s.cache.Delete(ctx, electionCachePrefix+"id:"+id.String())
	s.cache.Delete(ctx, electionCachePrefix+"slug:"+slug)
	s.cache.DeletePattern(ctx, electionsCachePrefix+"*")
}
