package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/humfurie/pulpulitiko/api/internal/models"
	"github.com/humfurie/pulpulitiko/api/internal/repository"
	"github.com/humfurie/pulpulitiko/api/pkg/cache"
)

const (
	pollCachePrefix        = "poll:"
	pollsCachePrefix       = "polls:"
	pollResultsCachePrefix = "poll_results:"
	pollCacheTTL           = 5 * time.Minute
	pollResultsCacheTTL    = 1 * time.Minute
)

type PollService struct {
	repo  *repository.PollRepository
	cache *cache.RedisCache
}

func NewPollService(repo *repository.PollRepository, cache *cache.RedisCache) *PollService {
	return &PollService{
		repo:  repo,
		cache: cache,
	}
}

// Polls

func (s *PollService) CreatePoll(ctx context.Context, userID uuid.UUID, req *models.CreatePollRequest) (*models.Poll, error) {
	poll, err := s.repo.CreatePoll(ctx, userID, req)
	if err != nil {
		return nil, err
	}

	_ = s.cache.DeletePattern(ctx, pollsCachePrefix+"*")

	return poll, nil
}

func (s *PollService) GetPollByID(ctx context.Context, id uuid.UUID, userID *uuid.UUID, ipHash *string) (*models.Poll, error) {
	poll, err := s.repo.GetPollByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if poll == nil {
		return nil, nil
	}

	// Check if user has voted
	if userID != nil || ipHash != nil {
		hasVoted, optionID := s.repo.HasUserVoted(ctx, id, userID, ipHash)
		if hasVoted {
			poll.UserVote = optionID
		}
	}

	// Calculate percentages for options
	if poll.TotalVotes > 0 {
		for i := range poll.Options {
			poll.Options[i].Percentage = float64(poll.Options[i].VoteCount) / float64(poll.TotalVotes) * 100
		}
	}

	return poll, nil
}

func (s *PollService) GetPollBySlug(ctx context.Context, slug string, userID *uuid.UUID, ipHash *string) (*models.Poll, error) {
	poll, err := s.repo.GetPollBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if poll == nil {
		return nil, nil
	}

	// Check if user has voted
	if userID != nil || ipHash != nil {
		hasVoted, optionID := s.repo.HasUserVoted(ctx, poll.ID, userID, ipHash)
		if hasVoted {
			poll.UserVote = optionID
		}
	}

	// Calculate percentages for options
	if poll.TotalVotes > 0 {
		for i := range poll.Options {
			poll.Options[i].Percentage = float64(poll.Options[i].VoteCount) / float64(poll.TotalVotes) * 100
		}
	}

	return poll, nil
}

func (s *PollService) ListPolls(ctx context.Context, filter *models.PollFilter, page, perPage int) (*models.PaginatedPolls, error) {
	return s.repo.ListPolls(ctx, filter, page, perPage)
}

func (s *PollService) GetActivePolls(ctx context.Context, page, perPage int) (*models.PaginatedPolls, error) {
	filter := &models.PollFilter{
		ActiveOnly: true,
	}
	return s.repo.ListPolls(ctx, filter, page, perPage)
}

func (s *PollService) GetFeaturedPolls(ctx context.Context, limit int) ([]models.PollListItem, error) {
	cacheKey := fmt.Sprintf("%sfeatured:%d", pollsCachePrefix, limit)

	var polls []models.PollListItem
	if err := s.cache.Get(ctx, cacheKey, &polls); err == nil {
		return polls, nil
	}

	polls, err := s.repo.GetFeaturedPolls(ctx, limit)
	if err != nil {
		return nil, err
	}

	_ = s.cache.Set(ctx, cacheKey, polls, pollCacheTTL)

	return polls, nil
}

func (s *PollService) GetUserPolls(ctx context.Context, userID uuid.UUID, page, perPage int) (*models.PaginatedPolls, error) {
	filter := &models.PollFilter{
		UserID: &userID,
	}
	return s.repo.ListPolls(ctx, filter, page, perPage)
}

func (s *PollService) UpdatePoll(ctx context.Context, id uuid.UUID, req *models.UpdatePollRequest) (*models.Poll, error) {
	poll, err := s.repo.UpdatePoll(ctx, id, req)
	if err != nil {
		return nil, err
	}

	s.invalidatePollCache(ctx, id)

	return poll, nil
}

func (s *PollService) AdminUpdatePoll(ctx context.Context, id uuid.UUID, req *models.AdminUpdatePollRequest) (*models.Poll, error) {
	poll, err := s.repo.AdminUpdatePoll(ctx, id, req)
	if err != nil {
		return nil, err
	}

	s.invalidatePollCache(ctx, id)

	return poll, nil
}

func (s *PollService) SubmitForApproval(ctx context.Context, id uuid.UUID) error {
	req := &models.AdminUpdatePollRequest{}
	status := models.PollStatusPendingApproval
	req.Status = &status

	_, err := s.repo.AdminUpdatePoll(ctx, id, req)
	if err != nil {
		return err
	}

	s.invalidatePollCache(ctx, id)
	return nil
}

func (s *PollService) ApprovePoll(ctx context.Context, id uuid.UUID, approverID uuid.UUID, approved bool, reason *string) error {
	err := s.repo.ApprovePoll(ctx, id, approverID, approved, reason)
	if err != nil {
		return err
	}

	s.invalidatePollCache(ctx, id)
	return nil
}

func (s *PollService) ClosePoll(ctx context.Context, id uuid.UUID) error {
	err := s.repo.ClosePoll(ctx, id)
	if err != nil {
		return err
	}

	s.invalidatePollCache(ctx, id)
	return nil
}

func (s *PollService) DeletePoll(ctx context.Context, id uuid.UUID) error {
	err := s.repo.DeletePoll(ctx, id)
	if err != nil {
		return err
	}

	s.invalidatePollCache(ctx, id)
	return nil
}

func (s *PollService) IncrementViewCount(ctx context.Context, id uuid.UUID) error {
	return s.repo.IncrementViewCount(ctx, id)
}

// Voting

func (s *PollService) CastVote(ctx context.Context, pollID, optionID uuid.UUID, userID *uuid.UUID, ip string) (*models.VoteResponse, error) {
	// Get poll to check settings
	poll, err := s.repo.GetPollByID(ctx, pollID)
	if err != nil {
		return nil, err
	}
	if poll == nil {
		return &models.VoteResponse{
			Success: false,
			Message: "Poll not found",
		}, nil
	}

	// Check if poll is active
	if poll.Status != models.PollStatusActive {
		return &models.VoteResponse{
			Success: false,
			Message: "This poll is not currently active",
		}, nil
	}

	// Check time constraints
	now := time.Now()
	if poll.StartsAt != nil && now.Before(*poll.StartsAt) {
		return &models.VoteResponse{
			Success: false,
			Message: "This poll has not started yet",
		}, nil
	}
	if poll.EndsAt != nil && now.After(*poll.EndsAt) {
		return &models.VoteResponse{
			Success: false,
			Message: "This poll has ended",
		}, nil
	}

	// Hash IP for anonymous voting
	var ipHash *string
	if poll.IsAnonymous && userID == nil {
		hash := sha256.Sum256([]byte(ip + pollID.String()))
		hashStr := hex.EncodeToString(hash[:])
		ipHash = &hashStr
	}

	// Cast vote
	err = s.repo.CastVote(ctx, pollID, optionID, userID, ipHash)
	if err != nil {
		return &models.VoteResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// Invalidate results cache
	_ = s.cache.Delete(ctx, pollResultsCachePrefix+pollID.String())

	// Get updated results
	results, err := s.GetPollResults(ctx, pollID)
	if err != nil {
		return &models.VoteResponse{
			Success: true,
			Message: "Vote recorded successfully",
		}, nil
	}

	return &models.VoteResponse{
		Success: true,
		Message: "Vote recorded successfully",
		Results: results,
	}, nil
}

func (s *PollService) HasUserVoted(ctx context.Context, pollID uuid.UUID, userID *uuid.UUID, ipHash *string) (bool, *uuid.UUID) {
	return s.repo.HasUserVoted(ctx, pollID, userID, ipHash)
}

func (s *PollService) GetPollResults(ctx context.Context, pollID uuid.UUID) (*models.PollResults, error) {
	cacheKey := pollResultsCachePrefix + pollID.String()

	var results models.PollResults
	if err := s.cache.Get(ctx, cacheKey, &results); err == nil {
		return &results, nil
	}

	resultsPtr, err := s.repo.GetPollResults(ctx, pollID)
	if err != nil {
		return nil, err
	}

	if resultsPtr != nil {
		_ = s.cache.Set(ctx, cacheKey, resultsPtr, pollResultsCacheTTL)
	}

	return resultsPtr, nil
}

// Comments

func (s *PollService) CreatePollComment(ctx context.Context, pollID, userID uuid.UUID, req *models.CreatePollCommentRequest) (*models.PollComment, error) {
	return s.repo.CreatePollComment(ctx, pollID, userID, req)
}

func (s *PollService) GetPollComments(ctx context.Context, pollID uuid.UUID, page, perPage int) (*models.PaginatedPollComments, error) {
	return s.repo.GetPollComments(ctx, pollID, page, perPage)
}

func (s *PollService) DeletePollComment(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeletePollComment(ctx, id)
}

// Helper methods

func (s *PollService) invalidatePollCache(ctx context.Context, id uuid.UUID) {
	_ = s.cache.Delete(ctx, pollCachePrefix+"id:"+id.String())
	_ = s.cache.Delete(ctx, pollResultsCachePrefix+id.String())
	_ = s.cache.DeletePattern(ctx, pollsCachePrefix+"*")
}

// HashIP creates a hash of IP + poll ID for anonymous vote tracking
func HashIP(ip string, pollID uuid.UUID) string {
	hash := sha256.Sum256([]byte(ip + pollID.String()))
	return hex.EncodeToString(hash[:])
}
