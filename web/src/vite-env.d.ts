/// <reference types="vite/client" />

declare const __APP_VERSION__: string
declare const __APP_BUILD_TIME__: string

interface ImportMetaEnv {
  readonly VITE_API_BASE: string
  readonly VITE_ENABLE_MOCK?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
