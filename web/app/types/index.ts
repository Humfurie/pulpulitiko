// Article types
export type ArticleStatus = 'draft' | 'published' | 'archived'

export interface Author {
  id: string
  name: string
  slug: string
  bio?: string
  avatar?: string
  email?: string
  created_at: string
  updated_at: string
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
export type UserRole = 'admin' | 'editor'

export interface User {
  id: string
  email: string
  name: string
  role: UserRole
  created_at: string
  updated_at: string
}

export interface LoginRequest {
  email: string
  password: string
}

export interface LoginResponse {
  token: string
  user: User
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
