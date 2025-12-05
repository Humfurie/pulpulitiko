<script setup lang="ts">
import type {
  RegionListItem,
  ProvinceListItem,
  CityMunicipalityListItem,
  BarangayListItem,
  LocationHierarchy
} from '~/types'

const props = defineProps<{
  modelValue?: string // Barangay ID (most granular selection)
  showBarangay?: boolean // Whether to include barangay level (default: true)
  showDistrict?: boolean // Whether to show congressional district info
  required?: boolean
  disabled?: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string | undefined]
  'change': [hierarchy: LocationHierarchy]
}>()

const api = useApi()

// Selected IDs
const selectedRegionId = ref<string | undefined>()
const selectedProvinceId = ref<string | undefined>()
const selectedCityId = ref<string | undefined>()
const selectedBarangayId = ref<string | undefined>()

// Loading states
const loadingRegions = ref(false)
const loadingProvinces = ref(false)
const loadingCities = ref(false)
const loadingBarangays = ref(false)

// Data
const regions = ref<RegionListItem[]>([])
const provinces = ref<ProvinceListItem[]>([])
const cities = ref<CityMunicipalityListItem[]>([])
const barangays = ref<BarangayListItem[]>([])

// Computed options for selects
const regionOptions = computed(() =>
  regions.value.map(r => ({ label: r.name, value: r.id }))
)

const provinceOptions = computed(() =>
  provinces.value.map(p => ({ label: p.name, value: p.id }))
)

const cityOptions = computed(() =>
  cities.value.map(c => ({
    label: c.is_city ? `${c.name} (City)` : c.name,
    value: c.id
  }))
)

const barangayOptions = computed(() =>
  barangays.value.map(b => ({ label: b.name, value: b.id }))
)

// Load regions on mount
onMounted(async () => {
  loadingRegions.value = true
  try {
    regions.value = await api.getRegions()
  } catch (error) {
    console.error('Failed to load regions:', error)
  } finally {
    loadingRegions.value = false
  }
})

// Watch region selection
watch(selectedRegionId, async (newRegionId) => {
  // Reset downstream selections
  selectedProvinceId.value = undefined
  selectedCityId.value = undefined
  selectedBarangayId.value = undefined
  provinces.value = []
  cities.value = []
  barangays.value = []

  if (!newRegionId) return

  loadingProvinces.value = true
  try {
    provinces.value = await api.getProvincesByRegion(newRegionId)
  } catch (error) {
    console.error('Failed to load provinces:', error)
  } finally {
    loadingProvinces.value = false
  }
})

// Watch province selection
watch(selectedProvinceId, async (newProvinceId) => {
  // Reset downstream selections
  selectedCityId.value = undefined
  selectedBarangayId.value = undefined
  cities.value = []
  barangays.value = []

  if (!newProvinceId) return

  loadingCities.value = true
  try {
    cities.value = await api.getCitiesByProvince(newProvinceId)
  } catch (error) {
    console.error('Failed to load cities:', error)
  } finally {
    loadingCities.value = false
  }
})

// Watch city selection
watch(selectedCityId, async (newCityId) => {
  // Reset barangay selection
  selectedBarangayId.value = undefined
  barangays.value = []

  if (!newCityId || props.showBarangay === false) return

  loadingBarangays.value = true
  try {
    const result = await api.getBarangaysByCity(newCityId, 1, 1000) // Load all barangays
    barangays.value = result.barangays
  } catch (error) {
    console.error('Failed to load barangays:', error)
  } finally {
    loadingBarangays.value = false
  }
})

// Watch barangay selection and emit
watch(selectedBarangayId, (newBarangayId) => {
  emit('update:modelValue', newBarangayId)

  // Build and emit hierarchy
  const hierarchy: LocationHierarchy = {}

  if (selectedRegionId.value) {
    const region = regions.value.find(r => r.id === selectedRegionId.value)
    if (region) hierarchy.region = region
  }

  if (selectedProvinceId.value) {
    const province = provinces.value.find(p => p.id === selectedProvinceId.value)
    if (province) hierarchy.province = province
  }

  if (selectedCityId.value) {
    const city = cities.value.find(c => c.id === selectedCityId.value)
    if (city) hierarchy.city_municipality = city
  }

  if (newBarangayId) {
    const barangay = barangays.value.find(b => b.id === newBarangayId)
    if (barangay) hierarchy.barangay = barangay
  }

  emit('change', hierarchy)
})

// If barangay not shown, emit on city selection
watch(selectedCityId, (newCityId) => {
  if (props.showBarangay !== false) return

  emit('update:modelValue', newCityId)

  const hierarchy: LocationHierarchy = {}

  if (selectedRegionId.value) {
    const region = regions.value.find(r => r.id === selectedRegionId.value)
    if (region) hierarchy.region = region
  }

  if (selectedProvinceId.value) {
    const province = provinces.value.find(p => p.id === selectedProvinceId.value)
    if (province) hierarchy.province = province
  }

  if (newCityId) {
    const city = cities.value.find(c => c.id === newCityId)
    if (city) hierarchy.city_municipality = city
  }

  emit('change', hierarchy)
})

// Initialize from modelValue if provided
watch(() => props.modelValue, async (barangayId) => {
  if (!barangayId || barangayId === selectedBarangayId.value) return

  try {
    const hierarchy = await api.getLocationHierarchy(barangayId)

    if (hierarchy.region) {
      selectedRegionId.value = hierarchy.region.id
      // Wait for provinces to load
      await nextTick()
      await new Promise(resolve => setTimeout(resolve, 100))
    }

    if (hierarchy.province) {
      selectedProvinceId.value = hierarchy.province.id
      await nextTick()
      await new Promise(resolve => setTimeout(resolve, 100))
    }

    if (hierarchy.city_municipality) {
      selectedCityId.value = hierarchy.city_municipality.id
      await nextTick()
      await new Promise(resolve => setTimeout(resolve, 100))
    }

    if (hierarchy.barangay) {
      selectedBarangayId.value = hierarchy.barangay.id
    }
  } catch (error) {
    console.error('Failed to initialize location picker:', error)
  }
}, { immediate: true })

// Clear selection
function clearSelection() {
  selectedRegionId.value = undefined
  selectedProvinceId.value = undefined
  selectedCityId.value = undefined
  selectedBarangayId.value = undefined
}

defineExpose({
  clearSelection
})
</script>

<template>
  <div class="space-y-3">
    <!-- Region -->
    <UFormGroup label="Region" :required="required">
      <USelectMenu
        v-model="selectedRegionId"
        :options="regionOptions"
        placeholder="Select Region"
        :loading="loadingRegions"
        :disabled="disabled"
        searchable
        searchable-placeholder="Search regions..."
        clear-search-on-close
        value-attribute="value"
        option-attribute="label"
      />
    </UFormGroup>

    <!-- Province -->
    <UFormGroup label="Province" :required="required && !!selectedRegionId">
      <USelectMenu
        v-model="selectedProvinceId"
        :options="provinceOptions"
        placeholder="Select Province"
        :loading="loadingProvinces"
        :disabled="disabled || !selectedRegionId"
        searchable
        searchable-placeholder="Search provinces..."
        clear-search-on-close
        value-attribute="value"
        option-attribute="label"
      />
    </UFormGroup>

    <!-- City/Municipality -->
    <UFormGroup label="City/Municipality" :required="required && !!selectedProvinceId">
      <USelectMenu
        v-model="selectedCityId"
        :options="cityOptions"
        placeholder="Select City/Municipality"
        :loading="loadingCities"
        :disabled="disabled || !selectedProvinceId"
        searchable
        searchable-placeholder="Search cities..."
        clear-search-on-close
        value-attribute="value"
        option-attribute="label"
      />
    </UFormGroup>

    <!-- Barangay (optional) -->
    <UFormGroup v-if="showBarangay !== false" label="Barangay" :required="required && !!selectedCityId">
      <USelectMenu
        v-model="selectedBarangayId"
        :options="barangayOptions"
        placeholder="Select Barangay"
        :loading="loadingBarangays"
        :disabled="disabled || !selectedCityId"
        searchable
        searchable-placeholder="Search barangays..."
        clear-search-on-close
        value-attribute="value"
        option-attribute="label"
      />
    </UFormGroup>
  </div>
</template>
