import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  base: process.env.NODE_ENV === 'production'
    ? '/spinwheel/'
    : '/',
  build: {
    sourcemap: false,  // Don't expose source maps in production
    rollupOptions: {
      output: {
        manualChunks: {
          vendor: ['react', 'react-dom'],
          roulette: ['react-custom-roulette']
        }
      }
    }
  }
})