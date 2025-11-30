import { defineSitemapEventHandler } from '#imports'
import type { SitemapUrl } from '#sitemap/types'

interface ArticleListItem {
  slug: string
  published_at?: string
}

interface Category {
  slug: string
}

interface Tag {
  slug: string
}

interface Author {
  slug: string
}

interface ApiResponse<T> {
  success: boolean
  data: T
}

interface PaginatedArticles {
  articles: ArticleListItem[]
}

export default defineSitemapEventHandler(async () => {
  const config = useRuntimeConfig()
  const apiUrl = config.apiInternalUrl || 'http://localhost:8080/api'

  const urls: SitemapUrl[] = []

  try {
    // Fetch articles
    const articlesRes = await $fetch<ApiResponse<PaginatedArticles>>(`${apiUrl}/articles?per_page=1000`)
    if (articlesRes.success && articlesRes.data.articles) {
      for (const article of articlesRes.data.articles) {
        urls.push({
          loc: `/article/${article.slug}`,
          lastmod: article.published_at ? new Date(article.published_at).toISOString() : undefined,
          changefreq: 'weekly',
          priority: 0.8
        })
      }
    }

    // Fetch categories
    const categoriesRes = await $fetch<ApiResponse<Category[]>>(`${apiUrl}/categories`)
    if (categoriesRes.success && categoriesRes.data) {
      for (const category of categoriesRes.data) {
        urls.push({
          loc: `/category/${category.slug}`,
          changefreq: 'daily',
          priority: 0.7
        })
      }
    }

    // Fetch tags
    const tagsRes = await $fetch<ApiResponse<Tag[]>>(`${apiUrl}/tags`)
    if (tagsRes.success && tagsRes.data) {
      for (const tag of tagsRes.data) {
        urls.push({
          loc: `/tag/${tag.slug}`,
          changefreq: 'weekly',
          priority: 0.6
        })
      }
    }

    // Fetch authors
    const authorsRes = await $fetch<ApiResponse<Author[]>>(`${apiUrl}/authors`)
    if (authorsRes.success && authorsRes.data) {
      for (const author of authorsRes.data) {
        urls.push({
          loc: `/author/${author.slug}`,
          changefreq: 'weekly',
          priority: 0.6
        })
      }
    }
  } catch (error) {
    console.error('Error fetching sitemap data:', error)
  }

  return urls
})
