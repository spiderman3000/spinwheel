# Spinwheel Frontend

A super lightweight, minimal, and high-performance frontend for the Spinwheel application.

## Tech Stack
- **Framework:** [Preact](https://preactjs.com/) (3kB React alternative)
- **Build Tool:** [Vite](https://vitejs.dev/)
- **Styling:** Vanilla CSS with CSS Variables (No Tailwind dependency)
- **Wheel:** Custom HTML5 Canvas implementation (No heavy UI libraries)
- **Language:** TypeScript

## Key Benefits
- **Minimal Bundle Size:** ~19kB for the entire application.
- **Dependency Free:** No external UI libraries or CSS frameworks.
- **Fast Performance:** Canvas-based animations for smooth spinning.
- **Easy to Understand:** Simple codebase with standard web primitives.

## Environment Setup

1. Copy `.env.example` to `.env.local`:
```bash
   cp .env.example .env.local
```

2. Update `VITE_API_URL` with your backend URL.

## Development

Install dependencies:
```bash
npm install
```

Start the development server:
```bash
npm run dev
```

Build for production:
```bash
npm run build
```

Preview production build:
```bash
npm run preview
```
