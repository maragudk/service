package html

import (
	g "github.com/maragudk/gomponents"
	. "github.com/maragudk/gomponents/html"
)

func HomePage() g.Node {
	return Page(PageProps{Title: "Service", Description: "This is a service."},

		H1(g.Text(`Service`)),
		P(g.Raw(`Hi! 🤓`)),
	)
}
