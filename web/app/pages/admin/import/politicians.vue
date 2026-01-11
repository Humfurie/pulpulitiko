<script setup lang="ts">
import type { ApiResponse } from '~/types'

definePageMeta({
  layout: 'admin',
  middleware: 'admin'
})

const auth = useAuth()
const config = useRuntimeConfig()
const baseUrl = import.meta.server ? config.apiInternalUrl : config.public.apiUrl

// Validation result interface
interface ValidationError {
  row: number
  field: string
  error: string
  value?: string
  suggestions?: string[]
}

interface ValidationResult {
  total_rows: number
  valid_rows: number
  invalid_rows: number
  errors?: ValidationError[]
}

// Import log interface
interface ImportLog {
  id: string
  filename: string
  status: string
  total_rows: number
  successful_imports: number
  failed_imports: number
  politicians_created: number
  validation_errors?: ValidationError[]
  started_at: string
}

// File upload state
const selectedFile = ref<File | null>(null)
const fileInput = ref<HTMLInputElement | null>(null)
const dragging = ref(false)

// Validation state
const validating = ref(false)
const validationResult = ref<ValidationResult | null>(null)

// Import state
const importing = ref(false)
const importSuccess = ref(false)
const error = ref('')

// Import logs state
const importLogs = ref<ImportLog[]>([])
const loadingLogs = ref(false)

// Handle file selection
function handleFileSelect(event: Event) {
  const target = event.target as HTMLInputElement
  if (target.files && target.files.length > 0) {
    selectedFile.value = target.files[0] || null
    validationResult.value = null
    importSuccess.value = false
    error.value = ''
  }
}

// Handle drag and drop
function handleDragOver(event: DragEvent) {
  event.preventDefault()
  dragging.value = true
}

function handleDragLeave() {
  dragging.value = false
}

function handleDrop(event: DragEvent) {
  event.preventDefault()
  dragging.value = false

  if (event.dataTransfer?.files && event.dataTransfer.files.length > 0) {
    selectedFile.value = event.dataTransfer.files[0] || null
    validationResult.value = null
    importSuccess.value = false
    error.value = ''
  }
}

// Validate file
async function validateFile() {
  if (!selectedFile.value) {
    error.value = 'Please select a file first'
    return
  }

  validating.value = true
  error.value = ''

  try {
    const formData = new FormData()
    formData.append('file', selectedFile.value)

    const response = await $fetch<ApiResponse<ValidationResult>>(`${baseUrl}/admin/import/politicians/validate`, {
      method: 'POST',
      headers: auth.getAuthHeaders(),
      body: formData
    })

    if (response.success) {
      validationResult.value = response.data
    }
  } catch (e: unknown) {
    const err = e as { data?: { error?: { message?: string } } }
    error.value = err?.data?.error?.message || 'Failed to validate file'
  }

  validating.value = false
}

// Import file
async function importFile() {
  if (!selectedFile.value) {
    error.value = 'Please select a file first'
    return
  }

  importing.value = true
  error.value = ''

  try {
    const formData = new FormData()
    formData.append('file', selectedFile.value)

    const response = await $fetch<ApiResponse<{ message: string }>>(`${baseUrl}/admin/import/politicians`, {
      method: 'POST',
      headers: auth.getAuthHeaders(),
      body: formData
    })

    if (response.success) {
      importSuccess.value = true
      selectedFile.value = null
      validationResult.value = null

      // Refresh import logs
      await fetchImportLogs()
    }
  } catch (e: unknown) {
    const err = e as { data?: { error?: { message?: string } } }
    error.value = err?.data?.error?.message || 'Failed to import file'
  }

  importing.value = false
}

// Fetch import logs
async function fetchImportLogs() {
  loadingLogs.value = true

  try {
    const response = await $fetch<ApiResponse<{ import_logs: ImportLog[] }>>(`${baseUrl}/admin/import/politicians/logs`, {
      headers: auth.getAuthHeaders()
    })

    if (response.success) {
      importLogs.value = response.data.import_logs || []
    }
  } catch (e: unknown) {
    console.error('Failed to load import logs:', e)
  }

  loadingLogs.value = false
}

// Download template
async function downloadTemplate() {
  try {
    const blob = await $fetch(`${baseUrl}/admin/import/politicians/template`, {
      headers: auth.getAuthHeaders(),
      responseType: 'blob'
    })

    // Create download link
    const url = window.URL.createObjectURL(blob as Blob)
    const link = document.createElement('a')
    link.href = url
    link.download = 'politician_import_template.xlsx'
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)
  } catch {
    error.value = 'Failed to download template'
  }
}

// Download error report
async function downloadErrorReport(logId: string) {
  try {
    const blob = await $fetch(`${baseUrl}/admin/import/politicians/logs/${logId}/errors`, {
      headers: auth.getAuthHeaders(),
      responseType: 'blob'
    })

    const url = window.URL.createObjectURL(blob as Blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `import_errors_${logId}.xlsx`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)
  } catch {
    error.value = 'Failed to download error report'
  }
}

// Get status badge color
function getStatusColor(status: string) {
  switch (status) {
    case 'completed': return 'success'
    case 'processing': return 'info'
    case 'failed': return 'error'
    default: return 'neutral'
  }
}

onMounted(() => {
  fetchImportLogs()
})

useSeoMeta({
  title: 'Import Politicians - Pulpulitiko Admin'
})
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Import Politicians</h1>
      <UButton variant="outline" icon="i-heroicons-arrow-down-tray" @click="downloadTemplate">
        Download Template
      </UButton>
    </div>

    <UAlert v-if="error" color="error" :title="error" class="mb-4" />
    <UAlert v-if="importSuccess" color="success" title="Import started successfully!" class="mb-4">
      The import is being processed in the background. Check the import logs below for progress.
    </UAlert>

    <!-- File Upload Section -->
    <UCard class="mb-6">
      <template #header>
        <h2 class="text-lg font-semibold text-gray-900 dark:text-white">Upload Excel File</h2>
      </template>

      <div class="space-y-4">
        <!-- Drag and Drop Area -->
        <div
          :class="[
            'border-2 border-dashed rounded-lg p-8 text-center transition-colors',
            dragging ? 'border-primary-500 bg-primary-50 dark:bg-primary-900/10' : 'border-gray-300 dark:border-gray-700'
          ]"
          @dragover="handleDragOver"
          @dragleave="handleDragLeave"
          @drop="handleDrop"
          @click="fileInput?.click()"
        >
          <input
            ref="fileInput"
            type="file"
            accept=".xlsx,.xls"
            class="hidden"
            @change="handleFileSelect"
          >

          <UIcon
            name="i-heroicons-arrow-up-tray"
            class="mx-auto h-12 w-12 text-gray-400"
          />
          <p class="mt-2 text-sm text-gray-600 dark:text-gray-400">
            <span class="font-semibold text-primary-600 cursor-pointer">Click to upload</span>
            or drag and drop
          </p>
          <p class="text-xs text-gray-500 dark:text-gray-500 mt-1">
            Excel file (.xlsx, .xls) up to 10MB
          </p>
        </div>

        <!-- Selected File Info -->
        <div v-if="selectedFile" class="flex items-center justify-between p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
          <div class="flex items-center gap-3">
            <UIcon name="i-heroicons-document" class="h-8 w-8 text-primary-500" />
            <div>
              <p class="text-sm font-medium text-gray-900 dark:text-white">{{ selectedFile.name }}</p>
              <p class="text-xs text-gray-500">{{ (selectedFile.size / 1024).toFixed(2) }} KB</p>
            </div>
          </div>
          <UButton
            variant="ghost"
            color="error"
            icon="i-heroicons-x-mark"
            size="sm"
            @click="selectedFile = null; validationResult = null"
          />
        </div>

        <!-- Action Buttons -->
        <div v-if="selectedFile" class="flex gap-3">
          <UButton
            :loading="validating"
            :disabled="validating"
            variant="outline"
            block
            @click="validateFile"
          >
            Validate File
          </UButton>
          <UButton
            :loading="importing"
            :disabled="importing || validating"
            color="primary"
            block
            @click="importFile"
          >
            Import
          </UButton>
        </div>
      </div>
    </UCard>

    <!-- Validation Results -->
    <UCard v-if="validationResult" class="mb-6">
      <template #header>
        <h2 class="text-lg font-semibold text-gray-900 dark:text-white">Validation Results</h2>
      </template>

      <div class="space-y-4">
        <!-- Summary Stats -->
        <div class="grid grid-cols-3 gap-4">
          <div class="text-center p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
            <div class="text-2xl font-bold text-gray-900 dark:text-white">{{ validationResult.total_rows }}</div>
            <div class="text-sm text-gray-600 dark:text-gray-400">Total Rows</div>
          </div>
          <div class="text-center p-4 bg-green-50 dark:bg-green-900/20 rounded-lg">
            <div class="text-2xl font-bold text-green-600 dark:text-green-400">{{ validationResult.valid_rows }}</div>
            <div class="text-sm text-gray-600 dark:text-gray-400">Valid</div>
          </div>
          <div class="text-center p-4 bg-red-50 dark:bg-red-900/20 rounded-lg">
            <div class="text-2xl font-bold text-red-600 dark:text-red-400">{{ validationResult.invalid_rows }}</div>
            <div class="text-sm text-gray-600 dark:text-gray-400">Invalid</div>
          </div>
        </div>

        <!-- Errors List -->
        <div v-if="validationResult.errors && validationResult.errors.length > 0" class="space-y-2">
          <h3 class="font-medium text-gray-900 dark:text-white">Validation Errors:</h3>
          <div class="max-h-64 overflow-y-auto space-y-2">
            <UAlert
              v-for="(err, index) in validationResult.errors.slice(0, 10)"
              :key="index"
              color="error"
              :title="`Row ${err.row}: ${err.field}`"
              :description="err.error"
            >
              <template v-if="err.value" #description>
                <p>{{ err.error }}</p>
                <p class="text-xs mt-1">Value: "{{ err.value }}"</p>
                <p v-if="err.suggestions && err.suggestions.length > 0" class="text-xs mt-1">
                  Suggestions: {{ err.suggestions.join(', ') }}
                </p>
              </template>
            </UAlert>
          </div>
          <p v-if="validationResult.errors.length > 10" class="text-sm text-gray-500">
            ... and {{ validationResult.errors.length - 10 }} more errors
          </p>
        </div>
      </div>
    </UCard>

    <!-- Import Logs -->
    <UCard>
      <template #header>
        <div class="flex items-center justify-between">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">Import History</h2>
          <UButton variant="ghost" icon="i-heroicons-arrow-path" size="sm" @click="fetchImportLogs">
            Refresh
          </UButton>
        </div>
      </template>

      <div v-if="loadingLogs" class="py-8 text-center">
        <UIcon name="i-heroicons-arrow-path" class="size-8 animate-spin text-gray-400" />
      </div>

      <div v-else-if="importLogs.length === 0" class="py-8 text-center text-gray-500">
        No import logs yet
      </div>

      <div v-else class="space-y-4">
        <div
          v-for="log in importLogs"
          :key="log.id"
          class="p-4 border border-gray-200 dark:border-gray-800 rounded-lg"
        >
          <div class="flex items-start justify-between">
            <div class="flex-1">
              <div class="flex items-center gap-2 mb-2">
                <span class="font-medium text-gray-900 dark:text-white">{{ log.filename }}</span>
                <UBadge :color="getStatusColor(log.status)" size="sm">{{ log.status }}</UBadge>
              </div>

              <div class="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
                <div>
                  <span class="text-gray-600 dark:text-gray-400">Total:</span>
                  <span class="ml-1 font-medium">{{ log.total_rows }}</span>
                </div>
                <div>
                  <span class="text-gray-600 dark:text-gray-400">Success:</span>
                  <span class="ml-1 font-medium text-green-600">{{ log.successful_imports }}</span>
                </div>
                <div>
                  <span class="text-gray-600 dark:text-gray-400">Failed:</span>
                  <span class="ml-1 font-medium text-red-600">{{ log.failed_imports }}</span>
                </div>
                <div>
                  <span class="text-gray-600 dark:text-gray-400">Created:</span>
                  <span class="ml-1 font-medium">{{ log.politicians_created }}</span>
                </div>
              </div>

              <p class="text-xs text-gray-500 mt-2">
                Started: {{ new Date(log.started_at).toLocaleString() }}
              </p>
            </div>

            <UButton
              v-if="log.status === 'failed' && log.validation_errors && log.validation_errors.length > 0"
              variant="ghost"
              size="sm"
              icon="i-heroicons-arrow-down-tray"
              @click="downloadErrorReport(log.id)"
            >
              Download Errors
            </UButton>
          </div>
        </div>
      </div>
    </UCard>
  </div>
</template>
