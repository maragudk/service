package html

import (
	g "github.com/maragudk/gomponents"
	"github.com/maragudk/gomponents-heroicons/v2/solid"
	. "github.com/maragudk/gomponents/html"
)

func HomePage() g.Node {
	return Page(PageProps{Title: "Service", Description: "This is a service."},

		Div(Class("prose-headings:font-serif"),
			H1(Class("inline-flex items-center"), solid.Sparkles(Class("h-12 w-12 mr-2")), g.Text(`Service`)),

			P(Class("lead"), g.Raw(`Hi! 🤓 This is a service template in Go.`)),

			P(g.Raw(`<a href="https://github.com/maragudk/service">Check out the source code on Github</a>. It’s nice.`)),
		),
	)
}
