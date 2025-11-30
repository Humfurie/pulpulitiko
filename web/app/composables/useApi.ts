import type {
  ApiResponse,
  Article,
  ArticleListItem,
  Author,
  AuthorWithArticles,
  Category,
  CategoryWithArticles,
  PaginatedArticles,
  Tag,
  TagWithArticles,
  UploadResult
} from '~/types'

export function useApi() {
  const config = useRuntimeConfig()
  // Use internal Docker URL for SSR, public URL for client-side
  const baseUrl = import.meta.server
    ? config.apiInternalUrl
    : config.public.apiUrl

  async function fetchApi<T>(endpoint: string, options?: RequestInit): Promise<T> {
    const response = await $fetch<ApiResponse<T>>(`${baseUrl}${endpoint}`, {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...options?.headers
      }
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
    }
  }
}
