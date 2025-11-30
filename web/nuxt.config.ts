// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: '2025-07-15',
  devtools: { enabled: true },

  css: ['~/assets/css/main.css'],

  modules: [
    '@nuxt/eslint',
    '@nuxt/hints',
    '@nuxt/image',
    '@nuxt/scripts',
    '@nuxt/test-utils',
    '@nuxt/ui',
    '@nuxtjs/sitemap'
  ],

  runtimeConfig: {
    apiInternalUrl: process.env.NUXT_API_INTERNAL_URL || 'http://api:8080/api',
    public: {
      apiUrl: process.env.NUXT_PUBLIC_API_URL || 'http://localhost:8080/api',
      siteUrl: process.env.NUXT_PUBLIC_SITE_URL || 'https://pulpulitiko.com'
    }
  },

  site: {
    url: process.env.NUXT_PUBLIC_SITE_URL || 'https://pulpulitiko.com',
    name: 'Pulpulitiko'
  },

  sitemap: {
    sources: ['/api/__sitemap__/urls']
  },

  app: {
    head: {
      title: 'Pulpulitiko - Philippine Politics News',
      htmlAttrs: {
        lang: 'en'
      },
      meta: [
        { charset: 'utf-8' },
        { name: 'viewport', content: 'width=device-width, initial-scale=1' },
        { name: 'description', content: 'Your trusted source for Philippine political news and commentary' },
        { property: 'og:site_name', content: 'Pulpulitiko' },
        { property: 'og:type', content: 'website' },
        { name: 'twitter:card', content: 'summary_large_image' }
      ],
      link: [
        { rel: 'icon', type: 'image/x-icon', href: '/favicon.ico' }
      ]
    }
  },

  routeRules: {
    '/': { prerender: true },
    '/article/**': { isr: 900 }, // 15 minutes
    '/category/**': { isr: 1800 }, // 30 minutes
    '/tag/**': { isr: 1800 },
    '/author/**': { isr: 1800 } // 30 minutes
  },

  // Vite config for Docker HMR
  vite: {
    server: {
      watch: {
        usePolling: true
      },
      hmr: {
        clientPort: Number(process.env.NUXT_VITE_HMR_CLIENT_PORT) || undefined
      }
    }
  }
})