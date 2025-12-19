// @ts-check
import withNuxt from './.nuxt/eslint.config.mjs'

export default withNuxt(
  // Your custom configs here
  {
    files: ['**/*.test.ts', '**/*.spec.ts', 'test/**/*.ts'],
    rules: {
      '@typescript-eslint/no-explicit-any': 'off',
    },
  },
  {
    files: ['**/*.vue', '**/*.ts'],
    rules: {
      'vue/no-v-html': 'warn', // Allow v-html with warning (sanitized content)
    },
  },
)
