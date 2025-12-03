import type {
  ApiResponse,
  Article,
  ArticleListItem,
  Author,
  AuthorWithArticles,
  Category,
  CategoryWithArticles,
  Comment,
  CommentAuthor,
  CommentCountResponse,
  CreateCommentRequest,
  PaginatedArticles,
  Tag,
  TagWithArticles,
  UploadResult,
  UserProfile
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
    }
  }
}
