# Task Timer ⏱️

A modern, lightweight Progressive Web Application (PWA) for tracking task time effortlessly. Built with **Go** compiled to **WebAssembly (WASM)** using [go-app](https://go-app.dev/) and styled with **Tailwind CSS v4**.

---

## Features

- ⏱️ **Real-Time Tracking**: Live timers with active status indicators (running vs. paused) and formatted duration displays (`HHh MMm SSs` / `MMm SSs`).
- ⏯️ **Pause & Resume**: Stop and resume tasks anytime without losing accumulated time.
- ➕ **Add Extra Time**: Quickly adjust or add extra minutes to any task on the fly.
- 💾 **Local Persistence**: Tasks are persisted across sessions directly in browser `localStorage`.
- 📱 **Progressive Web App (PWA)**: Installable, responsive, with built-in app update notifications.
- 🎨 **Clean UI**: Crafted with modern Tailwind CSS v4 styling, badges, and responsive layouts.

---

## Tech Stack

- **Language**: [Go](https://go.dev/) (compiled to WebAssembly for the client, native binary for the static server)
- **Web Framework**: [go-app (v11)](https://github.com/maxence-charriere/go-app) — Progressive web app framework for Go & WebAssembly
- **Styling**: [Tailwind CSS v4](https://tailwindcss.com/)
- **Package Manager**: [pnpm](https://pnpm.io/)

---

## Prerequisites

Ensure the following tools are installed:

- **Go** (1.22+ recommended)
- **Node.js** & **pnpm** (e.g. pnpm 9+)
- **Make** (optional, for running Makefile targets)

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

Or execute the steps manually:

```bash
# 1. Compile the Go app to WebAssembly
GOARCH=wasm GOOS=js go build -o web/app.wasm

# 2. Compile Tailwind CSS
pnpm exec tailwindcss -i styles/main.css -o web/styles.css

# 3. Build the server binary
go build -o bin/task-timer
```

> **Note for Windows PowerShell users:**
> ```powershell
> $env:GOARCH="wasm"; $env:GOOS="js"; go build -o web/app.wasm; Remove-Item Env:\GOARCH, Env:\GOOS
> pnpm exec tailwindcss -i styles/main.css -o web/styles.css
> go build -o bin/task-timer.exe
> ```

### 4. Run the Application

Start the server using `make`:

```bash
make run
```

Or execute the binary directly:

```bash
./bin/task-timer
# On Windows:
# .\bin\task-timer.exe
```

Open your browser at:
```
http://localhost:8080
```

---

## Project Structure

```text
task-timer/
├── bin/                  # Compiled server binaries
├── components/           # Reusable go-app UI components
│   └── task_component.go # Task card with timer controls and real-time ticker
├── models/               # Data structures
│   └── models.go         # Task model definition
├── pages/                # Application pages and views
│   └── home_page.go      # Dashboard, task input form, and list management
├── styles/               # Source stylesheets
│   └── main.css          # Tailwind CSS entrypoint
├── web/                  # Static web assets served by go-app
│   ├── app.wasm          # Compiled WebAssembly binary
│   └── styles.css        # Compiled Tailwind stylesheet
├── go.mod                # Go module definition
├── Makefile              # Build and execution automation
├── package.json          # Node/pnpm dependencies & scripts
└── main.go               # Server entrypoint and go-app route configuration
```

---

## License

This project is licensed under the [MIT License](LICENSE).
