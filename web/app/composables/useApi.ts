import type {
  ApiResponse,
  Article,
  ArticleListItem,
  Author,
  AuthorWithArticles,
  Barangay,
  Bill,
  BillFilter,
  BillTopic,
  BillVote,
  Candidate,
  CandidateListItem,
  Category,
  CategoryWithArticles,
  CityMunicipalityListItem,
  CityWithBarangays,
  Comment,
  CommentAuthor,
  CommentCountResponse,
  Committee,
  CommitteeListItem,
  CongressionalDistrict,
  CreateCommentRequest,
  CreatePollCommentRequest,
  CreatePollRequest,
  DistrictListItem,
  Election,
  ElectionCalendarItem,
  ElectionFilter,
  ElectionListItem,
  ElectionPositionListItem,
  GovernmentPosition,
  GovernmentPositionListItem,
  LegislativeChamber,
  LegislativeSession,
  LegislativeSessionListItem,
  LocationHierarchy,
  LocationSearchResult,
  PaginatedArticles,
  PaginatedBarangays,
  PaginatedBills,
  PaginatedCandidates,
  PaginatedElections,
  PaginatedNotifications,
  PaginatedPoliticalParties,
  PaginatedPoliticianComments,
  PaginatedPoliticianVotes,
  PaginatedPollComments,
  PaginatedPolls,
  PaginatedVoterEducation,
  PoliticalParty,
  PoliticalPartyListItem,
  Politician,
  PoliticianComment,
  PoliticianListItem,
  PoliticianVote,
  PoliticianVotingRecord,
  PoliticianWithArticles,
  Poll,
  PollCategory,
  PollComment,
  PollFilter,
  PollListItem,
  PollResults,
  PollStatus,
  ProvinceListItem,
  ProvinceWithCities,
  RegionListItem,
  RegionWithProvinces,
  Tag,
  TagWithArticles,
  UpdatePollRequest,
  UploadResult,
  UserProfile,
  VoteResponse,
  VoterEducation
} from '~/types'

type HttpMethod = 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH'

// eslint-disable-next-line @typescript-eslint/no-explicit-any
type RequestBody = Record<string, any> | FormData | null | undefined

interface FetchOptions {
  method?: HttpMethod
  headers?: Record<string, string>
  body?: RequestBody
}

export function useApi() {
  const config = useRuntimeConfig()
  // Use internal Docker URL for SSR, public URL for client-side
  const baseUrl = import.meta.server
    ? config.apiInternalUrl
    : config.public.apiUrl

  async function fetchApi<T>(endpoint: string, options?: FetchOptions): Promise<T> {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      ...options?.headers
    }

    const response = await $fetch<ApiResponse<T>>(`${baseUrl}${endpoint}`, {
      method: options?.method,
      headers,
      body: options?.body
    })

    if (!response.success) {
      throw new Error((response as unknown as { error: string }).error || 'API request failed')
    }

    return response.data
  }

  return {
    // Articles
    async getArticles(page = 1, perPage = 10): Promise<PaginatedArticles> {
      return fetchApi<PaginatedArticles>(`/articles?page=${page}&per_page=${perPage}`)
    },

    async getArticleBySlug(slug: string): Promise<Article> {
      return fetchApi<Article>(`/articles/${slug}`)
    },

    async trackArticleView(slug: string): Promise<void> {
      try {
        await $fetch(`${baseUrl}/articles/${slug}/view`, { method: 'POST' })
      } catch {
        // Silently fail - view tracking shouldn't break the page
      }
    },

    async getTrendingArticles(): Promise<ArticleListItem[]> {
      return fetchApi<ArticleListItem[]>('/articles/trending')
    },

    async getRelatedArticles(slug: string): Promise<ArticleListItem[]> {
      return fetchApi<ArticleListItem[]>(`/articles/${slug}/related`)
    },

    async searchArticles(query: string, page = 1, perPage = 10): Promise<PaginatedArticles> {
      return fetchApi<PaginatedArticles>(
        `/search?q=${encodeURIComponent(query)}&page=${page}&per_page=${perPage}`
      )
    },

    // Categories
    async getCategories(): Promise<Category[]> {
      return fetchApi<Category[]>('/categories')
    },

    async getCategoryArticles(slug: string, page = 1, perPage = 10): Promise<CategoryWithArticles> {
      return fetchApi<CategoryWithArticles>(
        `/categories/${slug}?page=${page}&per_page=${perPage}`
      )
    },

    // Tags
    async getTags(): Promise<Tag[]> {
      return fetchApi<Tag[]>('/tags')
    },

    async getTagArticles(slug: string, page = 1, perPage = 10): Promise<TagWithArticles> {
      return fetchApi<TagWithArticles>(`/tags/${slug}?page=${page}&per_page=${perPage}`)
    },

    // Authors
    async getAuthors(): Promise<Author[]> {
      return fetchApi<Author[]>('/authors')
    },

    async getAuthorArticles(slug: string, page = 1, perPage = 10): Promise<AuthorWithArticles> {
      return fetchApi<AuthorWithArticles>(`/authors/${slug}?page=${page}&per_page=${perPage}`)
    },

    // Politicians
    async getPoliticians(): Promise<Politician[]> {
      return fetchApi<Politician[]>('/politicians')
    },

    async searchPoliticians(query: string, limit = 10): Promise<Politician[]> {
      return fetchApi<Politician[]>(`/politicians/search?q=${encodeURIComponent(query)}&limit=${limit}`)
    },

    async getPoliticianArticles(slug: string, page = 1, perPage = 10): Promise<PoliticianWithArticles> {
      return fetchApi<PoliticianWithArticles>(`/politicians/${slug}?page=${page}&per_page=${perPage}`)
    },

    // Upload
    async uploadFile(file: File, authHeaders: HeadersInit): Promise<UploadResult> {
      const formData = new FormData()
      formData.append('file', file)

      // Extract only the Authorization header - don't set Content-Type for FormData
      const headers: Record<string, string> = {}
      if (authHeaders && typeof authHeaders === 'object') {
        const authRecord = authHeaders as Record<string, string>
        if (authRecord.Authorization) {
          headers.Authorization = authRecord.Authorization
        }
      }

      const response = await $fetch<ApiResponse<UploadResult>>(`${baseUrl}/admin/upload`, {
        method: 'POST',
        headers,
        body: formData
      })

      if (!response.success) {
        throw new Error((response as unknown as { error: string }).error || 'Upload failed')
      }

      return response.data
    },

    // Comments
    async getArticleComments(
      slug: string,
      authHeaders?: Record<string, string>,
      page = 1,
      pageSize = 10,
      sort: 'recent' | 'liked' | 'oldest' = 'recent'
    ): Promise<Comment[]> {
      const params = new URLSearchParams({
        page: String(page),
        page_size: String(pageSize),
        sort
      })
      return fetchApi<Comment[]>(`/articles/${slug}/comments?${params}`, { headers: authHeaders })
    },

    async getCommentCount(slug: string): Promise<CommentCountResponse> {
      return fetchApi<CommentCountResponse>(`/articles/${slug}/comments/count`)
    },

    async getComment(id: string): Promise<Comment> {
      return fetchApi<Comment>(`/comments/${id}`)
    },

    async getCommentReplies(id: string, authHeaders?: Record<string, string>): Promise<Comment[]> {
      return fetchApi<Comment[]>(`/comments/${id}/replies`, { headers: authHeaders })
    },

    async createComment(slug: string, data: CreateCommentRequest, authHeaders: Record<string, string>): Promise<Comment> {
      return fetchApi<Comment>(`/articles/${slug}/comments`, {
        method: 'POST',
        headers: authHeaders,
        body: data
      })
    },

    async updateComment(id: string, content: string, authHeaders: Record<string, string>): Promise<Comment> {
      return fetchApi<Comment>(`/comments/${id}`, {
        method: 'PUT',
        headers: authHeaders,
        body: { content }
      })
    },

    async deleteComment(id: string, authHeaders: Record<string, string>): Promise<void> {
      await fetchApi<{ message: string }>(`/comments/${id}`, {
        method: 'DELETE',
        headers: authHeaders
      })
    },

    async addReaction(commentId: string, reaction: string, authHeaders: Record<string, string>): Promise<void> {
      await fetchApi<{ message: string }>(`/comments/${commentId}/reactions`, {
        method: 'POST',
        headers: authHeaders,
        body: { reaction }
      })
    },

    async removeReaction(commentId: string, reaction: string, authHeaders: Record<string, string>): Promise<void> {
      await fetchApi<{ message: string }>(`/comments/${commentId}/reactions/${reaction}`, {
        method: 'DELETE',
        headers: authHeaders
      })
    },

    // Users (for mentions)
    async getMentionableUsers(): Promise<CommentAuthor[]> {
      return fetchApi<CommentAuthor[]>('/users/mentionable')
    },

    // User profiles
    async getUserProfile(slug: string): Promise<UserProfile> {
      return fetchApi<UserProfile>(`/users/${slug}/profile`)
    },

    async getUserComments(slug: string, page = 1, pageSize = 10): Promise<Comment[]> {
      return fetchApi<Comment[]>(`/users/${slug}/comments?page=${page}&page_size=${pageSize}`)
    },

    async getUserReplies(slug: string, page = 1, pageSize = 10): Promise<Comment[]> {
      return fetchApi<Comment[]>(`/users/${slug}/replies?page=${page}&page_size=${pageSize}`)
    },

    // Search Analytics
    async trackSearch(query: string, resultsCount: number, sessionId?: string): Promise<{ search_query_id: string }> {
      return fetchApi<{ search_query_id: string }>('/search/track', {
        method: 'POST',
        body: { query, results_count: resultsCount, session_id: sessionId }
      })
    },

    async trackClick(searchQueryId: string, articleId: string, position: number): Promise<void> {
      await fetchApi<{ message: string }>('/search/click', {
        method: 'POST',
        body: { search_query_id: searchQueryId, article_id: articleId, position }
      })
    },

    // Politician Comments
    async getPoliticianComments(
      slug: string,
      authHeaders?: Record<string, string>,
      page = 1,
      perPage = 10
    ): Promise<PaginatedPoliticianComments> {
      const params = new URLSearchParams({
        page: String(page),
        per_page: String(perPage)
      })
      return fetchApi<PaginatedPoliticianComments>(`/politicians/${slug}/comments?${params}`, { headers: authHeaders })
    },

    async getPoliticianCommentCount(slug: string): Promise<CommentCountResponse> {
      return fetchApi<CommentCountResponse>(`/politicians/${slug}/comments/count`)
    },

    async getPoliticianComment(id: string): Promise<PoliticianComment> {
      return fetchApi<PoliticianComment>(`/politician-comments/${id}`)
    },

    async getPoliticianCommentReplies(id: string, authHeaders?: Record<string, string>): Promise<PoliticianComment[]> {
      return fetchApi<PoliticianComment[]>(`/politician-comments/${id}/replies`, { headers: authHeaders })
    },

    async createPoliticianComment(slug: string, data: CreateCommentRequest, authHeaders: Record<string, string>): Promise<PoliticianComment> {
      return fetchApi<PoliticianComment>(`/politicians/${slug}/comments`, {
        method: 'POST',
        headers: authHeaders,
        body: data
      })
    },

    async updatePoliticianComment(id: string, content: string, authHeaders: Record<string, string>): Promise<PoliticianComment> {
      return fetchApi<PoliticianComment>(`/politician-comments/${id}`, {
        method: 'PUT',
        headers: authHeaders,
        body: { content }
      })
    },

    async deletePoliticianComment(id: string, authHeaders: Record<string, string>): Promise<void> {
      await fetchApi<{ message: string }>(`/politician-comments/${id}`, {
        method: 'DELETE',
        headers: authHeaders
      })
    },

    async addPoliticianCommentReaction(commentId: string, reaction: string, authHeaders: Record<string, string>): Promise<void> {
      await fetchApi<{ message: string }>(`/politician-comments/${commentId}/reactions`, {
        method: 'POST',
        headers: authHeaders,
        body: { reaction }
      })
    },

    async removePoliticianCommentReaction(commentId: string, reaction: string, authHeaders: Record<string, string>): Promise<void> {
      await fetchApi<{ message: string }>(`/politician-comments/${commentId}/reactions/${reaction}`, {
        method: 'DELETE',
        headers: authHeaders
      })
    },

    // Notifications
    async getNotifications(
      authHeaders: Record<string, string>,
      page = 1,
      perPage = 20,
      unreadOnly = false
    ): Promise<PaginatedNotifications> {
      const params = new URLSearchParams({
        page: String(page),
        per_page: String(perPage)
      })
      if (unreadOnly) {
        params.set('unread_only', 'true')
      }
      return fetchApi<PaginatedNotifications>(`/notifications?${params}`, { headers: authHeaders })
    },

    async getUnreadNotificationCount(authHeaders: Record<string, string>): Promise<{ count: number }> {
      return fetchApi<{ count: number }>('/notifications/unread-count', { headers: authHeaders })
    },

    async markNotificationAsRead(id: string, authHeaders: Record<string, string>): Promise<void> {
      await fetchApi<{ message: string }>(`/notifications/${id}/read`, {
        method: 'POST',
        headers: authHeaders
      })
    },

    async markAllNotificationsAsRead(authHeaders: Record<string, string>): Promise<void> {
      await fetchApi<{ message: string }>('/notifications/read-all', {
        method: 'POST',
        headers: authHeaders
      })
    },

    async deleteNotification(id: string, authHeaders: Record<string, string>): Promise<void> {
      await fetchApi<{ message: string }>(`/notifications/${id}`, {
        method: 'DELETE',
        headers: authHeaders
      })
    },

    // =====================================================
    // LOCATIONS (Philippine Geographic Hierarchy)
    // =====================================================

    // Regions
    async getRegions(): Promise<RegionListItem[]> {
      return fetchApi<RegionListItem[]>('/locations/regions')
    },

    async getRegionBySlug(slug: string): Promise<RegionWithProvinces> {
      return fetchApi<RegionWithProvinces>(`/locations/regions/${slug}`)
    },

    // Provinces
    async getAllProvinces(): Promise<ProvinceListItem[]> {
      return fetchApi<ProvinceListItem[]>('/locations/provinces')
    },

    async getProvinceBySlug(slug: string): Promise<ProvinceWithCities> {
      return fetchApi<ProvinceWithCities>(`/locations/provinces/${slug}`)
    },

    async getProvincesByRegion(regionId: string): Promise<ProvinceListItem[]> {
      return fetchApi<ProvinceListItem[]>(`/locations/provinces/by-region/${regionId}`)
    },

    // Cities/Municipalities
    async getCityBySlug(slug: string): Promise<CityWithBarangays> {
      return fetchApi<CityWithBarangays>(`/locations/cities/${slug}`)
    },

    async getCitiesByProvince(provinceId: string): Promise<CityMunicipalityListItem[]> {
      return fetchApi<CityMunicipalityListItem[]>(`/locations/cities/by-province/${provinceId}`)
    },

    // Barangays
    async getBarangayBySlug(slug: string): Promise<Barangay> {
      return fetchApi<Barangay>(`/locations/barangays/${slug}`)
    },

    async getBarangaysByCity(cityId: string, page = 1, perPage = 50): Promise<PaginatedBarangays> {
      return fetchApi<PaginatedBarangays>(`/locations/barangays/by-city/${cityId}?page=${page}&per_page=${perPage}`)
    },

    // Congressional Districts
    async getDistrictBySlug(slug: string): Promise<CongressionalDistrict> {
      return fetchApi<CongressionalDistrict>(`/locations/districts/${slug}`)
    },

    async getDistrictsByProvince(provinceId: string): Promise<DistrictListItem[]> {
      return fetchApi<DistrictListItem[]>(`/locations/districts/by-province/${provinceId}`)
    },

    // Search & Hierarchy
    async searchLocations(query: string, limit = 20): Promise<LocationSearchResult[]> {
      return fetchApi<LocationSearchResult[]>(`/locations/search?q=${encodeURIComponent(query)}&limit=${limit}`)
    },

    async getLocationHierarchy(barangayId: string): Promise<LocationHierarchy> {
      return fetchApi<LocationHierarchy>(`/locations/hierarchy/${barangayId}`)
    },

    // =====================================================
    // POLITICAL PARTIES
    // =====================================================

    async getParties(page = 1, perPage = 20, majorOnly = false, activeOnly = true): Promise<PaginatedPoliticalParties> {
      const params = new URLSearchParams({
        page: String(page),
        per_page: String(perPage),
        major_only: String(majorOnly),
        active_only: String(activeOnly)
      })
      return fetchApi<PaginatedPoliticalParties>(`/parties?${params}`)
    },

    async getAllParties(activeOnly = true): Promise<PoliticalPartyListItem[]> {
      return fetchApi<PoliticalPartyListItem[]>(`/parties/all?active_only=${activeOnly}`)
    },

    async getPartyBySlug(slug: string): Promise<PoliticalParty> {
      return fetchApi<PoliticalParty>(`/parties/${slug}`)
    },

    // =====================================================
    // GOVERNMENT POSITIONS
    // =====================================================

    async getAllPositions(): Promise<GovernmentPositionListItem[]> {
      return fetchApi<GovernmentPositionListItem[]>('/positions')
    },

    async getPositionsByLevel(level: string): Promise<GovernmentPositionListItem[]> {
      return fetchApi<GovernmentPositionListItem[]>(`/positions/level/${level}`)
    },

    async getPositionBySlug(slug: string): Promise<GovernmentPosition> {
      return fetchApi<GovernmentPosition>(`/positions/${slug}`)
    },

    // =====================================================
    // FIND MY REPRESENTATIVES
    // =====================================================

    async findMyRepresentatives(barangayId: string): Promise<PoliticianListItem[]> {
      return fetchApi<PoliticianListItem[]>(`/my-representatives?barangay_id=${barangayId}`)
    },

    // =====================================================
    // LEGISLATION / BILLS TRACKER
    // =====================================================

    // Legislative Sessions
    async getLegislativeSessions(): Promise<LegislativeSessionListItem[]> {
      return fetchApi<LegislativeSessionListItem[]>('/legislation/sessions')
    },

    async getCurrentSession(): Promise<LegislativeSession> {
      return fetchApi<LegislativeSession>('/legislation/sessions/current')
    },

    // Committees
    async getCommittees(chamber?: LegislativeChamber): Promise<CommitteeListItem[]> {
      const params = chamber ? `?chamber=${chamber}` : ''
      return fetchApi<CommitteeListItem[]>(`/legislation/committees${params}`)
    },

    async getCommitteeBySlug(slug: string): Promise<Committee> {
      return fetchApi<Committee>(`/legislation/committees/${slug}`)
    },

    // Topics
    async getBillTopics(): Promise<BillTopic[]> {
      return fetchApi<BillTopic[]>('/legislation/topics')
    },

    // Bills
    async getBills(filter?: BillFilter, page = 1, perPage = 20): Promise<PaginatedBills> {
      const params = new URLSearchParams({
        page: String(page),
        per_page: String(perPage)
      })
      if (filter?.chamber) params.set('chamber', filter.chamber)
      if (filter?.status) params.set('status', filter.status)
      if (filter?.session_id) params.set('session_id', filter.session_id)
      if (filter?.topic_id) params.set('topic_id', filter.topic_id)
      if (filter?.author_id) params.set('author_id', filter.author_id)
      if (filter?.search) params.set('search', filter.search)
      return fetchApi<PaginatedBills>(`/legislation/bills?${params}`)
    },

    async getBillBySlug(slug: string): Promise<Bill> {
      return fetchApi<Bill>(`/legislation/bills/${slug}`)
    },

    async getBillById(id: string): Promise<Bill> {
      return fetchApi<Bill>(`/legislation/bills/id/${id}`)
    },

    async getBillVotes(billId: string): Promise<BillVote[]> {
      return fetchApi<BillVote[]>(`/legislation/bills/${billId}/votes`)
    },

    async getPoliticianVotesForBillVote(voteId: string): Promise<PoliticianVote[]> {
      return fetchApi<PoliticianVote[]>(`/legislation/votes/${voteId}/politicians`)
    },

    // Politician Voting Records
    async getPoliticianVotingHistory(politicianId: string, page = 1, perPage = 20): Promise<PaginatedPoliticianVotes> {
      return fetchApi<PaginatedPoliticianVotes>(`/legislation/politicians/${politicianId}/votes?page=${page}&per_page=${perPage}`)
    },

    async getPoliticianVotingRecord(politicianId: string): Promise<PoliticianVotingRecord> {
      return fetchApi<PoliticianVotingRecord>(`/legislation/politicians/${politicianId}/voting-record`)
    },

    // =====================================================
    // ELECTIONS
    // =====================================================

    // Elections
    async getElections(filter?: ElectionFilter, page = 1, perPage = 20): Promise<PaginatedElections> {
      const params = new URLSearchParams({
        page: String(page),
        per_page: String(perPage)
      })
      if (filter?.election_type) params.set('election_type', filter.election_type)
      if (filter?.status) params.set('status', filter.status)
      if (filter?.year) params.set('year', String(filter.year))
      if (filter?.is_featured !== undefined) params.set('is_featured', String(filter.is_featured))
      if (filter?.search) params.set('search', filter.search)
      return fetchApi<PaginatedElections>(`/elections?${params}`)
    },

    async getElectionById(id: string): Promise<Election> {
      return fetchApi<Election>(`/elections/${id}`)
    },

    async getElectionBySlug(slug: string): Promise<Election> {
      return fetchApi<Election>(`/elections/slug/${slug}`)
    },

    async getUpcomingElections(limit = 5): Promise<ElectionListItem[]> {
      return fetchApi<ElectionListItem[]>(`/elections/upcoming?limit=${limit}`)
    },

    async getFeaturedElections(): Promise<ElectionListItem[]> {
      return fetchApi<ElectionListItem[]>('/elections/featured')
    },

    async getElectionCalendar(year?: number): Promise<ElectionCalendarItem[]> {
      const params = year ? `?year=${year}` : ''
      return fetchApi<ElectionCalendarItem[]>(`/elections/calendar${params}`)
    },

    async getElectionPositions(electionId: string): Promise<ElectionPositionListItem[]> {
      return fetchApi<ElectionPositionListItem[]>(`/elections/${electionId}/positions`)
    },

    // Candidates
    async getCandidates(
      filter?: { election_id?: string; position_id?: string; party_id?: string; status?: string; is_winner?: boolean },
      page = 1,
      perPage = 20
    ): Promise<PaginatedCandidates> {
      const params = new URLSearchParams({
        page: String(page),
        per_page: String(perPage)
      })
      if (filter?.election_id) params.set('election_id', filter.election_id)
      if (filter?.position_id) params.set('position_id', filter.position_id)
      if (filter?.party_id) params.set('party_id', filter.party_id)
      if (filter?.status) params.set('status', filter.status)
      if (filter?.is_winner !== undefined) params.set('is_winner', String(filter.is_winner))
      return fetchApi<PaginatedCandidates>(`/candidates?${params}`)
    },

    async getCandidateById(id: string): Promise<Candidate> {
      return fetchApi<Candidate>(`/candidates/${id}`)
    },

    async getCandidatesForPosition(positionId: string): Promise<CandidateListItem[]> {
      return fetchApi<CandidateListItem[]>(`/candidates/position/${positionId}`)
    },

    // Voter Education
    async getVoterEducation(electionId?: string, category?: string, page = 1, perPage = 20): Promise<PaginatedVoterEducation> {
      const params = new URLSearchParams({
        page: String(page),
        per_page: String(perPage)
      })
      if (electionId) params.set('election_id', electionId)
      if (category) params.set('category', category)
      return fetchApi<PaginatedVoterEducation>(`/voter-education?${params}`)
    },

    async getVoterEducationBySlug(slug: string): Promise<VoterEducation> {
      return fetchApi<VoterEducation>(`/voter-education/${slug}`)
    },

    // =====================================================
    // POLLS
    // =====================================================

    // Public poll endpoints
    async getPolls(
      filter?: PollFilter,
      page = 1,
      perPage = 12
    ): Promise<PaginatedPolls> {
      const params = new URLSearchParams({
        page: String(page),
        per_page: String(perPage)
      })
      if (filter?.category) params.set('category', filter.category)
      if (filter?.politician_id) params.set('politician_id', filter.politician_id)
      if (filter?.election_id) params.set('election_id', filter.election_id)
      if (filter?.search) params.set('search', filter.search)
      return fetchApi<PaginatedPolls>(`/polls?${params}`)
    },

    async getPollBySlug(slug: string): Promise<Poll> {
      return fetchApi<Poll>(`/polls/slug/${slug}`)
    },

    async getPollById(id: string): Promise<Poll> {
      return fetchApi<Poll>(`/polls/${id}`)
    },

    async getFeaturedPolls(limit = 5): Promise<PollListItem[]> {
      return fetchApi<PollListItem[]>(`/polls/featured?limit=${limit}`)
    },

    async getPollResults(pollId: string): Promise<PollResults> {
      return fetchApi<PollResults>(`/polls/${pollId}/results`)
    },

    async castVote(pollId: string, optionId: string, authHeaders?: Record<string, string>): Promise<VoteResponse> {
      return fetchApi<VoteResponse>(`/polls/${pollId}/vote`, {
        method: 'POST',
        headers: authHeaders,
        body: { option_id: optionId }
      })
    },

    // Poll comments
    async getPollComments(
      pollId: string,
      authHeaders?: Record<string, string>,
      page = 1,
      perPage = 20
    ): Promise<PaginatedPollComments> {
      const params = new URLSearchParams({
        page: String(page),
        per_page: String(perPage)
      })
      return fetchApi<PaginatedPollComments>(`/polls/${pollId}/comments?${params}`, { headers: authHeaders })
    },

    async createPollComment(
      pollId: string,
      data: CreatePollCommentRequest,
      authHeaders: Record<string, string>
    ): Promise<PollComment> {
      return fetchApi<PollComment>(`/polls/${pollId}/comments`, {
        method: 'POST',
        headers: authHeaders,
        body: data
      })
    },

    // Authenticated user poll endpoints
    async getMyPolls(authHeaders: Record<string, string>, page = 1, perPage = 10): Promise<PaginatedPolls> {
      const params = new URLSearchParams({
        page: String(page),
        per_page: String(perPage)
      })
      return fetchApi<PaginatedPolls>(`/my-polls?${params}`, { headers: authHeaders })
    },

    async createPoll(data: CreatePollRequest, authHeaders: Record<string, string>): Promise<Poll> {
      return fetchApi<Poll>('/my-polls', {
        method: 'POST',
        headers: authHeaders,
        body: data
      })
    },

    async updatePoll(pollId: string, data: UpdatePollRequest, authHeaders: Record<string, string>): Promise<Poll> {
      return fetchApi<Poll>(`/my-polls/${pollId}`, {
        method: 'PUT',
        headers: authHeaders,
        body: data
      })
    },

    async submitPollForApproval(pollId: string, authHeaders: Record<string, string>): Promise<void> {
      await fetchApi<{ message: string }>(`/my-polls/${pollId}/submit`, {
        method: 'POST',
        headers: authHeaders
      })
    },

    async deletePoll(pollId: string, authHeaders: Record<string, string>): Promise<void> {
      await fetchApi<{ message: string }>(`/my-polls/${pollId}`, {
        method: 'DELETE',
        headers: authHeaders
      })
    },

    // Admin poll endpoints
    async adminGetPolls(
      authHeaders: Record<string, string>,
      filter?: { category?: PollCategory; status?: PollStatus; search?: string; is_featured?: boolean },
      page = 1,
      perPage = 20
    ): Promise<PaginatedPolls> {
      const params = new URLSearchParams({
        page: String(page),
        per_page: String(perPage)
      })
      if (filter?.category) params.set('category', filter.category)
      if (filter?.status) params.set('status', filter.status)
      if (filter?.search) params.set('search', filter.search)
      if (filter?.is_featured !== undefined) params.set('is_featured', String(filter.is_featured))
      return fetchApi<PaginatedPolls>(`/admin/polls?${params}`, { headers: authHeaders })
    },

    async adminUpdatePoll(
      pollId: string,
      data: UpdatePollRequest & { status?: PollStatus; is_featured?: boolean },
      authHeaders: Record<string, string>
    ): Promise<Poll> {
      return fetchApi<Poll>(`/admin/polls/${pollId}`, {
        method: 'PUT',
        headers: authHeaders,
        body: data
      })
    },

    async approvePoll(pollId: string, approved: boolean, reason: string | undefined, authHeaders: Record<string, string>): Promise<void> {
      await fetchApi<{ message: string }>(`/admin/polls/${pollId}/approve`, {
        method: 'POST',
        headers: authHeaders,
        body: { approved, reason }
      })
    },

    async closePoll(pollId: string, authHeaders: Record<string, string>): Promise<void> {
      await fetchApi<{ message: string }>(`/admin/polls/${pollId}/close`, {
        method: 'POST',
        headers: authHeaders
      })
    },

    async adminDeletePoll(pollId: string, authHeaders: Record<string, string>): Promise<void> {
      await fetchApi<{ message: string }>(`/admin/polls/${pollId}`, {
        method: 'DELETE',
        headers: authHeaders
      })
    },

    async adminDeletePollComment(commentId: string, authHeaders: Record<string, string>): Promise<void> {
      await fetchApi<{ message: string }>(`/admin/polls/comments/${commentId}`, {
        method: 'DELETE',
        headers: authHeaders
      })
    }
  }
}
