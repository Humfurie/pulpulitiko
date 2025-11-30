import type {
  ApiResponse,
  Article,
  ArticleListItem,
  Category,
  CategoryWithArticles,
  PaginatedArticles,
  Tag,
  TagWithArticles
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
    }
  }
}
