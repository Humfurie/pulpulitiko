/// <reference types="vite/client" />
/// <reference types="@nuxt/types" />

// Environment variables type declarations
interface ImportMetaEnv {
  readonly NUXT_PUBLIC_API_URL: string
  readonly NUXT_PUBLIC_SITE_URL: string
  readonly NUXT_API_INTERNAL_URL: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}

// DOMPurify types
declare module 'isomorphic-dompurify' {
  import { DOMPurifyI } from 'dompurify'
  const DOMPurify: DOMPurifyI
  export default DOMPurify
}
