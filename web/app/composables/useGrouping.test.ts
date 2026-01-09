import { describe, it, expect, beforeEach, vi } from 'vitest'
import { ref, nextTick } from 'vue'
import { mount } from '@vue/test-utils'
import { useGrouping } from './useGrouping'

// Test data types
interface TestPerson {
  id: number
  name: string
  department: string
  role: string
  age: number
}

interface TestProduct {
  id: string
  name: string
  category: string
  price: number
}

describe('useGrouping', () => {
  // Sample test data
  const testPeople: TestPerson[] = [
    { id: 1, name: 'Alice', department: 'Engineering', role: 'Developer', age: 30 },
    { id: 2, name: 'Bob', department: 'Engineering', role: 'Manager', age: 35 },
    { id: 3, name: 'Charlie', department: 'Sales', role: 'Sales Rep', age: 28 },
    { id: 4, name: 'Diana', department: 'Sales', role: 'Manager', age: 40 },
    { id: 5, name: 'Eve', department: 'Marketing', role: 'Designer', age: 27 },
  ]

  const testProducts: TestProduct[] = [
    { id: 'p1', name: 'Laptop', category: 'Electronics', price: 1200 },
    { id: 'p2', name: 'Mouse', category: 'Electronics', price: 25 },
    { id: 'p3', name: 'Desk', category: 'Furniture', price: 500 },
    { id: 'p4', name: 'Chair', category: 'Furniture', price: 300 },
    { id: 'p5', name: 'Monitor', category: 'Electronics', price: 400 },
  ]

  beforeEach(() => {
    // Clear localStorage before each test
    localStorage.clear()
    vi.clearAllMocks()
  })

  describe('Basic Grouping', () => {
    it('should group items by specified field', async () => {
      const items = ref(testPeople)
      const groupBy = ref('department')
      const getGroupKey = (item: TestPerson, field: string) => item[field as keyof TestPerson] as string

      const { groupedItems } = useGrouping(items, groupBy, getGroupKey)

      await nextTick()

      expect(Object.keys(groupedItems.value)).toHaveLength(3)
      expect(groupedItems.value.Engineering).toHaveLength(2)
      expect(groupedItems.value.Sales).toHaveLength(2)
      expect(groupedItems.value.Marketing).toHaveLength(1)
    })

    it('should return all items in empty group when groupBy is empty', async () => {
      const items = ref(testPeople)
      const groupBy = ref('')
      const getGroupKey = (item: TestPerson, field: string) => item[field as keyof TestPerson] as string

      const { groupedItems } = useGrouping(items, groupBy, getGroupKey)

      await nextTick()

      expect(Object.keys(groupedItems.value)).toEqual([''])
      expect(groupedItems.value['']).toHaveLength(5)
      expect(groupedItems.value['']).toEqual(testPeople)
    })

    it('should handle empty items array', async () => {
      const items = ref<TestPerson[]>([])
      const groupBy = ref('department')
      const getGroupKey = (item: TestPerson, field: string) => item[field as keyof TestPerson] as string

      const { groupedItems, groupNames } = useGrouping(items, groupBy, getGroupKey)

      await nextTick()

      expect(Object.keys(groupedItems.value)).toEqual([''])
      expect(groupedItems.value['']).toHaveLength(0)
      expect(groupNames.value).toHaveLength(0)
    })

    it('should sort groups alphabetically', async () => {
      const items = ref(testPeople)
      const groupBy = ref('department')
      const getGroupKey = (item: TestPerson, field: string) => item[field as keyof TestPerson] as string

      const { groupedItems } = useGrouping(items, groupBy, getGroupKey)

      await nextTick()

      const groupKeys = Object.keys(groupedItems.value)
      expect(groupKeys).toEqual(['Engineering', 'Marketing', 'Sales'])
    })

    it('should handle different data types', async () => {
      const items = ref(testProducts)
      const groupBy = ref('category')
      const getGroupKey = (item: TestProduct, field: string) => item[field as keyof TestProduct] as string

      const { groupedItems } = useGrouping(items, groupBy, getGroupKey)

      await nextTick()

      expect(Object.keys(groupedItems.value)).toHaveLength(2)
      expect(groupedItems.value.Electronics).toHaveLength(3)
      expect(groupedItems.value.Furniture).toHaveLength(2)
    })

    it('should handle single item', async () => {
      const singlePerson = testPeople[0]
      if (!singlePerson) throw new Error('Test data is missing')
      const items = ref<TestPerson[]>([singlePerson])
      const groupBy = ref('department')
      const getGroupKey = (item: TestPerson, field: string) => item[field as keyof TestPerson] as string

      const { groupedItems } = useGrouping(items, groupBy, getGroupKey)

      await nextTick()

      expect(Object.keys(groupedItems.value)).toEqual(['Engineering'])
      expect(groupedItems.value.Engineering).toBeDefined()
      expect(groupedItems.value.Engineering!).toHaveLength(1)
    })
  })

  describe('Group Names', () => {
    it('should return correct group names', async () => {
      const items = ref(testPeople)
      const groupBy = ref('department')
      const getGroupKey = (item: TestPerson, field: string) => item[field as keyof TestPerson] as string

      const { groupNames } = useGrouping(items, groupBy, getGroupKey)

      await nextTick()

      expect(groupNames.value).toEqual(['Engineering', 'Marketing', 'Sales'])
    })

    it('should exclude empty string from group names', async () => {
      const items = ref(testPeople)
      const groupBy = ref('')
      const getGroupKey = (item: TestPerson, field: string) => item[field as keyof TestPerson] as string

      const { groupNames } = useGrouping(items, groupBy, getGroupKey)

      await nextTick()

      expect(groupNames.value).toHaveLength(0)
    })

    it('should update group names when groupBy changes', async () => {
      const items = ref(testPeople)
      const groupBy = ref('department')
      const getGroupKey = (item: TestPerson, field: string) => item[field as keyof TestPerson] as string

      const { groupNames } = useGrouping(items, groupBy, getGroupKey)

      await nextTick()
      expect(groupNames.value).toEqual(['Engineering', 'Marketing', 'Sales'])

      groupBy.value = 'role'
      await nextTick()

      expect(groupNames.value).toEqual(['Designer', 'Developer', 'Manager', 'Sales Rep'])
    })
  })

  describe('Expand/Collapse Functionality', () => {
    it('should toggle group expansion', async () => {
      const items = ref(testPeople)
      const groupBy = ref('department')
      const getGroupKey = (item: TestPerson, field: string) => item[field as keyof TestPerson] as string

      const { expandedGroups, toggleGroup } = useGrouping(items, groupBy, getGroupKey)

      await nextTick()

      expect(expandedGroups.value.has('Engineering')).toBe(false)

      toggleGroup('Engineering')
      await nextTick()

      expect(expandedGroups.value.has('Engineering')).toBe(true)

      toggleGroup('Engineering')
      await nextTick()

      expect(expandedGroups.value.has('Engineering')).toBe(false)
    })

    it('should expand all groups', async () => {
      const items = ref(testPeople)
      const groupBy = ref('department')
      const getGroupKey = (item: TestPerson, field: string) => item[field as keyof TestPerson] as string

      const { expandedGroups, expandAll, groupNames } = useGrouping(items, groupBy, getGroupKey)

      await nextTick()

      expect(expandedGroups.value.size).toBe(0)

      expandAll()
      await nextTick()

      expect(expandedGroups.value.size).toBe(groupNames.value.length)
      expect(expandedGroups.value.has('Engineering')).toBe(true)
      expect(expandedGroups.value.has('Sales')).toBe(true)
      expect(expandedGroups.value.has('Marketing')).toBe(true)
    })

    it('should collapse all groups', async () => {
      const items = ref(testPeople)
      const groupBy = ref('department')
      const getGroupKey = (item: TestPerson, field: string) => item[field as keyof TestPerson] as string

      const { expandedGroups, expandAll, collapseAll } = useGrouping(items, groupBy, getGroupKey)

      await nextTick()

      expandAll()
      await nextTick()
      expect(expandedGroups.value.size).toBeGreaterThan(0)

      collapseAll()
      await nextTick()

      expect(expandedGroups.value.size).toBe(0)
    })

    it('should handle multiple group toggles', async () => {
      const items = ref(testPeople)
      const groupBy = ref('department')
      const getGroupKey = (item: TestPerson, field: string) => item[field as keyof TestPerson] as string

      const { expandedGroups, toggleGroup } = useGrouping(items, groupBy, getGroupKey)

      await nextTick()

      toggleGroup('Engineering')
      toggleGroup('Sales')
      await nextTick()

      expect(expandedGroups.value.has('Engineering')).toBe(true)
      expect(expandedGroups.value.has('Sales')).toBe(true)
      expect(expandedGroups.value.has('Marketing')).toBe(false)
    })
  })

  describe('Computed Properties', () => {
    it('should calculate hasExpandedGroups correctly', async () => {
      const items = ref(testPeople)
      const groupBy = ref('department')
      const getGroupKey = (item: TestPerson, field: string) => item[field as keyof TestPerson] as string

      const { hasExpandedGroups, toggleGroup } = useGrouping(items, groupBy, getGroupKey)

      await nextTick()

      expect(hasExpandedGroups.value).toBe(false)

      toggleGroup('Engineering')
      await nextTick()

      expect(hasExpandedGroups.value).toBe(true)
    })

    it('should calculate allGroupsExpanded correctly', async () => {
      const items = ref(testPeople)
      const groupBy = ref('department')
      const getGroupKey = (item: TestPerson, field: string) => item[field as keyof TestPerson] as string

      const { allGroupsExpanded, expandAll, toggleGroup } = useGrouping(items, groupBy, getGroupKey)

      await nextTick()

      expect(allGroupsExpanded.value).toBe(false)

      expandAll()
      await nextTick()

      expect(allGroupsExpanded.value).toBe(true)

      toggleGroup('Engineering')
      await nextTick()

      expect(allGroupsExpanded.value).toBe(false)
    })

    it('should return false for allGroupsExpanded when no groups exist', async () => {
      const items = ref<TestPerson[]>([])
      const groupBy = ref('department')
      const getGroupKey = (item: TestPerson, field: string) => item[field as keyof TestPerson] as string

      const { allGroupsExpanded } = useGrouping(items, groupBy, getGroupKey)

      await nextTick()

      expect(allGroupsExpanded.value).toBe(false)
    })
  })

  describe('LocalStorage Persistence', () => {
    it('should save expanded groups to localStorage', async () => {
      const items = ref(testPeople)
      const groupBy = ref('department')
      const getGroupKey = (item: TestPerson, field: string) => item[field as keyof TestPerson] as string
      const storageKey = 'test-expanded-groups'

      const TestComponent = {
        setup() {
          return useGrouping(items, groupBy, getGroupKey, { storageKey })
        },
        template: '<div></div>',
      }

      const wrapper = mount(TestComponent)
      await nextTick()

      const { toggleGroup } = wrapper.vm as any
      toggleGroup('Engineering')
      await nextTick()

      const saved = localStorage.getItem(storageKey)
      expect(saved).toBeTruthy()
      expect(JSON.parse(saved!)).toContain('Engineering')

      wrapper.unmount()
    })

    it('should load expanded groups from localStorage on mount', async () => {
      const storageKey = 'test-expanded-groups-load'
      localStorage.setItem(storageKey, JSON.stringify(['Engineering', 'Sales']))

      const items = ref(testPeople)
      const groupBy = ref('department')
      const getGroupKey = (item: TestPerson, field: string) => item[field as keyof TestPerson] as string

      let composableResult: any

      const TestComponent = {
        setup() {
          composableResult = useGrouping(items, groupBy, getGroupKey, { storageKey })
          return composableResult
        },
        template: '<div></div>',
      }

      const wrapper = mount(TestComponent)
      await nextTick()

      expect(composableResult.expandedGroups.value.has('Engineering')).toBe(true)
      expect(composableResult.expandedGroups.value.has('Sales')).toBe(true)
      expect(composableResult.expandedGroups.value.has('Marketing')).toBe(false)

      wrapper.unmount()
    })

    it('should not persist when no storage key provided', async () => {
      const items = ref(testPeople)
      const groupBy = ref('department')
      const getGroupKey = (item: TestPerson, field: string) => item[field as keyof TestPerson] as string

      const TestComponent = {
        setup() {
          return useGrouping(items, groupBy, getGroupKey)
        },
        template: '<div></div>',
      }

      const wrapper = mount(TestComponent)
      await nextTick()

      const { toggleGroup } = wrapper.vm as any
      toggleGroup('Engineering')
      await nextTick()

      expect(localStorage.length).toBe(0)

      wrapper.unmount()
    })

    it('should handle corrupted localStorage data gracefully', async () => {
      const storageKey = 'test-corrupted'
      localStorage.setItem(storageKey, 'invalid-json{')

      const items = ref(testPeople)
      const groupBy = ref('department')
      const getGroupKey = (item: TestPerson, field: string) => item[field as keyof TestPerson] as string

      let composableResult: any

      const TestComponent = {
        setup() {
          composableResult = useGrouping(items, groupBy, getGroupKey, { storageKey })
          return composableResult
        },
        template: '<div></div>',
      }

      const wrapper = mount(TestComponent)
      await nextTick()

      expect(composableResult.expandedGroups.value.size).toBe(0)
      expect(console.warn).toHaveBeenCalled()

      wrapper.unmount()
    })

    it('should handle localStorage errors when saving', async () => {
      const items = ref(testPeople)
      const groupBy = ref('department')
      const getGroupKey = (item: TestPerson, field: string) => item[field as keyof TestPerson] as string
      const storageKey = 'test-save-error'

      // Mock setItem to throw an error
      const originalSetItem = localStorage.setItem
      localStorage.setItem = vi.fn().mockImplementation(() => {
        throw new Error('Storage quota exceeded')
      })

      const TestComponent = {
        setup() {
          return useGrouping(items, groupBy, getGroupKey, { storageKey })
        },
        template: '<div></div>',
      }

      const wrapper = mount(TestComponent)
      await nextTick()

      const { toggleGroup } = wrapper.vm as any
      toggleGroup('Engineering')
      await nextTick()

      expect(console.warn).toHaveBeenCalledWith(
        'Failed to save expanded groups to localStorage:',
        expect.any(Error)
      )

      // Restore original implementation
      localStorage.setItem = originalSetItem
      wrapper.unmount()
    })
  })

  describe('Default Expanded Option', () => {
    it('should expand all groups on mount when defaultExpanded is true', async () => {
      const items = ref(testPeople)
      const groupBy = ref('department')
      const getGroupKey = (item: TestPerson, field: string) => item[field as keyof TestPerson] as string

      let composableResult: any

      const TestComponent = {
        setup() {
          composableResult = useGrouping(items, groupBy, getGroupKey, { defaultExpanded: true })
          return composableResult
        },
        template: '<div></div>',
      }

      const wrapper = mount(TestComponent)
      await nextTick()

      expect(composableResult.expandedGroups.value.size).toBe(composableResult.groupNames.value.length)
      expect(composableResult.expandedGroups.value.has('Engineering')).toBe(true)
      expect(composableResult.expandedGroups.value.has('Sales')).toBe(true)
      expect(composableResult.expandedGroups.value.has('Marketing')).toBe(true)

      wrapper.unmount()
    })

    it('should not expand groups when defaultExpanded is false', async () => {
      const items = ref(testPeople)
      const groupBy = ref('department')
      const getGroupKey = (item: TestPerson, field: string) => item[field as keyof TestPerson] as string

      let composableResult: any

      const TestComponent = {
        setup() {
          composableResult = useGrouping(items, groupBy, getGroupKey, { defaultExpanded: false })
          return composableResult
        },
        template: '<div></div>',
      }

      const wrapper = mount(TestComponent)
      await nextTick()

      expect(composableResult.expandedGroups.value.size).toBe(0)

      wrapper.unmount()
    })

    it('should not override localStorage state with defaultExpanded', async () => {
      const storageKey = 'test-default-override'
      localStorage.setItem(storageKey, JSON.stringify(['Engineering']))

      const items = ref(testPeople)
      const groupBy = ref('department')
      const getGroupKey = (item: TestPerson, field: string) => item[field as keyof TestPerson] as string

      let composableResult: any

      const TestComponent = {
        setup() {
          composableResult = useGrouping(items, groupBy, getGroupKey, {
            storageKey,
            defaultExpanded: true,
          })
          return composableResult
        },
        template: '<div></div>',
      }

      const wrapper = mount(TestComponent)
      await nextTick()

      expect(composableResult.expandedGroups.value.size).toBe(1)
      expect(composableResult.expandedGroups.value.has('Engineering')).toBe(true)
      expect(composableResult.expandedGroups.value.has('Sales')).toBe(false)

      wrapper.unmount()
    })

    it('should not expand when no groups exist', async () => {
      const items = ref<TestPerson[]>([])
      const groupBy = ref('department')
      const getGroupKey = (item: TestPerson, field: string) => item[field as keyof TestPerson] as string

      let composableResult: any

      const TestComponent = {
        setup() {
          composableResult = useGrouping(items, groupBy, getGroupKey, { defaultExpanded: true })
          return composableResult
        },
        template: '<div></div>',
      }

      const wrapper = mount(TestComponent)
      await nextTick()

      expect(composableResult.expandedGroups.value.size).toBe(0)

      wrapper.unmount()
    })
  })

  describe('Reactivity', () => {
    it('should react to items changes', async () => {
      const items = ref(testPeople.slice(0, 2))
      const groupBy = ref('department')
      const getGroupKey = (item: TestPerson, field: string) => item[field as keyof TestPerson] as string

      const { groupedItems } = useGrouping(items, groupBy, getGroupKey)

      await nextTick()
      expect(Object.keys(groupedItems.value)).toHaveLength(1)

      items.value = testPeople
      await nextTick()

      expect(Object.keys(groupedItems.value)).toHaveLength(3)
    })

    it('should react to groupBy changes', async () => {
      const items = ref(testPeople)
      const groupBy = ref('department')
      const getGroupKey = (item: TestPerson, field: string) => item[field as keyof TestPerson] as string

      const { groupedItems } = useGrouping(items, groupBy, getGroupKey)

      await nextTick()
      expect(Object.keys(groupedItems.value)).toEqual(['Engineering', 'Marketing', 'Sales'])

      groupBy.value = 'role'
      await nextTick()

      expect(Object.keys(groupedItems.value)).toEqual(['Designer', 'Developer', 'Manager', 'Sales Rep'])
    })

    it('should maintain expanded state when groupBy changes', async () => {
      const items = ref(testPeople)
      const groupBy = ref('department')
      const getGroupKey = (item: TestPerson, field: string) => item[field as keyof TestPerson] as string

      const { expandedGroups, toggleGroup } = useGrouping(items, groupBy, getGroupKey)

      await nextTick()

      toggleGroup('Engineering')
      await nextTick()

      expect(expandedGroups.value.has('Engineering')).toBe(true)

      groupBy.value = 'role'
      await nextTick()

      // Should keep the same expanded groups set (even if not applicable)
      expect(expandedGroups.value.has('Engineering')).toBe(true)
    })

    it('should handle adding items to groups', async () => {
      const items = ref([...testPeople])
      const groupBy = ref('department')
      const getGroupKey = (item: TestPerson, field: string) => item[field as keyof TestPerson] as string

      const { groupedItems } = useGrouping(items, groupBy, getGroupKey)

      await nextTick()
      expect(groupedItems.value.Engineering).toBeDefined()
      const initialEngineeringCount = groupedItems.value.Engineering!.length

      items.value.push({
        id: 6,
        name: 'Frank',
        department: 'Engineering',
        role: 'Developer',
        age: 29,
      })
      await nextTick()

      expect(groupedItems.value.Engineering).toBeDefined()
      expect(groupedItems.value.Engineering!.length).toBe(initialEngineeringCount + 1)
    })

    it('should handle removing items', async () => {
      const items = ref([...testPeople])
      const groupBy = ref('department')
      const getGroupKey = (item: TestPerson, field: string) => item[field as keyof TestPerson] as string

      const { groupedItems } = useGrouping(items, groupBy, getGroupKey)

      await nextTick()
      const initialGroupCount = Object.keys(groupedItems.value).length

      items.value = items.value.filter(p => p.department !== 'Marketing')
      await nextTick()

      expect(Object.keys(groupedItems.value).length).toBe(initialGroupCount - 1)
      expect(groupedItems.value.Marketing).toBeUndefined()
    })
  })

  describe('Large Dataset Performance', () => {
    it('should handle large datasets efficiently', async () => {
      const largeDataset = Array.from({ length: 1000 }, (_, i) => ({
        id: i,
        name: `Person ${i}`,
        department: `Dept ${i % 20}`,
        role: `Role ${i % 10}`,
        age: 20 + (i % 40),
      }))

      const items = ref(largeDataset)
      const groupBy = ref('department')
      const getGroupKey = (item: TestPerson, field: string) => item[field as keyof TestPerson] as string

      const startTime = performance.now()
      const { groupedItems } = useGrouping(items, groupBy, getGroupKey)
      await nextTick()
      const endTime = performance.now()

      expect(Object.keys(groupedItems.value)).toHaveLength(20)
      expect(endTime - startTime).toBeLessThan(100) // Should be fast
    })

    it('should handle many group operations', async () => {
      const items = ref(testPeople)
      const groupBy = ref('department')
      const getGroupKey = (item: TestPerson, field: string) => item[field as keyof TestPerson] as string

      const { toggleGroup, expandAll, collapseAll } = useGrouping(items, groupBy, getGroupKey)

      await nextTick()

      // Perform many operations
      for (let i = 0; i < 100; i++) {
        expandAll()
        await nextTick()
        collapseAll()
        await nextTick()
        toggleGroup('Engineering')
        await nextTick()
      }

      // Should complete without errors
      expect(true).toBe(true)
    })
  })

  describe('Edge Cases', () => {
    it('should handle undefined group keys', async () => {
      const itemsWithUndefined = [
        { id: 1, name: 'Test', department: undefined as any },
        { id: 2, name: 'Test2', department: 'Engineering' },
      ]
      const items = ref(itemsWithUndefined)
      const groupBy = ref('department')
      const getGroupKey = (item: any, field: string) => {
        const value = item[field]
        return value === undefined ? 'Unknown' : value
      }

      const { groupedItems } = useGrouping(items, groupBy, getGroupKey)

      await nextTick()

      expect(groupedItems.value.Unknown).toBeDefined()
      expect(groupedItems.value.Engineering).toBeDefined()
    })

    it('should handle null group keys', async () => {
      const itemsWithNull = [
        { id: 1, name: 'Test', department: null as any },
        { id: 2, name: 'Test2', department: 'Engineering' },
      ]
      const items = ref(itemsWithNull)
      const groupBy = ref('department')
      const getGroupKey = (item: any, field: string) => {
        const value = item[field]
        return value === null ? 'Unassigned' : value
      }

      const { groupedItems } = useGrouping(items, groupBy, getGroupKey)

      await nextTick()

      expect(groupedItems.value.Unassigned).toBeDefined()
      expect(groupedItems.value.Engineering).toBeDefined()
    })

    it('should handle special characters in group names', async () => {
      const specialItems = [
        { id: 1, name: 'Test', category: 'A&B' },
        { id: 2, name: 'Test2', category: 'C/D' },
        { id: 3, name: 'Test3', category: 'E-F' },
      ]
      const items = ref(specialItems)
      const groupBy = ref('category')
      const getGroupKey = (item: any, field: string) => item[field]

      const { groupedItems, toggleGroup, expandedGroups } = useGrouping(items, groupBy, getGroupKey)

      await nextTick()

      expect(groupedItems.value['A&B']).toBeDefined()
      expect(groupedItems.value['C/D']).toBeDefined()

      toggleGroup('A&B')
      await nextTick()

      expect(expandedGroups.value.has('A&B')).toBe(true)
    })

    it('should handle very long group names', async () => {
      const longName = 'A'.repeat(1000)
      const longNameItems = [
        { id: 1, name: 'Test', category: longName },
      ]
      const items = ref(longNameItems)
      const groupBy = ref('category')
      const getGroupKey = (item: any, field: string) => item[field]

      const { groupedItems } = useGrouping(items, groupBy, getGroupKey)

      await nextTick()

      expect(groupedItems.value[longName]).toBeDefined()
      expect(groupedItems.value[longName]!).toHaveLength(1)
    })
  })

  describe('TypeScript Types', () => {
    it('should maintain correct types for expanded groups', async () => {
      const items = ref(testPeople)
      const groupBy = ref('department')
      const getGroupKey = (item: TestPerson, field: string) => item[field as keyof TestPerson] as string

      const { expandedGroups } = useGrouping(items, groupBy, getGroupKey)

      await nextTick()

      // TypeScript should recognize this as readonly
      expect(expandedGroups.value).toBeInstanceOf(Set)
    })

    it('should work with complex generic types', async () => {
      interface ComplexItem {
        id: string
        metadata: {
          category: string
          tags: string[]
        }
      }

      const complexItems: ComplexItem[] = [
        { id: '1', metadata: { category: 'A', tags: ['tag1'] } },
        { id: '2', metadata: { category: 'B', tags: ['tag2'] } },
      ]

      const items = ref(complexItems)
      const groupBy = ref('category')
      const getGroupKey = (item: ComplexItem) => item.metadata.category

      const { groupedItems } = useGrouping(items, groupBy, getGroupKey)

      await nextTick()

      expect(groupedItems.value.A).toBeDefined()
      expect(groupedItems.value.B).toBeDefined()
    })
  })

  describe('Integration with Component', () => {
    it('should work correctly in a component context', async () => {
      const TestComponent = {
        setup() {
          const items = ref(testPeople)
          const groupBy = ref('department')
          const getGroupKey = (item: TestPerson, field: string) =>
            item[field as keyof TestPerson] as string

          const grouping = useGrouping(items, groupBy, getGroupKey)

          return {
            items,
            groupBy,
            ...grouping,
          }
        },
        template: `
          <div>
            <div v-for="(group, name) in groupedItems" :key="name">
              <h3>{{ name }}</h3>
              <div v-for="item in group" :key="item.id">
                {{ item.name }}
              </div>
            </div>
          </div>
        `,
      }

      const wrapper = mount(TestComponent)
      await nextTick()

      expect(wrapper.html()).toContain('Engineering')
      expect(wrapper.html()).toContain('Alice')

      wrapper.unmount()
    })

    it('should handle component unmount cleanly', async () => {
      const TestComponent = {
        setup() {
          const items = ref(testPeople)
          const groupBy = ref('department')
          const getGroupKey = (item: TestPerson, field: string) =>
            item[field as keyof TestPerson] as string

          return useGrouping(items, groupBy, getGroupKey, {
            storageKey: 'test-unmount',
          })
        },
        template: '<div></div>',
      }

      const wrapper = mount(TestComponent)
      await nextTick()

      const { toggleGroup } = wrapper.vm as any
      toggleGroup('Engineering')
      await nextTick()

      wrapper.unmount()

      // Should have saved to localStorage before unmounting
      const saved = localStorage.getItem('test-unmount')
      expect(saved).toBeTruthy()
    })
  })
})
