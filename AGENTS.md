# Repository Guidelines

## Project Structure & Module Organization

The Go API starts at `cmd/server/main.go`. Backend code lives under `internal/`: HTTP code in `api/`, business logic in `service/`, persistence in `repository/`, and shared data in `model/`. The Vue 3 frontend is in `web/src/`; place pages in `views/`, reusable UI in `components/`, API clients in `api/`, and shared definitions in `types/`. Runtime configuration belongs in `configs/`, while deployment files live in `deploy/`. Production builds embed `web/dist` through `web/embed.go`.

## Build, Test, and Development Commands

- `make all`: install frontend dependencies, build the Vue app, and compile the `gosee` binary.
- `make run`: run the backend on port 8080 using `configs/config.yaml`.
- `make dev`: start the Vite development server on port 5173; run it alongside `make run`.
- `make check`: run Vue/TypeScript type checking and `go vet ./...`.
- `go test ./...`: run all backend tests.
- `cd web && pnpm build`: type-check and build the frontend only.
- `make build-linux`: create a static Linux amd64 deployment binary.
- `make publish`: check, build, upload, back up, and atomically replace the production binary. It never restarts or stops the service.

Use pnpm for frontend dependencies and commit changes to `web/pnpm-lock.yaml`.

## Production Release Boundary

Production updates must follow `DEPLOY.md`. Automation and assistants may only build and upload the binary, preserve `gosee.bak`, atomically replace `gosee`, and verify file checksums. The service operator owns all process and container lifecycle actions. Do not run restart, stop, kill, Docker lifecycle, or systemd lifecycle commands unless the user explicitly requests that separate action.

## Coding Style & Naming Conventions

Format Go files with `gofmt`; use lowercase package names and descriptive filenames such as `alert_event.go`. Exported Go identifiers use PascalCase; unexported identifiers use camelCase. Keep handlers thin, business rules in services, and database access in repositories. Frontend code uses TypeScript, two-space indentation, PascalCase component filenames (`ServerForm.vue`), and camelCase composables (`usePolling.ts`). Follow existing API response and error-handling patterns.

## Testing Guidelines

The repository currently has no project-owned automated tests. Add Go tests beside the implementation as `*_test.go`, favoring table-driven cases for services, utilities, and handlers. Add frontend tests as `*.spec.ts` when a test runner is introduced. Before submitting, run `go test ./...` and `make check`; also run `make all` for changes affecting frontend embedding or deployment.

## Commit & Pull Request Guidelines

Follow the existing Conventional Commit style: `feat:`, `fix:`, `build:`, and `docs:`; an optional scope is encouraged, for example `fix(deploy): support older systemd`. Keep each commit focused. Pull requests should explain the change, verification commands, configuration or migration impact, and linked issues. Include screenshots for visible UI changes and never commit databases, generated binaries, credentials, or production secrets.

## Security & Configuration

Do not reuse development defaults in production. Replace `jwt.secret` and `security.encryption_key`, keep SSH credentials out of logs, and treat `configs/config.prod.yaml` as a template rather than a secret store.
