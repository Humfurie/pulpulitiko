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

export interface Article {
  id: string
  slug: string
  title: string
  summary?: string
  content: string
  featured_image?: string
  author_id?: string
  category_id?: string
  status: ArticleStatus
  view_count: number
  published_at?: string
  created_at: string
  updated_at: string
  author?: Author
  category?: Category
  tags?: Tag[]
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
  status?: ArticleStatus
  tag_ids?: string[]
}

export interface UpdateArticleRequest {
  slug?: string
  title?: string
  summary?: string
  content?: string
  featured_image?: string
  author_id?: string
  category_id?: string
  status?: ArticleStatus
  tag_ids?: string[]
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
