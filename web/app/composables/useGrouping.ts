import type { Ref } from 'vue'

export interface GroupByOption {
  label: string
  value: string
}

export interface UseGroupingOptions {
  storageKey?: string
  defaultExpanded?: boolean
}

/**
 * Composable for managing grouped data with collapsible sections
 *
 * @param items - Ref containing the array of items to group
 * @param groupBy - Ref containing the current grouping field
 * @param getGroupKey - Function that returns the group key for an item
 * @param options - Optional configuration
 */
export function useGrouping<T>(
  items: Ref<T[]>,
  groupBy: Ref<string>,
  getGroupKey: (item: T, groupByValue: string) => string,
  options: UseGroupingOptions = {}
) {
  const { storageKey, defaultExpanded = false } = options

  // Initialize expanded groups from localStorage if available
  const expandedGroups = ref<Set<string>>(new Set())

  // Load from localStorage on mount
  onMounted(() => {
    if (storageKey) {
      try {
        const saved = localStorage.getItem(storageKey)
        if (saved) {
          expandedGroups.value = new Set(JSON.parse(saved))
        }
      } catch (error) {
        console.warn('Failed to load expanded groups from localStorage:', error)
      }
    }
  })

  // Save to localStorage when changed
  watch(
    expandedGroups,
    (newValue) => {
      if (storageKey) {
        try {
          localStorage.setItem(storageKey, JSON.stringify([...newValue]))
        } catch (error) {
          console.warn('Failed to save expanded groups to localStorage:', error)
        }
      }
    },
    { deep: true }
  )

  // Grouped items computed property
  const groupedItems = computed(() => {
    if (!groupBy.value || items.value.length === 0) {
      return { '': items.value }
    }

    const groups: Record<string, T[]> = {}

    items.value.forEach((item) => {
      const key = getGroupKey(item, groupBy.value)
      if (!groups[key]) {
        groups[key] = []
      }
      groups[key].push(item)
    })

    // Sort groups alphabetically
    const sortedGroups: Record<string, T[]> = {}
    Object.keys(groups)
      .sort()
      .forEach((key) => {
        sortedGroups[key] = groups[key]!
      })

    return sortedGroups
  })

  // Group names computed property
  const groupNames = computed(() => Object.keys(groupedItems.value).filter(key => key !== ''))

  // Toggle a specific group
  function toggleGroup(groupName: string) {
    const newSet = new Set(expandedGroups.value)
    if (newSet.has(groupName)) {
      newSet.delete(groupName)
    } else {
      newSet.add(groupName)
    }
    expandedGroups.value = newSet
  }

  // Expand all groups
  function expandAll() {
    expandedGroups.value = new Set(groupNames.value)
  }

  // Collapse all groups
  function collapseAll() {
    expandedGroups.value = new Set()
  }

  // Check if any groups are expanded
  const hasExpandedGroups = computed(() => expandedGroups.value.size > 0)

  // Check if all groups are expanded
  const allGroupsExpanded = computed(
    () => groupNames.value.length > 0 && expandedGroups.value.size === groupNames.value.length
  )

  // Auto-expand groups on mount if defaultExpanded is true
  onMounted(() => {
    if (defaultExpanded && expandedGroups.value.size === 0 && groupNames.value.length > 0) {
      expandAll()
    }
  })

  return {
    expandedGroups: readonly(expandedGroups),
    groupedItems,
    groupNames,
    toggleGroup,
    expandAll,
    collapseAll,
    hasExpandedGroups,
    allGroupsExpanded,
  }
}
