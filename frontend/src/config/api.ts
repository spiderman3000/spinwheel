// frontend/src/config/api.ts
export const API_CONFIG = {
    baseURL: import.meta.env.VITE_API_URL || 'http://localhost:8080/v1',
    timeout: 10000,
};

// For development
export const isDevelopment = import.meta.env.DEV;
export const isProduction = import.meta.env.PROD;