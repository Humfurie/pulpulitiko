<script setup lang="ts">
import { useEditor, EditorContent } from '@tiptap/vue-3'
import StarterKit from '@tiptap/starter-kit'
import Image from '@tiptap/extension-image'
import Link from '@tiptap/extension-link'
import Placeholder from '@tiptap/extension-placeholder'

const props = defineProps<{
  modelValue: string
  placeholder?: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const auth = useAuth()
const api = useApi()
const uploading = ref(false)
const fileInput = ref<HTMLInputElement | null>(null)

const editor = useEditor({
  content: props.modelValue,
  extensions: [
    StarterKit.configure({
      heading: {
        levels: [2, 3, 4]
      }
    }),
    Image.configure({
      inline: false,
      allowBase64: false,
      HTMLAttributes: {
        class: 'rounded-lg max-w-full'
      }
    }),
    Link.configure({
      openOnClick: false,
      HTMLAttributes: {
        class: 'text-primary-500 underline'
      }
    }),
    Placeholder.configure({
      placeholder: props.placeholder || 'Write your content here...'
    })
  ],
  editorProps: {
    attributes: {
      class: 'prose prose-sm sm:prose-base dark:prose-invert max-w-none focus:outline-none min-h-[400px] px-4 py-3'
    },
    handleDrop: (view, event, _slice, moved) => {
      if (!moved && event.dataTransfer?.files?.length) {
        const file = event.dataTransfer.files[0]
        if (file.type.startsWith('image/')) {
          event.preventDefault()
          uploadAndInsertImage(file)
          return true
        }
      }
      return false
    },
    handlePaste: (view, event) => {
      const items = event.clipboardData?.items
      if (items) {
        for (const item of items) {
          if (item.type.startsWith('image/')) {
            event.preventDefault()
            const file = item.getAsFile()
            if (file) uploadAndInsertImage(file)
            return true
          }
        }
      }
      return false
    }
  },
  onUpdate: ({ editor }) => {
    emit('update:modelValue', editor.getHTML())
  }
})

watch(() => props.modelValue, (newValue) => {
  if (editor.value && editor.value.getHTML() !== newValue) {
    editor.value.commands.setContent(newValue, { emitUpdate: false })
  }
})

async function uploadAndInsertImage(file: File) {
  if (!file.type.startsWith('image/')) {
    alert('Please select an image file')
    return
  }

  uploading.value = true
  try {
    const result = await api.uploadFile(file, auth.getAuthHeaders())
    editor.value?.chain().focus().setImage({ src: result.url }).run()
  } catch (e: unknown) {
    const err = e as { message?: string }
    alert(err.message || 'Failed to upload image')
  } finally {
    uploading.value = false
  }
}

function handleFileSelect(event: Event) {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (file) {
    uploadAndInsertImage(file)
    target.value = ''
  }
}

function addLink() {
  const url = prompt('Enter URL:')
  if (url) {
    editor.value?.chain().focus().setLink({ href: url }).run()
  }
}

onBeforeUnmount(() => {
  editor.value?.destroy()
})
</script>

<template>
  <div class="rich-text-editor border border-gray-200 dark:border-gray-700 rounded-lg overflow-hidden bg-white dark:bg-gray-900">
    <!-- Toolbar -->
    <div v-if="editor" class="toolbar flex flex-wrap items-center gap-1 p-2 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800">
      <!-- Text formatting -->
      <UButton
        size="xs"
        :variant="editor.isActive('bold') ? 'solid' : 'ghost'"
        icon="i-heroicons-bold"
        @click="editor.chain().focus().toggleBold().run()"
      />
      <UButton
        size="xs"
        :variant="editor.isActive('italic') ? 'solid' : 'ghost'"
        icon="i-heroicons-italic"
        @click="editor.chain().focus().toggleItalic().run()"
      />
      <UButton
        size="xs"
        :variant="editor.isActive('strike') ? 'solid' : 'ghost'"
        icon="i-heroicons-strikethrough"
        @click="editor.chain().focus().toggleStrike().run()"
      />

      <div class="w-px h-6 bg-gray-300 dark:bg-gray-600 mx-1" />

      <!-- Headings -->
      <UButton
        size="xs"
        :variant="editor.isActive('heading', { level: 2 }) ? 'solid' : 'ghost'"
        @click="editor.chain().focus().toggleHeading({ level: 2 }).run()"
      >
        H2
      </UButton>
      <UButton
        size="xs"
        :variant="editor.isActive('heading', { level: 3 }) ? 'solid' : 'ghost'"
        @click="editor.chain().focus().toggleHeading({ level: 3 }).run()"
      >
        H3
      </UButton>
      <UButton
        size="xs"
        :variant="editor.isActive('heading', { level: 4 }) ? 'solid' : 'ghost'"
        @click="editor.chain().focus().toggleHeading({ level: 4 }).run()"
      >
        H4
      </UButton>

      <div class="w-px h-6 bg-gray-300 dark:bg-gray-600 mx-1" />

      <!-- Lists -->
      <UButton
        size="xs"
        :variant="editor.isActive('bulletList') ? 'solid' : 'ghost'"
        icon="i-heroicons-list-bullet"
        @click="editor.chain().focus().toggleBulletList().run()"
      />
      <UButton
        size="xs"
        :variant="editor.isActive('orderedList') ? 'solid' : 'ghost'"
        icon="i-heroicons-numbered-list"
        @click="editor.chain().focus().toggleOrderedList().run()"
      />

      <div class="w-px h-6 bg-gray-300 dark:bg-gray-600 mx-1" />

      <!-- Block elements -->
      <UButton
        size="xs"
        :variant="editor.isActive('blockquote') ? 'solid' : 'ghost'"
        icon="i-heroicons-chat-bubble-bottom-center-text"
        @click="editor.chain().focus().toggleBlockquote().run()"
      />
      <UButton
        size="xs"
        :variant="editor.isActive('codeBlock') ? 'solid' : 'ghost'"
        icon="i-heroicons-code-bracket"
        @click="editor.chain().focus().toggleCodeBlock().run()"
      />

      <div class="w-px h-6 bg-gray-300 dark:bg-gray-600 mx-1" />

      <!-- Link -->
      <UButton
        size="xs"
        :variant="editor.isActive('link') ? 'solid' : 'ghost'"
        icon="i-heroicons-link"
        @click="addLink"
      />
      <UButton
        v-if="editor.isActive('link')"
        size="xs"
        variant="ghost"
        color="error"
        icon="i-heroicons-link-slash"
        @click="editor.chain().focus().unsetLink().run()"
      />

      <div class="w-px h-6 bg-gray-300 dark:bg-gray-600 mx-1" />

      <!-- Image upload -->
      <UButton
        size="xs"
        variant="ghost"
        icon="i-heroicons-photo"
        :loading="uploading"
        @click="fileInput?.click()"
      />
      <input
        ref="fileInput"
        type="file"
        accept="image/*"
        class="hidden"
        @change="handleFileSelect"
      >

      <div class="w-px h-6 bg-gray-300 dark:bg-gray-600 mx-1" />

      <!-- Undo/Redo -->
      <UButton
        size="xs"
        variant="ghost"
        icon="i-heroicons-arrow-uturn-left"
        :disabled="!editor.can().undo()"
        @click="editor.chain().focus().undo().run()"
      />
      <UButton
        size="xs"
        variant="ghost"
        icon="i-heroicons-arrow-uturn-right"
        :disabled="!editor.can().redo()"
        @click="editor.chain().focus().redo().run()"
      />

      <!-- Upload indicator -->
      <span v-if="uploading" class="ml-2 text-xs text-gray-500">
        Uploading image...
      </span>
    </div>

    <!-- Editor content -->
    <EditorContent :editor="editor" />
  </div>
</template>

<style>
.rich-text-editor .ProseMirror {
  min-height: 400px;
}

.rich-text-editor .ProseMirror p.is-editor-empty:first-child::before {
  content: attr(data-placeholder);
  float: left;
  color: #adb5bd;
  pointer-events: none;
  height: 0;
}

.rich-text-editor .ProseMirror:focus {
  outline: none;
}

.rich-text-editor .ProseMirror img {
  max-width: 100%;
  height: auto;
  margin: 1rem 0;
}

.rich-text-editor .ProseMirror blockquote {
  border-left: 3px solid #e5e7eb;
  padding-left: 1rem;
  margin: 1rem 0;
  color: #6b7280;
}

.dark .rich-text-editor .ProseMirror blockquote {
  border-left-color: #374151;
  color: #9ca3af;
}

.rich-text-editor .ProseMirror pre {
  background: #1f2937;
  color: #e5e7eb;
  padding: 1rem;
  border-radius: 0.5rem;
  overflow-x: auto;
}

.rich-text-editor .ProseMirror code {
  background: #f3f4f6;
  padding: 0.25rem 0.5rem;
  border-radius: 0.25rem;
  font-size: 0.875rem;
}

.dark .rich-text-editor .ProseMirror code {
  background: #374151;
}

.rich-text-editor .ProseMirror pre code {
  background: transparent;
  padding: 0;
}

.rich-text-editor .ProseMirror ul,
.rich-text-editor .ProseMirror ol {
  padding-left: 1.5rem;
  margin: 0.5rem 0;
}

.rich-text-editor .ProseMirror h2 {
  font-size: 1.5rem;
  font-weight: 700;
  margin-top: 1.5rem;
  margin-bottom: 0.75rem;
}

.rich-text-editor .ProseMirror h3 {
  font-size: 1.25rem;
  font-weight: 600;
  margin-top: 1.25rem;
  margin-bottom: 0.5rem;
}

.rich-text-editor .ProseMirror h4 {
  font-size: 1.125rem;
  font-weight: 600;
  margin-top: 1rem;
  margin-bottom: 0.5rem;
}
</style>
