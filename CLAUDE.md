# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

Task Timer is a PWA written entirely in Go using [go-app v11](https://github.com/maxence-charriere/go-app). The same `main` package is compiled twice: to WebAssembly (the client UI) and to a native binary (the HTTP server that serves the PWA shell and `web/` assets). Styling is Tailwind CSS v4 (utility classes inline in Go code). Persistence is browser IndexedDB — there is no backend API or server-side storage.

Requires Go 1.27+ (uses the stdlib `uuid` package for UUIDv7 IDs) and pnpm.

## Commands

```bash
pnpm install          # Tailwind CLI dependencies (needed before first build)
make build            # web/app.wasm (GOOS=js GOARCH=wasm) + bin/task-timer + web/styles.css
make run              # live-reload dev server via `go tool air` on http://localhost:8080
make test             # go test ./...
go test ./components -run TestParseDurationInput   # single test
docker build -t task-timer . && docker run -p 8080:8080 task-timer
```

Air only watches `.go/.tpl/.tmpl/.html` files — editing `styles/main.css` alone does not trigger a rebuild. `web/app.wasm`, `web/*.css`, and `bin/` are build outputs (gitignored); the app won't work until `make build` has produced them.

## Architecture

**Entry point (`main.go`)**: registers the single route `/` → `pages.Home`, calls `app.RunWhenOnBrowser()` (which takes over when running as WASM and never returns), and otherwise serves the `app.Handler` on `:8080`. PWA metadata/icons are configured there.

**State ownership**: `pages.Home` owns the task list and the IndexedDB handle. `components.Task` owns per-task runtime state (ticker, running flag, displayed duration). They communicate through go-app **actions**, not direct calls:

- `saveTask` (value: `models.Task`) — Task → Home; Home persists via `Put`.
- `deleteTask` (value: task id) — Task → Home; Home deletes and reloads the list.
- `stopOtherTasks` (value: active task id) — broadcast by Home when adding a task and by a Task on resume; every other running Task stops itself and emits `saveTask`. This enforces the single-active-timer rule.

**Task time model** (`models.Task`): `StartUnix` is Unix seconds; `Duration` is a `time.Duration` (nanoseconds) stored as int64. `Duration == 0` means the task is *running* (elapsed = now − StartUnix); a non-zero `Duration` means stopped with that accumulated time. Resuming rewrites `StartUnix = now − elapsed` and sets `Duration = 0`. Keep this invariant when changing timer logic.

**`CompoID()` on `components.Task`**: implements `app.DismountEnforcer` so go-app remounts (rather than reuses) a Task component when the task at a list index changes, e.g. after deleting. Without it, `OnMount` state (ticker, running flag) leaks between tasks.

**`pkg/indexeddb`**: a thin wrapper over the browser IndexedDB API via `app.Value`. Every call (`Open`, `Add`, `Put`, `Delete`, `Get`, `GetAll`) blocks on a channel until the JS success/error callback fires, so these must be called inside `ctx.Async(...)` and results applied back with `ctx.Dispatch(...)` — see `pages/home_page.go`. Records are plain `map[string]any` with lowercase keys (`id`, `name`, `description`, `startUnix`, `duration`, `loggedUnix`; `description`/`loggedUnix` may be missing on older records); DB `task_timer` v2 with object stores `tasks` and `settings`, both keyed by `id` (the upgrade handler creates whichever is missing via `HasObjectStore`, so bump the version and add a guarded `CreateObjectStore` for new stores). Mapping between records and `models.Task` is done by hand in `home_page.go` (`taskRecord` / `loadTasks`).

**Jira integration (`pkg/jira`, optional)**: configured per user in the browser, not on the server. The header "Jira" button opens `components.JiraSettingsDialog`; `pages.Home` owns the `jira.Config` (base URL, email, API token), persists it in the `settings` object store (record `id: "jira"`, keys `baseUrl`, `email`, `apiToken`; deleted on disconnect) and passes it to each `components.Task` as the `JiraConfig` prop, which shows the "Add worklog" button when `IsSet()`. Jira Cloud rejects cross-origin API-token requests, so the native server always registers stateless proxies `POST /api/jira/worklog` → `/rest/api/3/issue/{key}/worklog` (Basic auth, ADF comment) and `POST /api/jira/test` → `/rest/api/3/myself` (the dialog's "Test connection"); the client posts a `jira.ProxyRequest` (config + worklog) or a bare `jira.Config` via `postJiraProxy` (`components/jira_proxy.go`) with `net/http` (fetch) inside `ctx.Async`. `jira.NormalizeConfig` validates the config on both sides and only allows `https://<site>.atlassian.net` / `.jira.com`, so the proxy can't be used to reach arbitrary hosts — keep that restriction. On success the task's `LoggedUnix` (IndexedDB key `loggedUnix`, may be missing) is set and a LOGGED badge shown.

**Duration input parsing**: `parseDurationInput` / `formatDurationTemplate` in `components/edit_dialog_component.go` implement the edit-dialog formats (`1h 15m`, `1:30`, `1:30:00`, `1 hour 30 mins`, plain number = minutes). These pure helpers are what `components/task_component_test.go` covers.
