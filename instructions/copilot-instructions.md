# Spinwheel Copilot Instructions

## Project Overview
Spinwheel is a gamified wheel-of-fortune experience built as a multi-part stack. Currently, the **frontend is active and functional** (`/frontend`), while backend (`/backend`) and contract (`/contract`) directories are scaffolded but empty—focus development on the React frontend unless specifically asked about backend integration.

## Architecture & Key Components

### Frontend Stack (React + TypeScript + Vite)
- **Framework**: React 19.1, TypeScript 5.8, Vite 6.3
- **Styling**: Tailwind CSS 3.4 with PostCSS for dark mode support
- **Roulette Library**: `react-custom-roulette` (v1.4.1) for the spinning wheel UI
- **Build**: TypeScript compilation → Vite bundling (`npm run build`)
- **Dev Server**: `npm run dev` (Vite with HMR)

### Component Structure (`/frontend/src/components`)
1. **Wheel.tsx** - Core spinning wheel component using `react-custom-roulette`
   - Hardcoded data: 4 wheel segments with distinct colors and Apple-inspired styling
   - Configuration: `spinDuration={0.5}`, no initial animation, glass-morphism design
   - Uses frosted glass UI pattern with Tailwind arbitrary values (e.g., `bg-white/30`)

2. **ItemList.tsx** - Sidebar component for managing wheel items (stub implementation)
   
3. **Item.tsx** / **ItemRow.tsx** - Individual item/input row components
   - Pattern: Frosted glass inputs with dark mode support
   - Focused state: blue ring (`focus:ring-blue-500`)

4. **Playground.tsx** - Layout container orchestrating Wheel + ItemList

5. **Navbar.tsx** - Top navigation component

### Data & State Management
- **Current**: Local component state (no Redux/Context yet)
- **Wheel Data**: Static array in Wheel.tsx with `{ option, backgroundColor, textColor }`
- **No API integration** currently—backend connection is a TODO

## Developer Workflows

### Run Development Server
```bash
cd /workspaces/spinwheel/frontend
npm install  # One-time setup
npm run dev  # Start Vite dev server (HMR enabled)
```

### Build for Production
```bash
npm run build  # Runs: tsc -b && vite build
# Output: /frontend/dist
```

### Linting
```bash
npm run lint  # ESLint with React Hooks & Refresh rules
```

### Preview Built Output
```bash
npm run preview  # Preview dist/ locally
```

## Coding Patterns & Conventions

### Component Structure
- **Functional components** with React.FC typing: `const MyComponent: React.FC = () => { ... }`
- **Exports**: Use named exports in `index.ts` for barrel imports (see `/components/index.ts`)
- **Naming**: PascalCase for component files

### Styling Approach
- **Tailwind-first**: Inline Tailwind classes, no separate CSS modules
- **Glass morphism**: Apply `backdrop-blur-md`, `bg-white/20`, border patterns for frosted glass effect
- **Dark mode**: Use `dark:` prefix classes consistently (e.g., `dark:bg-white/10 dark:text-white`)
- **Color palette**: Mix Apple-inspired neutrals (#f5f5f7, #1a1a1a) with accent blue (#0071e3)

### TypeScript Configuration
- **Base config**: `tsconfig.app.json` for application code
- **Node config**: `tsconfig.node.json` for build tools
- No path aliases configured—use relative imports

### ESLint Rules
- **react-refresh/only-export-components**: Warn if non-component exports in .tsx files with `allowConstantExport: true`
- **react-hooks**: Enforces correct hooks usage
- React 19 support via @vitejs/plugin-react (Babel-based Fast Refresh)

## Integration Points & Next Steps

### Backend Connection (TODO)
- Backend endpoints not yet defined
- Likely needed for: wheel item persistence, spin results, user progress
- **Integration plan**: Add API client layer (e.g., fetch or axios) between frontend and backend

### State Management (Future)
- Current local state may need centralization when backend connects
- Consider Context API or lightweight state manager if complexity grows

### Deployment
- Frontend deploys from `/frontend/dist` after `npm run build`
- Backend deployment strategy TBD

## Critical Files to Reference
- `/frontend/src/components/Wheel.tsx` - Main business logic for roulette UI
- `/frontend/tailwind.config.js` - Styling theme configuration
- `/frontend/vite.config.ts` - Build configuration
- `/frontend/package.json` - Dependencies and scripts
- `/frontend/eslint.config.js` - Linting rules
