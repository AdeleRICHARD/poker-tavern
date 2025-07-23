// API Configuration
export const API_CONFIG = {
  BASE_URL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080',
  WS_BASE_URL: import.meta.env.VITE_WS_BASE_URL || 'ws://localhost:8080',
} as const;

// Helper function to get full API URL
export function getApiUrl(endpoint: string): string {
  return `${API_CONFIG.BASE_URL}${endpoint}`;
}

// Helper function to get WebSocket URL
export function getWsUrl(endpoint: string): string {
  return `${API_CONFIG.WS_BASE_URL}${endpoint}`;
}
