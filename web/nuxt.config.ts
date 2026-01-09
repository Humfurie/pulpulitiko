declare const process: { env: Record<string, string | undefined> }

// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: '2025-07-15',
  devtools: { enabled: true },

  css: ['~/assets/css/main.css'],

  modules: [
    '@nuxt/eslint',
    '@nuxt/fonts',
    '@nuxt/hints',
    '@nuxt/image',
    '@nuxt/scripts',
    '@nuxt/test-utils',
    '@nuxt/ui',
    '@nuxtjs/sitemap'
  ],

  fonts: {
    families: [
      { name: 'Plus Jakarta Sans', provider: 'google', weights: [400, 500, 600, 700, 800] }
    ],
    defaults: {
      weights: [400, 500, 600, 700, 800]
    }
  },

  icon: {
    clientBundle: {
      scan: true,
      includeCustomCollections: true
    },
    serverBundle: 'local'
  },

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
    pageTransition: { name: 'page', mode: 'out-in' },
    layoutTransition: { name: 'layout', mode: 'out-in' },
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
        { rel: 'icon', type: 'image/png', href: '/pulpulitiko.png' },
        { rel: 'apple-touch-icon', href: '/pulpulitiko.png' },
        { rel: 'alternate', type: 'application/rss+xml', title: 'Pulpulitiko RSS Feed', href: '/rss' }
      ],
      style: [
        {
          innerHTML: `
            html { visibility: hidden; }
            html.ready { visibility: visible; }
            #__nuxt_loader {
              position: fixed;
              inset: 0;
              z-index: 99999;
              display: flex;
              align-items: center;
              justify-content: center;
              background: #fafaf9;
              visibility: visible !important;
            }
            @media (prefers-color-scheme: dark) {
              #__nuxt_loader { background: #0c0a09; }
              #__nuxt_loader .logo-dark { color: white; }
            }
            #__nuxt_loader .logo-dark { color: #1c1917; }
            #__nuxt_loader .logo-accent { color: #f97316; }
            #__nuxt_loader .spinner {
              width: 32px;
              height: 32px;
              border: 3px solid #f97316;
              border-top-color: transparent;
              border-radius: 50%;
              animation: spin 1s linear infinite;
            }
            @keyframes spin { to { transform: rotate(360deg); } }
            @keyframes pulse { 0%, 100% { opacity: 1; } 50% { opacity: 0.5; } }
            #__nuxt_loader .pulse { animation: pulse 2s ease-in-out infinite; }
            #__nuxt_loader.fade-out { opacity: 0; transition: opacity 0.3s ease-out; }
          `
        }
      ]
    }
  },

  routeRules: {
    '/': { ssr: true },
    '/article/**': { isr: 900 }, // 15 minutes
    '/category/**': { isr: 1800 }, // 30 minutes
    '/tag/**': { isr: 1800 },
    '/author/**': { isr: 1800 } // 30 minutes
  },

  nitro: {
    prerender: {
      crawlLinks: false,
      routes: []
    }
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