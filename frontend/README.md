# Lector web interface

React 19, TypeScript, Vite, React Router, and TanStack Query. Fonts and assets are served locally.

From the project root, enter `nix develop path:.` and run `npm --prefix frontend ci`.

| Command                | Purpose                                                |
| ---------------------- | ------------------------------------------------------ |
| `npm run dev`          | Development server with `/api` proxy to localhost:3000 |
| `npm run build`        | TypeScript checks and production bundle                |
| `npm test`             | Component and interaction tests                        |
| `npm run test:e2e`     | Chromium tests against an isolated local backend       |
| `npm run lint`         | ESLint, React hooks, and accessibility rules           |
| `npm run format`       | Apply Prettier                                         |
| `npm run format:check` | Check formatting                                       |

Run these commands in `frontend/`. Build the backend first with `make build` before running browser tests. The root Makefile provides combined targets.

The welcome page lives at `/`; the working library lives at `/library`. Existing root URLs with filters redirect to the library and retain their query parameters. Direct links to titles and readers still open their destination.

On a full page load, an opening screen prepares local fonts, route modules, library statistics, the first library page, and current reading items in parallel while the destination renders underneath. Navigation within the app does not repeat the opening screen. Failed preparation or a ten-second timeout offers retry and continue controls. Motion respects the system reduced-motion preference.

Browser tests use an isolated backend on port 4300 and Vite on port 4173, leaving the normal server on port 3000 alone. `LECTOR_API_TARGET` can override the development API proxy.

Reader, detail, activity, preferences, and extension routes use separate bundles warmed during startup. Server state is cached and invalidated after mutations. Filters and pagination live in the URL. Extension results are displayed as text through host-owned components.
