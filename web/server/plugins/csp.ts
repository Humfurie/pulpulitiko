export default defineNitroPlugin((nitroApp) => {
  nitroApp.hooks.hook('render:response', (response) => {
    // Only apply CSP in non-development environments or when explicitly enabled
    const isDev = process.dev || process.env.NODE_ENV === 'development'
    const config = useRuntimeConfig()

    // CSP directives
    const cspDirectives = [
      // Default fallback
      "default-src 'self'",

      // Images from self, MinIO, and HTTPS sources
      `img-src 'self' ${config.public.minioEndpoint || 'https://minio.humfurie.org'} https: data:`,

      // Scripts - in dev we need unsafe-eval for HMR
      isDev
        ? "script-src 'self' 'unsafe-inline' 'unsafe-eval'"
        : "script-src 'self' 'unsafe-inline'",

      // Styles - unsafe-inline needed for Vue/Nuxt
      "style-src 'self' 'unsafe-inline' https://fonts.googleapis.com",

      // API and WebSocket connections
      `connect-src 'self' ${config.public.minioEndpoint || 'https://minio.humfurie.org'} ${isDev ? 'ws: wss:' : ''}`,

      // Fonts
      "font-src 'self' https://fonts.gstatic.com data:",

      // Frames (if needed for embeds)
      "frame-src 'self'",

      // Objects (disable by default)
      "object-src 'none'",

      // Base URI
      "base-uri 'self'",

      // Form actions
      "form-action 'self'",

      // Frame ancestors (prevent clickjacking)
      "frame-ancestors 'none'",

      // Upgrade insecure requests in production
      ...(!isDev ? ["upgrade-insecure-requests"] : []),

      // CSP violation reporting
      "report-uri /api/csp-report"
    ]

    // Join directives
    const cspHeader = cspDirectives.join('; ')

    // Use report-only mode for safety (can be changed to enforcement later)
    if (isDev) {
      response.headers!['Content-Security-Policy-Report-Only'] = cspHeader
    } else {
      // TODO: After testing in production, switch to enforcement mode:
      // response.headers['Content-Security-Policy'] = cspHeader
      response.headers!['Content-Security-Policy-Report-Only'] = cspHeader
    }

    // Additional security headers
    response.headers!['X-Content-Type-Options'] = 'nosniff'
    response.headers!['X-Frame-Options'] = 'DENY'
    response.headers!['X-XSS-Protection'] = '1; mode=block'
    response.headers!['Referrer-Policy'] = 'strict-origin-when-cross-origin'
    response.headers!['Permissions-Policy'] = 'geolocation=(), microphone=(), camera=()'
  })
})
