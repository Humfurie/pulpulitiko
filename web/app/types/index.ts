// Article types
export type ArticleStatus = 'draft' | 'published' | 'archived'

// Permission types
export interface Permission {
  id: string
  name: string
  slug: string
  description?: string
  category: string
  created_at: string
}

export interface PermissionCategory {
  category: string
  permissions: Permission[]
}

// Role types
export interface Role {
  id: string
  name: string
  slug: string
  description?: string
  is_system: boolean
  permissions?: Permission[]
  created_at: string
  updated_at: string
  deleted_at?: string
}

export interface RoleWithPermissionCount {
  id: string
  name: string
  slug: string
  description?: string
  is_system: boolean
  permission_count: number
  created_at: string
  updated_at: string
  deleted_at?: string
}

export interface CreateRoleRequest {
  name: string
  slug: string
  description?: string
  permission_ids?: string[]
}

export interface UpdateRoleRequest {
  name?: string
  slug?: string
  description?: string
  permission_ids?: string[]
}

export interface SocialLinks {
  twitter?: string
  facebook?: string
  linkedin?: string
  instagram?: string
  youtube?: string
  tiktok?: string
  website?: string
}

export interface Author {
  id: string
  name: string
  slug: string
  bio?: string
  avatar?: string
  email?: string
  phone?: string
  address?: string
  social_links?: SocialLinks
  role_id?: string
  role: string
  created_at: string
  updated_at: string
  deleted_at?: string
}

export interface Category {
  id: string
  name: string
  slug: string
  description?: string
  created_at: string
  updated_at: string
}

export interface Tag {
  id: string
  name: string
  slug: string
  created_at: string
  updated_at: string
}

// Politician types
export interface Politician {
  id: string
  name: string
  slug: string
  photo?: string
  position?: string
  party?: string
  short_bio?: string
  level?: string
  branch?: string
  position_id?: string
  party_id?: string
  district_id?: string
  created_at: string
  updated_at: string
  deleted_at?: string
  article_count?: number
  party_info?: PartyBrief
  position_info?: GovernmentPositionInfo
}

export interface PoliticianListItem {
  id: string
  name: string
  slug: string
  photo?: string
  position?: string
  party?: string
  level?: string
  branch?: string
  article_count: number
  party_info?: PartyBrief
}

export interface CreatePoliticianRequest {
  name: string
  slug: string
  photo?: string
  position?: string
  party?: string
  short_bio?: string
  level?: string
  branch?: string
  position_id?: string
  party_id?: string
  district_id?: string
}

export interface UpdatePoliticianRequest {
  name?: string
  slug?: string
  photo?: string
  position?: string
  party?: string
  short_bio?: string
  level?: string
  branch?: string
  position_id?: string
  party_id?: string
  district_id?: string
}

export interface PaginatedPoliticians {
  politicians: PoliticianListItem[]
  total: number
  page: number
  per_page: number
  total_pages: number
}

export interface PoliticianWithArticles {
  politician: Politician
  articles: PaginatedArticles
}

export interface Article {
  id: string
  slug: string
  title: string
  summary?: string
  content: string
  featured_image?: string
  author_id?: string
  category_id?: string
  primary_politician_id?: string
  status: ArticleStatus
  view_count: number
  published_at?: string
  created_at: string
  updated_at: string
  author?: Author
  category?: Category
  tags?: Tag[]
  primary_politician?: Politician
  mentioned_politicians?: Politician[]
}

export interface ArticleListItem {
  id: string
  slug: string
  title: string
  summary?: string
  featured_image?: string
  status: ArticleStatus
  view_count: number
  published_at?: string
  created_at: string
  author_name?: string
  category_name?: string
  category_slug?: string
  primary_politician_name?: string
  primary_politician_slug?: string
}

export interface PaginatedArticles {
  articles: ArticleListItem[]
  total: number
  page: number
  per_page: number
  total_pages: number
}

export interface CategoryWithArticles {
  category: Category
  articles: PaginatedArticles
}

export interface TagWithArticles {
  tag: Tag
  articles: PaginatedArticles
}

export interface AuthorWithArticles {
  author: Author
  articles: PaginatedArticles
}

// API Response types
export interface ApiResponse<T> {
  success: boolean
  data: T
  message?: string
}

export interface ApiError {
  success: false
  error: string
  message?: string
}

// Auth types
export interface User {
  id: string
  email: string
  name: string
  avatar?: string
  role_id?: string
  role: string
  created_at: string
  updated_at: string
  deleted_at?: string
}

export interface LoginRequest {
  email: string
  password: string
}

export interface LoginResponse {
  token: string
  user: User
  permissions: string[]
}

export interface CreateArticleRequest {
  slug: string
  title: string
  summary?: string
  content: string
  featured_image?: string
  author_id?: string
  category_id?: string
  primary_politician_id?: string
  status?: ArticleStatus
  tag_ids?: string[]
  politician_ids?: string[]
}

export interface UpdateArticleRequest {
  slug?: string
  title?: string
  summary?: string
  content?: string
  featured_image?: string
  author_id?: string
  category_id?: string
  primary_politician_id?: string
  status?: ArticleStatus
  tag_ids?: string[]
  politician_ids?: string[]
}

export interface PaginatedCategories {
  categories: Category[]
  total: number
  page: number
  per_page: number
  total_pages: number
}

export interface CreateCategoryRequest {
  name: string
  slug: string
  description?: string
}

export interface UpdateCategoryRequest {
  name?: string
  slug?: string
  description?: string
}

export interface PaginatedTags {
  tags: Tag[]
  total: number
  page: number
  per_page: number
  total_pages: number
}

export interface CreateTagRequest {
  name: string
  slug: string
}

export interface UpdateTagRequest {
  name?: string
  slug?: string
}

// Upload types
export interface UploadResult {
  key: string
  url: string
  size: number
  mime_type: string
}

// Author/User management types
export interface CreateAuthorRequest {
  name: string
  slug: string
  bio?: string
  avatar?: string
  email?: string
  phone?: string
  address?: string
  social_links?: SocialLinks
  role_id?: string
  role?: string
}

export interface UpdateAuthorRequest {
  name?: string
  slug?: string
  bio?: string
  avatar?: string
  email?: string
  phone?: string
  address?: string
  social_links?: SocialLinks
  role_id?: string
  role?: string
}

// Comment types
export type CommentStatus = 'active' | 'under_review' | 'spam' | 'hidden'

export interface CommentAuthor {
  id: string
  name: string
  avatar?: string
  is_system?: boolean // True for verified/staff users
}

export interface ReactionSummary {
  reaction: string
  count: number
  has_reacted: boolean
}

export interface Comment {
  id: string
  article_id: string
  user_id: string
  parent_id?: string
  content: string
  status: CommentStatus
  created_at: string
  updated_at: string
  deleted_at?: string
  // Moderation fields
  moderated_by?: string
  moderated_at?: string
  moderation_reason?: string
  // Relations
  author?: CommentAuthor
  reactions?: ReactionSummary[]
  reply_count?: number
  // For user profile page - article context
  article_slug?: string
  article_title?: string
}

export interface CreateCommentRequest {
  content: string
  parent_id?: string
}

export interface UpdateCommentRequest {
  content: string
}

export interface AddReactionRequest {
  reaction: string
}

export interface ModerateCommentRequest {
  status: CommentStatus
  reason?: string
}

export interface CommentCountResponse {
  count: number
}

// User profile types
export interface UserProfile {
  id: string
  name: string
  slug: string
  avatar?: string
  created_at: string
  comment_count: number
  reply_count: number
}

// Metrics types
export interface CategoryMetric {
  id: string
  name: string
  slug: string
  article_count: number
  total_views: number
}

export interface TagMetric {
  id: string
  name: string
  slug: string
  article_count: number
  total_views: number
}

export interface TopArticle {
  id: string
  slug: string
  title: string
  view_count: number
  category_name?: string
}

export interface DashboardMetrics {
  total_articles: number
  total_views: number
  total_categories: number
  total_tags: number
  top_articles: TopArticle[]
  category_metrics: CategoryMetric[]
  tag_metrics: TagMetric[]
}

// Messaging types
export type ConversationStatus = 'open' | 'closed' | 'archived'

export interface Conversation {
  id: string
  user_id: string
  user?: User
  subject?: string
  status: ConversationStatus
  last_message_at?: string
  last_message?: Message
  unread_count?: number
  created_at: string
  updated_at: string
}

export interface Message {
  id: string
  conversation_id: string
  sender_id: string
  sender?: User
  content: string
  is_read: boolean
  read_at?: string
  created_at: string
}

export interface CreateConversationRequest {
  subject?: string
  message: string
}

export interface CreateMessageRequest {
  content: string
}

export interface UpdateConversationRequest {
  status: ConversationStatus
}

export interface PaginatedConversations {
  conversations: Conversation[]
  total: number
  page: number
  per_page: number
  total_pages: number
}

export interface PaginatedMessages {
  messages: Message[]
  total: number
  page: number
  per_page: number
  total_pages: number
}

export interface UnreadCounts {
  total: number
  conversations: number
}

// WebSocket types
export type WSMessageType =
  | 'new_message'
  | 'message_read'
  | 'typing'
  | 'stop_typing'
  | 'user_online'
  | 'user_offline'
  | 'conversation_update'

export interface WSMessage {
  type: WSMessageType
  conversation_id?: string
  message?: Message
  user_id?: string
  timestamp: string
}

// Search Analytics types
export type TimeRange = '1h' | '1d' | '1w' | '1m' | '1y' | '5y' | 'lifetime'

export interface TrackSearchRequest {
  query: string
  session_id?: string
  results_count: number
}

export interface TrackSearchResponse {
  search_query_id: string
}

export interface TrackClickRequest {
  search_query_id: string
  article_id: string
  position: number
}

export interface TopSearchTerm {
  query: string
  count: number
  click_count: number
  ctr: number
}

export interface SearchTrend {
  period: string
  search_count: number
  unique_terms: number
  click_count: number
  unique_clicks: number
}

export interface PoliticianSearchStats {
  politician_id: string
  politician_name: string
  politician_slug: string
  search_count: number
}

export interface TopClickedArticle {
  article_id: string
  article_title: string
  article_slug: string
  click_count: number
  avg_position: number
}

export interface SearchAnalytics {
  time_range: TimeRange
  total_searches: number
  unique_search_terms: number
  total_clicks: number
  overall_ctr: number
  top_search_terms: TopSearchTerm[]
  search_trends: SearchTrend[]
  politician_searches: PoliticianSearchStats[]
  top_clicked_articles: TopClickedArticle[]
}

// Politician Comment types
export interface PoliticianComment {
  id: string
  politician_id: string
  user_id: string
  parent_id?: string
  content: string
  status: CommentStatus
  created_at: string
  updated_at: string
  deleted_at?: string
  moderated_by?: string
  moderated_at?: string
  moderation_reason?: string
  author?: CommentAuthor
  reactions?: ReactionSummary[]
  reply_count?: number
  politician_slug?: string
  politician_name?: string
}

export interface PaginatedPoliticianComments {
  comments: PoliticianComment[]
  total: number
  page: number
  per_page: number
  total_pages: number
}

// Notification types
export type NotificationType =
  | 'mention_article_comment'
  | 'mention_politician_comment'
  | 'reply_article_comment'
  | 'reply_politician_comment'
  | 'comment_reaction'

export interface NotificationActor {
  id: string
  name: string
  avatar?: string
}

export interface NotificationRef {
  id: string
  name: string
  slug: string
}

export interface Notification {
  id: string
  user_id: string
  type: NotificationType
  title: string
  message?: string
  actor_id?: string
  article_id?: string
  politician_id?: string
  comment_id?: string
  is_read: boolean
  read_at?: string
  created_at: string
  actor?: NotificationActor
  article?: NotificationRef
  politician?: NotificationRef
}

export interface PaginatedNotifications {
  notifications: Notification[]
  total: number
  unread_count: number
  page: number
  per_page: number
  total_pages: number
}

// =====================================================
// LOCATION TYPES (Philippine Geographic Hierarchy)
// =====================================================

export interface Region {
  id: string
  code: string
  name: string
  slug: string
  created_at: string
  updated_at: string
  deleted_at?: string
  provinces?: Province[]
  province_count?: number
}

export interface RegionListItem {
  id: string
  code: string
  name: string
  slug: string
  province_count: number
}

export interface Province {
  id: string
  region_id: string
  code: string
  name: string
  slug: string
  created_at: string
  updated_at: string
  deleted_at?: string
  region?: Region
  cities_municipalities?: CityMunicipality[]
  city_count?: number
}

export interface ProvinceListItem {
  id: string
  region_id: string
  code: string
  name: string
  slug: string
  region_name?: string
  city_count: number
}

export interface CityMunicipality {
  id: string
  province_id: string
  code: string
  name: string
  slug: string
  is_city: boolean
  is_capital: boolean
  is_huc: boolean // Highly Urbanized City
  is_icc: boolean // Independent Component City
  population?: number
  created_at: string
  updated_at: string
  deleted_at?: string
  province?: Province
  barangays?: Barangay[]
  barangay_count?: number
}

export interface CityMunicipalityListItem {
  id: string
  province_id: string
  code: string
  name: string
  slug: string
  is_city: boolean
  is_capital: boolean
  is_huc: boolean
  province_name?: string
  barangay_count: number
}

export interface Barangay {
  id: string
  city_municipality_id: string
  code: string
  name: string
  slug: string
  population?: number
  created_at: string
  updated_at: string
  deleted_at?: string
  city_municipality?: CityMunicipality
}

export interface BarangayListItem {
  id: string
  city_municipality_id: string
  code: string
  name: string
  slug: string
  city_municipality_name?: string
}

export interface CongressionalDistrict {
  id: string
  province_id?: string
  city_municipality_id?: string
  district_number: number
  name: string
  slug: string
  created_at: string
  updated_at: string
  deleted_at?: string
  province?: Province
  city_municipality?: CityMunicipality
  coverage?: CityMunicipality[]
}

export interface DistrictListItem {
  id: string
  district_number: number
  name: string
  slug: string
  province_name?: string
  city_name?: string
}

// Location hierarchy (full path)
export interface LocationHierarchy {
  region?: RegionListItem
  province?: ProvinceListItem
  city_municipality?: CityMunicipalityListItem
  barangay?: BarangayListItem
  district?: DistrictListItem
}

// Location search result
export interface LocationSearchResult {
  type: 'region' | 'province' | 'city' | 'barangay'
  id: string
  code: string
  name: string
  slug: string
  parent_name?: string
  full_path: string
}

// Paginated responses
export interface PaginatedBarangays {
  barangays: BarangayListItem[]
  total: number
  page: number
  per_page: number
  total_pages: number
}

export interface PaginatedDistricts {
  districts: DistrictListItem[]
  total: number
  page: number
  per_page: number
  total_pages: number
}

// Request types
export interface CreateRegionRequest {
  code: string
  name: string
  slug: string
}

export interface UpdateRegionRequest {
  code?: string
  name?: string
  slug?: string
}

export interface CreateProvinceRequest {
  region_id: string
  code: string
  name: string
  slug: string
}

export interface UpdateProvinceRequest {
  region_id?: string
  code?: string
  name?: string
  slug?: string
}

export interface CreateCityMunicipalityRequest {
  province_id: string
  code: string
  name: string
  slug: string
  is_city?: boolean
  is_capital?: boolean
  is_huc?: boolean
  is_icc?: boolean
  population?: number
}

export interface UpdateCityMunicipalityRequest {
  province_id?: string
  code?: string
  name?: string
  slug?: string
  is_city?: boolean
  is_capital?: boolean
  is_huc?: boolean
  is_icc?: boolean
  population?: number
}

export interface CreateBarangayRequest {
  city_municipality_id: string
  code: string
  name: string
  slug: string
  population?: number
}

export interface UpdateBarangayRequest {
  city_municipality_id?: string
  code?: string
  name?: string
  slug?: string
  population?: number
}

export interface CreateDistrictRequest {
  province_id?: string
  city_municipality_id?: string
  district_number: number
  name: string
  slug: string
}

// Response types for combined data
export interface RegionWithProvinces {
  region: Region
  provinces: ProvinceListItem[]
}

export interface ProvinceWithCities {
  province: Province
  cities: CityMunicipalityListItem[]
}

export interface CityWithBarangays {
  city: CityMunicipality
  barangays: PaginatedBarangays
}

// =====================================================
// GOVERNMENT STRUCTURE TYPES
// =====================================================

export type GovernmentLevel = 'national' | 'regional' | 'provincial' | 'city' | 'municipal' | 'barangay'
export type GovernmentBranch = 'executive' | 'legislative' | 'judicial'

// Political Party
export interface PoliticalParty {
  id: string
  name: string
  slug: string
  abbreviation?: string
  logo?: string
  color?: string
  description?: string
  founded_year?: number
  website?: string
  is_major: boolean
  is_active: boolean
  created_at: string
  updated_at: string
  deleted_at?: string
  member_count?: number
}

export interface PoliticalPartyListItem {
  id: string
  name: string
  slug: string
  abbreviation?: string
  logo?: string
  color?: string
  is_major: boolean
  is_active: boolean
  member_count: number
}

export interface PartyBrief {
  id: string
  name: string
  slug: string
  abbreviation?: string
  logo?: string
  color?: string
}

export interface CreatePoliticalPartyRequest {
  name: string
  slug: string
  abbreviation?: string
  logo?: string
  color?: string
  description?: string
  founded_year?: number
  website?: string
  is_major: boolean
  is_active: boolean
}

export interface UpdatePoliticalPartyRequest {
  name?: string
  slug?: string
  abbreviation?: string
  logo?: string
  color?: string
  description?: string
  founded_year?: number
  website?: string
  is_major?: boolean
  is_active?: boolean
}

export interface PaginatedPoliticalParties {
  parties: PoliticalPartyListItem[]
  total: number
  page: number
  per_page: number
  total_pages: number
}

// Government Position
export interface GovernmentPosition {
  id: string
  name: string
  slug: string
  level: GovernmentLevel
  branch: GovernmentBranch
  display_order: number
  description?: string
  max_terms?: number
  term_years: number
  is_elected: boolean
  created_at: string
  updated_at: string
}

export interface GovernmentPositionListItem {
  id: string
  name: string
  slug: string
  level: GovernmentLevel
  branch: GovernmentBranch
  display_order: number
  is_elected: boolean
}

export interface GovernmentPositionInfo {
  id: string
  name: string
  slug: string
  level: GovernmentLevel
  branch: GovernmentBranch
  is_elected: boolean
}

// Politician Jurisdiction
export interface PoliticianJurisdiction {
  id: string
  politician_id: string
  region_id?: string
  province_id?: string
  city_id?: string
  barangay_id?: string
  is_national: boolean
  created_at: string
  region?: RegionListItem
  province?: ProvinceListItem
  city?: CityMunicipalityListItem
  barangay?: BarangayListItem
}

export interface CreatePoliticianJurisdictionRequest {
  politician_id: string
  region_id?: string
  province_id?: string
  city_id?: string
  barangay_id?: string
  is_national: boolean
}

// =====================================================
// LEGISLATION / BILLS TRACKER TYPES
// =====================================================

// Bill Status constants
export type BillStatus =
  | 'filed'
  | 'pending_committee'
  | 'in_committee'
  | 'reported_out'
  | 'pending_second_reading'
  | 'approved_second_reading'
  | 'pending_third_reading'
  | 'approved_third_reading'
  | 'transmitted'
  | 'consolidated'
  | 'ratified'
  | 'signed_into_law'
  | 'vetoed'
  | 'lapsed'
  | 'withdrawn'
  | 'archived'

// Legislative Chamber
export type LegislativeChamber = 'senate' | 'house'

// Vote Type
export type VoteType = 'yea' | 'nay' | 'abstain' | 'absent'

// Legislative Session
export interface LegislativeSession {
  id: string
  congress_number: number
  session_number: number
  session_type: string
  start_date: string
  end_date?: string
  is_current: boolean
  created_at: string
  updated_at: string
}

export interface LegislativeSessionListItem {
  id: string
  congress_number: number
  session_number: number
  session_type: string
  is_current: boolean
  bill_count: number
}

// Committee
export interface Committee {
  id: string
  chamber: LegislativeChamber
  name: string
  slug: string
  description?: string
  chairperson_id?: string
  vice_chairperson_id?: string
  is_active: boolean
  created_at: string
  updated_at: string
  deleted_at?: string
  chairperson?: PoliticianListItem
  vice_chairperson?: PoliticianListItem
}

export interface CommitteeListItem {
  id: string
  chamber: LegislativeChamber
  name: string
  slug: string
  is_active: boolean
  bill_count: number
}

// Bill
export interface Bill {
  id: string
  session_id: string
  chamber: LegislativeChamber
  bill_number: string
  title: string
  slug: string
  short_title?: string
  summary?: string
  full_text?: string
  significance?: string
  status: BillStatus
  filed_date: string
  last_action_date?: string
  date_signed?: string
  republic_act_number?: string
  created_at: string
  updated_at: string
  deleted_at?: string
  // Joined fields
  session?: LegislativeSessionListItem
  authors?: BillAuthor[]
  principal_authors?: PoliticianListItem[]
  committees?: BillCommittee[]
  status_history?: BillStatusHistoryItem[]
  topics?: BillTopic[]
  votes?: BillVote[]
}

export interface BillListItem {
  id: string
  chamber: LegislativeChamber
  bill_number: string
  title: string
  slug: string
  short_title?: string
  status: BillStatus
  filed_date: string
  last_action_date?: string
  author_count: number
  topic_names?: string[]
}

// Bill Author
export interface BillAuthor {
  id: string
  bill_id: string
  politician_id: string
  is_principal_author: boolean
  created_at: string
  politician?: PoliticianListItem
}

// Bill Status History
export interface BillStatusHistoryItem {
  id: string
  bill_id: string
  status: BillStatus
  action_description?: string
  action_date: string
  created_at: string
}

// Bill Committee
export interface BillCommittee {
  id: string
  bill_id: string
  committee_id: string
  referred_date: string
  is_primary: boolean
  status: string
  created_at: string
  committee?: CommitteeListItem
}

// Bill Vote (aggregate for a voting session)
export interface BillVote {
  id: string
  bill_id: string
  chamber: LegislativeChamber
  reading: string
  vote_date: string
  yeas: number
  nays: number
  abstentions: number
  absent: number
  is_passed: boolean
  notes?: string
  created_at: string
}

// Individual Politician Vote
export interface PoliticianVote {
  id: string
  bill_vote_id: string
  politician_id: string
  vote: VoteType
  created_at: string
  politician?: PoliticianListItem
}

// Bill Topic
export interface BillTopic {
  id: string
  name: string
  slug: string
  description?: string
  created_at: string
  bill_count?: number
}

// Politician Voting Record Summary
export interface PoliticianVotingRecord {
  politician_id: string
  total_votes: number
  yea_votes: number
  nay_votes: number
  abstain_votes: number
  absent_votes: number
  attendance_rate: number
}

// Politician Bill Vote (for voting history)
export interface PoliticianBillVote {
  bill: BillListItem
  vote: VoteType
  vote_date: string
  reading: string
  bill_passed: boolean
}

// Paginated types
export interface PaginatedBills {
  bills: BillListItem[]
  total: number
  page: number
  per_page: number
  total_pages: number
}

export interface PaginatedPoliticianVotes {
  votes: PoliticianBillVote[]
  total: number
  page: number
  per_page: number
  total_pages: number
}

// Request types
export interface CreateBillRequest {
  session_id: string
  chamber: LegislativeChamber
  bill_number: string
  title: string
  slug: string
  short_title?: string
  summary?: string
  full_text?: string
  significance?: string
  status: BillStatus
  filed_date: string // YYYY-MM-DD
  principal_authors?: string[]
  co_authors?: string[]
  topic_ids?: string[]
}

export interface UpdateBillRequest {
  title?: string
  slug?: string
  short_title?: string
  summary?: string
  full_text?: string
  significance?: string
  status?: BillStatus
  last_action_date?: string // YYYY-MM-DD
  date_signed?: string // YYYY-MM-DD
  republic_act_number?: string
  topic_ids?: string[]
}

export interface AddBillStatusRequest {
  status: BillStatus
  action_description?: string
  action_date: string // YYYY-MM-DD
}

export interface AddBillVoteRequest {
  chamber: LegislativeChamber
  reading: 'second' | 'third'
  vote_date: string // YYYY-MM-DD
  yeas: number
  nays: number
  abstentions: number
  absent: number
  is_passed: boolean
  notes?: string
}

export interface AddPoliticianVoteRequest {
  politician_id: string
  vote: VoteType
}

// Filter type for bill queries
export interface BillFilter {
  chamber?: LegislativeChamber
  status?: BillStatus
  session_id?: string
  topic_id?: string
  author_id?: string
  search?: string
}
