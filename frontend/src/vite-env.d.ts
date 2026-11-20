/// <reference types="vite/client" />

interface ImportMetaEnv {
  /** Base URL of the calculator API as seen by the browser (default "/api/v1"). */
  readonly VITE_API_BASE_URL?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
