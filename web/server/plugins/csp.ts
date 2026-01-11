/**
 * Content Security Policy (CSP) Plugin
 *
 * Implements defense-in-depth by setting CSP headers to prevent XSS, clickjacking,
 * and other injection attacks.
 *
 * Environment Variables:
 * - NUXT_PUBLIC_CSP_ENFORCE: Set to 'true' to enable enforcement mode (default: report-only)
 *
 * Usage:
 * 1. Deploy with report-only mode (default)
 * 2. Monitor CSP reports at /api/csp-report
 * 3. Fix any violations
 * 4. Set NUXT_PUBLIC_CSP_ENFORCE=true to enable enforcement
 *
 * HTTPS Enforcement:
 * - Production: upgrade-insecure-requests directive is enabled
 * - Development: HTTP is allowed for local development convenience
 * - For production-like testing, use HTTPS with self-signed certificates
 *   or configure Traefik/nginx with SSL termination
 *
 * SECURITY: Always test in report-only mode before enabling enforcement to avoid
 * breaking legitimate functionality.
 */
export default defineNitroPlugin((nitroApp) => {
  nitroApp.hooks.hook('render:response', (response) => {
    const isDev = import.meta.dev
    const config = useRuntimeConfig()

    // Check if CSP enforcement is enabled (default: report-only for safety)
    const cspEnforce = config.public.cspEnforce === 'true' || config.public.cspEnforce === true

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

    // Choose header based on enforcement mode
    if (isDev) {
      // Development: always report-only
      response.headers!['Content-Security-Policy-Report-Only'] = cspHeader
    } else if (cspEnforce) {
      // Production with enforcement enabled
      response.headers!['Content-Security-Policy'] = cspHeader
      console.info('CSP: Enforcement mode ENABLED')
    } else {
      // Production with report-only (default)
      response.headers!['Content-Security-Policy-Report-Only'] = cspHeader
      console.info('CSP: Report-only mode (set NUXT_PUBLIC_CSP_ENFORCE=true to enforce)')
    }

    // Additional security headers
    response.headers!['X-Content-Type-Options'] = 'nosniff'
    response.headers!['X-Frame-Options'] = 'DENY'
    response.headers!['X-XSS-Protection'] = '1; mode=block'
    response.headers!['Referrer-Policy'] = 'strict-origin-when-cross-origin'
    response.headers!['Permissions-Policy'] = 'geolocation=(), microphone=(), camera=()'
  })
})
