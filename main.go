package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/maxence-charriere/go-app/v11/pkg/app"
	"github.com/meeron/task-timer/pages"
)

func main() {
	app.Route("/", func() app.Composer { return &pages.Home{} })
	app.RunWhenOnBrowser()

	// Standard HTTP routing (server-side):
	http.Handle("/", &app.Handler{
		Name:            "Task Timer",
		Title:           "Task Timer",
		Description:     "Measure your tasks time",
		Styles:          []string{"web/styles.css"},
		BackgroundColor: "#f8fafc",
		ThemeColor:      "#4f46e5",
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
