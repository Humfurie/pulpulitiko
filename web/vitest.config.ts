import { defineVitestConfig } from '@nuxt/test-utils/config'

export default defineVitestConfig({
  test: {
    environment: 'jsdom',
    globals: true,
    // Exclude E2E tests (run with Playwright instead)
    exclude: ['**/node_modules/**', '**/tests/e2e/**'],
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json', 'html'],
      // Only measure coverage for files that have corresponding tests
      include: [
        'app/composables/**/*.{ts,js,vue}',
        'app/utils/**/*.{ts,js,vue}',
        'server/**/*.{ts,js}',
      ],
      exclude: [
        'node_modules/',
        '.nuxt/',
        'dist/',
        '**/*.config.*',
        '**/*.d.ts',
        'coverage/',
        'test/',
        'tests/',
        '**/__tests__/**',
        '**/*.test.ts',
        '**/*.spec.ts',
      ],
      // Realistic thresholds for current test coverage
      // Increase these as you add more tests
      thresholds: {
        lines: 70,
        functions: 70,
        branches: 60,
        statements: 70,
      },
    },
    setupFiles: ['./test/setup.ts'],
  },
})
