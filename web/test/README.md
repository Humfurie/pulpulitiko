# Testing Setup

This directory contains the test setup and utilities for the Pulpulitiko frontend application.

## Setup Files

### setup.ts

Global test setup file that configures:
- localStorage mock implementation
- Console spy setup to suppress warnings during tests
- Automatic cleanup between tests

This file is automatically loaded before all tests via the vitest configuration.

## Test Structure

Tests are located alongside the source files they test, following the pattern:
- Source file: `app/composables/useGrouping.ts`
- Test file: `app/composables/useGrouping.test.ts`

## Running Tests

Available npm scripts:

- `npm run test` - Run tests in watch mode (interactive)
- `npm run test:ui` - Run tests with Vitest UI (visual interface)
- `npm run test:run` - Run tests once (CI mode)
- `npm run test:coverage` - Run tests and generate coverage report

## Coverage Thresholds

The project is configured with the following coverage thresholds:
- Lines: 90%
- Functions: 90%
- Branches: 90%
- Statements: 90%

Coverage reports are generated in the `/coverage` directory.

## Testing Utilities

### localStorage Mock

The localStorage mock provides a complete implementation of the Web Storage API for testing:
- `getItem(key)` - Retrieve stored values
- `setItem(key, value)` - Store values
- `removeItem(key)` - Remove values
- `clear()` - Clear all stored values
- `length` - Number of stored items
- `key(index)` - Get key at index

The mock is automatically reset before each test to ensure test isolation.

## Writing Tests

### Example Test Structure

```typescript
import { describe, it, expect, beforeEach } from 'vitest'
import { ref, nextTick } from 'vue'
import { mount } from '@vue/test-utils'

describe('MyComposable', () => {
  beforeEach(() => {
    // Setup code
  })

  it('should do something', async () => {
    // Test implementation
    await nextTick()
    expect(result).toBe(expected)
  })
})
```

### Testing Composables with Components

When testing composables that use lifecycle hooks (onMounted, watch, etc.), wrap them in a component:

```typescript
let composableResult: any

const TestComponent = {
  setup() {
    composableResult = useMyComposable()
    return composableResult
  },
  template: '<div></div>',
}

const wrapper = mount(TestComponent)
await nextTick()

// Access composable values via composableResult
expect(composableResult.someValue.value).toBe(expected)

wrapper.unmount()
```

## Configuration

The vitest configuration is in `/vitest.config.ts` and integrates with Nuxt via `@nuxt/test-utils/config`.

Key configuration:
- Environment: happy-dom (lightweight DOM implementation)
- Globals: Enabled (no need to import describe, it, expect)
- Coverage Provider: v8
- Setup File: ./test/setup.ts

## Best Practices

1. Always use `await nextTick()` after reactive changes
2. Clean up components with `wrapper.unmount()` after tests
3. Test edge cases and error handling
4. Use descriptive test names that explain what is being tested
5. Group related tests with `describe` blocks
6. Mock external dependencies appropriately
7. Ensure tests are deterministic and can run in isolation
8. Test TypeScript types when relevant
9. Include integration tests to verify component usage
10. Aim for high coverage while focusing on meaningful tests
