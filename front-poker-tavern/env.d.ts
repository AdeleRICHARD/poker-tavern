/// <reference types="vite/client" />

declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<{}, {}, any>
  export default component
}

declare module 'phaser' {
  export = Phaser
  export as namespace Phaser
}

interface ImportMetaEnv {
  readonly VITE_APP_TITLE: string
  readonly VITE_JIRA_API_URL?: string
  readonly VITE_JIRA_API_TOKEN?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
