/**
 * API Configuration
 * Reads URLs from environment variables (VITE_*)
 * Fallback to default localhost URLs for development
 */

export const API_CONFIG = {
  gateway: import.meta.env.VITE_API_GATEWAY_URL || 'http://localhost:8080',
  mockApi: import.meta.env.VITE_MOCK_API_URL || 'http://localhost:8100',
  cdn: import.meta.env.VITE_CDN_URL || 'http://localhost:8000',
} as const
