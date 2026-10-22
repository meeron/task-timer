package components

import "github.com/maxence-charriere/go-app/v11/pkg/app"

// message renders a paragraph that is hidden while msg is empty. Unlike
// app.If it always occupies its place in the parent, so showing or hiding it
// does not shift (and make go-app replace) the elements that follow, which
// would reset input focus and silently drop Dispatch calls made through
// their event contexts.
func message(class, msg string) app.UI {
	if msg == "" {
		class += " hidden"
	}
	return app.P().Class(class).Text(msg)
}
