# Task Timer ⏱️

A modern, lightweight Progressive Web Application (PWA) for tracking task time effortlessly. Built with **Go** compiled to **WebAssembly (WASM)** using [go-app](https://go-app.dev/) and styled with **Tailwind CSS v4**.

---

## Features

- ⏱️ **Real-Time Tracking**: Live timers with active status indicators (running vs. stopped) and formatted duration displays (`HHh MMm SSs` / `MMm SSs`).
- ⏯️ **Pause & Resume**: Stop and resume tasks anytime without losing accumulated time.
- ✏️ **Edit Task**: Rename a task and adjust or set its timer duration directly from the task card. Accepts multiple input formats (see [Edit Timer formats](#edit-timer-formats)).
- 🔂 **Single Active Timer**: Starting or resuming a task automatically stops any other running task, keeping you focused on one thing at a time.
- 🗑️ **Delete with Confirmation**: Remove tasks through a confirmation dialog to prevent accidental deletions.
- 📝 **Jira Worklogs** *(optional)*: Log a task's time and description as a worklog on a Jira Cloud issue (task name = issue key). See [Jira Integration](#jira-integration).
- 💾 **Local Persistence**: Tasks are persisted across sessions in the browser's **IndexedDB** via a lightweight custom Go wrapper (`pkg/indexeddb`).
- 📱 **Progressive Web App (PWA)**: Installable, responsive, with built-in app update notifications.
- 🎨 **Clean UI**: Crafted with modern Tailwind CSS v4 styling, badges, and responsive layouts.

---

## Tech Stack

| Layer           | Technology                                                                                     |
| --------------- | ---------------------------------------------------------------------------------------------- |
| Language        | [Go 1.27](https://go.dev/) — compiled to WebAssembly (client) and native binary (server)       |
| Web Framework   | [go-app v11](https://github.com/maxence-charriere/go-app) — PWA framework for Go & WebAssembly |
| Styling         | [Tailwind CSS v4](https://tailwindcss.com/)                                                    |
| Package Manager | [pnpm](https://pnpm.io/)                                                                       |
| UUID            | [`uuid`](https://go.dev/pkg/uuid) *(Go 1.27 stdlib)* — UUIDv7 task ID generation              |
| Live Reload     | [Air](https://github.com/air-verse/air) — hot-rebuild dev server (via `go tool air`)           |
| Persistence     | Browser **IndexedDB** — custom Go/WASM wrapper in `pkg/indexeddb`                              |

---

## Prerequisites

Ensure the following tools are installed:

- **Go** (1.27+)
- **Node.js** & **pnpm** (e.g. pnpm 9+)
- **Make** (optional, for running Makefile targets)
- **Docker** (optional, for containerized deployment)

---

## Getting Started

### 1. Clone the Repository

```bash
git clone https://github.com/meeron/task-timer.git
cd task-timer
```

### 2. Install Frontend Dependencies

Install the required Tailwind CSS dependencies:

```bash
pnpm install
```

### 3. Build the Application

You can build the WebAssembly binary, Tailwind stylesheet, and backend server using `make`:

```bash
make build
```

### 4. Run the Application

Start the development server with **live-reload** via [Air](https://github.com/air-verse/air):

```bash
make run
```

Air watches for `.go`, `.tpl`, `.tmpl`, and `.html` file changes and automatically rebuilds the WASM binary, Tailwind stylesheet, and server binary before restarting.

Open your browser at:

```
http://localhost:8080
```

---

## Docker

A multi-stage `Dockerfile` is included for containerized builds and deployment.

### Build the image

```bash
docker build -t task-timer .
```

### Run the container

```bash
docker run -p 8080:8080 task-timer
```

The app will be available at `http://localhost:8080`.

> The image uses a minimal Alpine runtime with a statically linked binary — no Go or Node.js required at runtime.

---

## Jira Integration

The integration is optional. It is enabled only when **all three** environment variables are set on the server; otherwise the **Add worklog** button is hidden and the proxy endpoint is not registered. If only some are set, a warning is logged at startup.

| Variable         | Example                         |
| ---------------- | ------------------------------- |
| `JIRA_BASE_URL`  | `https://your-org.atlassian.net` |
| `JIRA_EMAIL`     | `you@example.com`               |
| `JIRA_API_TOKEN` | [Atlassian API token](https://id.atlassian.com/manage-profile/security/api-tokens) |

For local development, put them in a `.env` file in the project root (gitignored and excluded from Docker builds). Real environment variables take precedence over `.env`, and Air restarts the server when `.env` changes:

```bash
cp .env.example .env   # then fill in JIRA_API_TOKEN
make run
```

In Docker, pass them as environment variables:

```bash
docker run -p 8080:8080 -e JIRA_BASE_URL=... -e JIRA_EMAIL=... -e JIRA_API_TOKEN=... task-timer
```

Name a task after its issue key (e.g. `ABC-123`) and click **Add worklog**. A running task is stopped first; the dialog is prefilled with the key, description and elapsed time (rounded to whole minutes). The browser posts to the server's `/api/jira/worklog` proxy, which calls the Jira Cloud REST API v3 — the API token never reaches the browser. Logged tasks get a **LOGGED** badge.

> The enabled flag is baked into the cached PWA shell, so after changing the variables and restarting the server, use the in-app *Update now* banner (or reload) to pick it up.

---

## Edit Timer Formats

When editing a task's elapsed time, the input field accepts several flexible formats:

| Format           | Example          | Description                                      |
| ---------------- | ---------------- | ------------------------------------------------ |
| Template         | `1h 15m`         | Hours and/or minutes with `h`/`m` suffixes       |
| Colon (HH:MM)    | `1:30`           | Hours and minutes separated by `:`               |
| Colon (HH:MM:SS) | `1:30:00`        | Hours, minutes, and seconds                      |
| Natural language | `1 hour 30 mins` | Long-form words like `hour`, `minute`, `seconds` |
| Plain number     | `45`             | Treated as minutes                               |

Quick-select presets (`15m`, `30m`, `45m`, `1h`, `1h 15m`, `2h`) are also available in the edit dialog.

---

## Running Tests

```bash
make test
# or
go test ./...
```

---

## Project Structure

```text
task-timer/
├── bin/                           # Compiled server binaries
├── components/                    # Reusable go-app UI components
│   ├── task_component.go          # Task card with timer controls and real-time ticker
│   ├── task_component_test.go     # Unit tests for task component helpers
│   ├── edit_dialog_component.go   # Modal dialog for editing timer duration
│   ├── worklog_dialog_component.go # Modal dialog for adding a Jira worklog
│   └── delete_dialog_component.go # Modal dialog for confirming task deletion
├── models/                        # Data structures
│   └── models.go                  # Task model definition
├── pages/                         # Application pages and views
│   └── home_page.go               # Dashboard, task input form, and list management
├── pkg/                           # Internal packages
│   ├── jira/                      # Optional Jira Cloud worklog client + server proxy
│   └── indexeddb/                 # Browser IndexedDB bindings for Go/WASM
│       ├── main.go                # IDBDatabase & IDBObjectStore interfaces + Open()
│       ├── database.go            # idbDatabase implementation (transactions)
│       └── object_store.go        # idbObjectStore implementation (CRUD operations)
├── styles/                        # Source stylesheets
│   └── main.css                   # Tailwind CSS entrypoint
├── web/                           # Static web assets served by go-app
│   ├── app.wasm                   # Compiled WebAssembly binary
│   └── styles.css                 # Compiled Tailwind stylesheet
├── Dockerfile                     # Multi-stage Docker build
├── go.mod                         # Go module definition
├── Makefile                       # Build, run, and test automation
├── package.json                   # Node/pnpm dependencies & scripts
└── main.go                        # Server entrypoint and go-app route configuration
```

---

## License

This project is licensed under the [MIT License](LICENSE).
