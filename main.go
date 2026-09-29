package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
	"github.com/meeron/task-timer/pages"
	"github.com/meeron/task-timer/pkg/jira"
)

func main() {
	app.Route("/", func() app.Composer { return &pages.Home{} })
	app.RunWhenOnBrowser()

	// Local development: load variables from .env if present. Variables already
	// set in the environment (e.g. docker run -e) take precedence.
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Printf("Warning: failed to load .env: %v", err)
	}

	// Optional Jira integration: the worklog proxy is registered and the client
	// is told to show the "Add worklog" button only when fully configured.
	env := map[string]string{}
	jiraCfg, err := jira.ConfigFromEnv(os.Getenv)
	switch {
	case err == nil:
		http.Handle("POST "+jira.WorklogPath, jira.NewWorklogHandler(jira.NewClient(jiraCfg)))
		env[jira.EnabledEnvKey] = "true"
		fmt.Printf("Jira integration enabled (%s)\n", jiraCfg.BaseURL)
	case !errors.Is(err, jira.ErrNotConfigured):
		log.Printf("Warning: %v", err)
	}

	// Standard HTTP routing (server-side):
	http.Handle("/", &app.Handler{
		Name:            "Task Timer",
		Title:           "Task Timer",
		Description:     "Measure your tasks time",
		Styles:          []string{"web/styles.css"},
		BackgroundColor: "#f8fafc",
		ThemeColor:      "#4f46e5",
		Env:             env,
		Icon: app.Icon{
			// 192x192 – standard PWA icon (also used as page icon fallback)
			Default: "/web/favicon-192x192.png",
			// 512x512 – high-res PWA icon for splash screens
			Large: "/web/favicon-512x512.png",
			// 512x512 – adaptive/maskable icon for Apple & Android home screens
			Maskable: "/web/favicon-512x512.png",
		},
		RawHeaders: []string{
			// Classic favicon for legacy browsers
			`<link rel="shortcut icon" href="/web/favicon.ico">`,

			// Standard browser favicon (32x32 shown in tabs, 16x16 in bookmarks)
			`<link rel="icon" type="image/png" sizes="16x16" href="/web/favicon-16x16.png">`,
			`<link rel="icon" type="image/png" sizes="32x32" href="/web/favicon-32x32.png">`,
			`<link rel="icon" type="image/png" sizes="96x96" href="/web/favicon-96x96.png">`,

			// Apple Touch Icons (used when added to iOS/macOS home screen)
			`<link rel="apple-touch-icon" sizes="57x57"   href="/web/favicon-57x57.png">`,
			`<link rel="apple-touch-icon" sizes="60x60"   href="/web/favicon-60x60.png">`,
			`<link rel="apple-touch-icon" sizes="72x72"   href="/web/favicon-72x72.png">`,
			`<link rel="apple-touch-icon" sizes="76x76"   href="/web/favicon-76x76.png">`,
			`<link rel="apple-touch-icon" sizes="114x114" href="/web/favicon-114x114.png">`,
			`<link rel="apple-touch-icon" sizes="120x120" href="/web/favicon-120x120.png">`,
			`<link rel="apple-touch-icon" sizes="144x144" href="/web/favicon-144x144.png">`,
			`<link rel="apple-touch-icon" sizes="152x152" href="/web/favicon-152x152.png">`,
			`<link rel="apple-touch-icon" sizes="180x180" href="/web/favicon-180x180.png">`,

			// Windows 8/10 Start Screen tile
			`<meta name="msapplication-TileImage" content="/web/favicon-144x144.png">`,
			`<meta name="msapplication-TileColor" content="#4f46e5">`,
		},
	})

	fmt.Println("Listening on port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
