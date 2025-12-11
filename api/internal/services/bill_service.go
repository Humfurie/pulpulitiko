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
	billCachePrefix       = "bill:"
	billsCachePrefix      = "bills:"
	sessionsCachePrefix   = "sessions:"
	committeesCachePrefix = "committees:"
	topicsCachePrefix     = "topics:"
	billCacheTTL          = 1 * time.Hour
	sessionsCacheTTL      = 24 * time.Hour
	committeesCacheTTL    = 24 * time.Hour
	topicsCacheTTL        = 24 * time.Hour
)

type BillService struct {
	repo  *repository.BillRepository
	cache *cache.RedisCache
}

func NewBillService(repo *repository.BillRepository, cache *cache.RedisCache) *BillService {
	return &BillService{
		repo:  repo,
		cache: cache,
	}
}

// Legislative Sessions

func (s *BillService) GetCurrentSession(ctx context.Context) (*models.LegislativeSession, error) {
	cacheKey := sessionsCachePrefix + "current"

	var session models.LegislativeSession
	if err := s.cache.Get(ctx, cacheKey, &session); err == nil {
		return &session, nil
	}

	sessionPtr, err := s.repo.GetCurrentSession(ctx)
	if err != nil {
		return nil, err
	}

	if sessionPtr != nil {
		_ = s.cache.Set(ctx, cacheKey, sessionPtr, sessionsCacheTTL)
	}

	return sessionPtr, nil
}

func (s *BillService) ListSessions(ctx context.Context) ([]models.LegislativeSessionListItem, error) {
	cacheKey := sessionsCachePrefix + "all"

	var sessions []models.LegislativeSessionListItem
	if err := s.cache.Get(ctx, cacheKey, &sessions); err == nil {
		return sessions, nil
	}

	sessions, err := s.repo.ListSessions(ctx)
	if err != nil {
		return nil, err
	}

	_ = s.cache.Set(ctx, cacheKey, sessions, sessionsCacheTTL)

	return sessions, nil
}

// Committees

func (s *BillService) ListCommittees(ctx context.Context, chamber *string) ([]models.CommitteeListItem, error) {
	chamberStr := "all"
	if chamber != nil {
		chamberStr = *chamber
	}
	cacheKey := committeesCachePrefix + chamberStr

	var committees []models.CommitteeListItem
	if err := s.cache.Get(ctx, cacheKey, &committees); err == nil {
		return committees, nil
	}

	committees, err := s.repo.ListCommittees(ctx, chamber)
	if err != nil {
		return nil, err
	}

	_ = s.cache.Set(ctx, cacheKey, committees, committeesCacheTTL)

	return committees, nil
}

func (s *BillService) GetCommitteeBySlug(ctx context.Context, slug string) (*models.Committee, error) {
	cacheKey := committeesCachePrefix + "slug:" + slug

	var committee models.Committee
	if err := s.cache.Get(ctx, cacheKey, &committee); err == nil {
		return &committee, nil
	}

	committeePtr, err := s.repo.GetCommitteeBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	if committeePtr != nil {
		_ = s.cache.Set(ctx, cacheKey, committeePtr, committeesCacheTTL)
	}

	return committeePtr, nil
}

// Bills

func (s *BillService) CreateBill(ctx context.Context, req *models.CreateBillRequest) (*models.Bill, error) {
	bill, err := s.repo.Create(ctx, req)
	if err != nil {
		return nil, err
	}

	// Invalidate bills list cache
	_ = s.cache.DeletePattern(ctx, billsCachePrefix+"*")

	return bill, nil
}

func (s *BillService) GetBillByID(ctx context.Context, id uuid.UUID) (*models.Bill, error) {
	cacheKey := billCachePrefix + "id:" + id.String()

	var bill models.Bill
	if err := s.cache.Get(ctx, cacheKey, &bill); err == nil {
		return &bill, nil
	}

	billPtr, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if billPtr != nil {
		_ = s.cache.Set(ctx, cacheKey, billPtr, billCacheTTL)
	}

	return billPtr, nil
}

func (s *BillService) GetBillBySlug(ctx context.Context, slug string) (*models.Bill, error) {
	cacheKey := billCachePrefix + "slug:" + slug

	var bill models.Bill
	if err := s.cache.Get(ctx, cacheKey, &bill); err == nil {
		return &bill, nil
	}

	billPtr, err := s.repo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	if billPtr != nil {
		_ = s.cache.Set(ctx, cacheKey, billPtr, billCacheTTL)
	}

	return billPtr, nil
}

func (s *BillService) ListBills(ctx context.Context, filter *models.BillFilter, page, perPage int) (*models.PaginatedBills, error) {
	// Don't cache filtered results to ensure freshness
	return s.repo.List(ctx, filter, page, perPage)
}

func (s *BillService) UpdateBill(ctx context.Context, id uuid.UUID, req *models.UpdateBillRequest) (*models.Bill, error) {
	bill, err := s.repo.Update(ctx, id, req)
	if err != nil {
		return nil, err
	}

	if bill != nil {
		// Invalidate caches
		_ = s.cache.Delete(ctx, billCachePrefix+"id:"+id.String())
		_ = s.cache.Delete(ctx, billCachePrefix+"slug:"+bill.Slug)
		_ = s.cache.DeletePattern(ctx, billsCachePrefix+"*")
	}

	return bill, nil
}

func (s *BillService) DeleteBill(ctx context.Context, id uuid.UUID) error {
	// Get bill first for cache invalidation
	bill, _ := s.repo.GetByID(ctx, id)

	err := s.repo.Delete(ctx, id)
	if err != nil {
		return err
	}

	// Invalidate caches
	_ = s.cache.Delete(ctx, billCachePrefix+"id:"+id.String())
	if bill != nil {
		_ = s.cache.Delete(ctx, billCachePrefix+"slug:"+bill.Slug)
	}
	_ = s.cache.DeletePattern(ctx, billsCachePrefix+"*")

	return nil
}

// Bill Status History

func (s *BillService) GetBillStatusHistory(ctx context.Context, billID uuid.UUID) ([]models.BillStatusHistoryItem, error) {
	return s.repo.GetBillStatusHistory(ctx, billID)
}

func (s *BillService) AddBillStatus(ctx context.Context, billID uuid.UUID, req *models.AddBillStatusRequest) error {
	err := s.repo.AddBillStatus(ctx, billID, req)
	if err != nil {
		return err
	}

	// Invalidate bill cache
	s.invalidateBillCache(ctx, billID)

	return nil
}

// Bill Authors

func (s *BillService) GetBillAuthors(ctx context.Context, billID uuid.UUID) ([]models.BillAuthor, error) {
	return s.repo.GetBillAuthors(ctx, billID)
}

// Bill Topics

func (s *BillService) GetBillTopics(ctx context.Context, billID uuid.UUID) ([]models.BillTopic, error) {
	return s.repo.GetBillTopics(ctx, billID)
}

func (s *BillService) ListAllTopics(ctx context.Context) ([]models.BillTopic, error) {
	cacheKey := topicsCachePrefix + "all"

	var topics []models.BillTopic
	if err := s.cache.Get(ctx, cacheKey, &topics); err == nil {
		return topics, nil
	}

	topics, err := s.repo.ListAllTopics(ctx)
	if err != nil {
		return nil, err
	}

	_ = s.cache.Set(ctx, cacheKey, topics, topicsCacheTTL)

	return topics, nil
}

// Bill Committees

func (s *BillService) GetBillCommittees(ctx context.Context, billID uuid.UUID) ([]models.BillCommittee, error) {
	return s.repo.GetBillCommittees(ctx, billID)
}

// Bill Votes

func (s *BillService) GetBillVotes(ctx context.Context, billID uuid.UUID) ([]models.BillVote, error) {
	return s.repo.GetBillVotes(ctx, billID)
}

func (s *BillService) AddBillVote(ctx context.Context, billID uuid.UUID, req *models.AddBillVoteRequest) (*models.BillVote, error) {
	vote, err := s.repo.AddBillVote(ctx, billID, req)
	if err != nil {
		return nil, err
	}

	// Invalidate bill cache
	s.invalidateBillCache(ctx, billID)

	return vote, nil
}

// Politician Votes

func (s *BillService) GetPoliticianVotesForBill(ctx context.Context, billVoteID uuid.UUID) ([]models.PoliticianVote, error) {
	return s.repo.GetPoliticianVotesForBill(ctx, billVoteID)
}

func (s *BillService) GetPoliticianVotingHistory(ctx context.Context, politicianID uuid.UUID, page, perPage int) (*models.PaginatedPoliticianVotes, error) {
	return s.repo.GetPoliticianVotingHistory(ctx, politicianID, page, perPage)
}

func (s *BillService) GetPoliticianVotingRecord(ctx context.Context, politicianID uuid.UUID) (*models.PoliticianVotingRecord, error) {
	cacheKey := fmt.Sprintf("politician:%s:voting_record", politicianID.String())

	var record models.PoliticianVotingRecord
	if err := s.cache.Get(ctx, cacheKey, &record); err == nil {
		return &record, nil
	}

	recordPtr, err := s.repo.GetPoliticianVotingRecord(ctx, politicianID)
	if err != nil {
		return nil, err
	}

	if recordPtr != nil {
		_ = s.cache.Set(ctx, cacheKey, recordPtr, billCacheTTL)
	}

	return recordPtr, nil
}

// Helper methods

func (s *BillService) invalidateBillCache(ctx context.Context, billID uuid.UUID) {
	bill, _ := s.repo.GetByID(ctx, billID)
	_ = s.cache.Delete(ctx, billCachePrefix+"id:"+billID.String())
	if bill != nil {
		_ = s.cache.Delete(ctx, billCachePrefix+"slug:"+bill.Slug)
	}
	_ = s.cache.DeletePattern(ctx, billsCachePrefix+"*")
}
