/// <reference types="vitest" />

// Vitest global types for IDE support
import type { TestAPI } from 'vitest'

declare global {
  const describe: TestAPI['describe']
  const it: TestAPI['it']
  const test: TestAPI['test']
  const expect: TestAPI['expect']
  const beforeEach: TestAPI['beforeEach']
  const afterEach: TestAPI['afterEach']
  const beforeAll: TestAPI['beforeAll']
  const afterAll: TestAPI['afterAll']
  const vi: TestAPI['vi']
}

export {}
