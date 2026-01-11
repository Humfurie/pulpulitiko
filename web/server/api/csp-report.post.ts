export default defineEventHandler(async (event) => {
  try {
    const body = await readBody(event)

    // Log CSP violations for monitoring
    console.warn('CSP Violation Report:', {
      timestamp: new Date().toISOString(),
      'document-uri': body['csp-report']?.['document-uri'],
      'violated-directive': body['csp-report']?.['violated-directive'],
      'blocked-uri': body['csp-report']?.['blocked-uri'],
      'source-file': body['csp-report']?.['source-file'],
      'line-number': body['csp-report']?.['line-number'],
      'original-policy': body['csp-report']?.['original-policy']
    })

    // TODO: In production, send to monitoring service (Sentry, Datadog, etc.)
    // Example with Sentry:
    // Sentry.captureMessage('CSP Violation', {
    //   level: 'warning',
    //   extra: body['csp-report']
    // })

    // Return success
    return { status: 'ok', received: true }
  } catch (error) {
    console.error('Error processing CSP report:', error)
    return { status: 'error', message: 'Failed to process CSP report' }
  }
})
