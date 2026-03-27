import { defineConfig } from 'vite';
import preact from '@preact/preset-vite';

// GitHub Pages project sites live under /<repo>/; CI sets VITE_BASE_PATH (see deploy workflow).
// Local dev and builds without VITE_BASE_PATH keep base "/".
const base =
  process.env.VITE_BASE_PATH && process.env.VITE_BASE_PATH !== '/'
    ? process.env.VITE_BASE_PATH.replace(/\/?$/, '/')
    : '/';

export default defineConfig({
  base,
  plugins: [preact()],
});
