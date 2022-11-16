package html

import (
	g "github.com/maragudk/gomponents"
	"github.com/maragudk/gomponents-heroicons/v2/solid"
	. "github.com/maragudk/gomponents/html"
)

func HomePage() g.Node {
	return Page(PageProps{Title: "Service", Description: "This is a service."},

		H1(Class("inline-flex items-center"), solid.Sparkles(Class("h-12 w-12 mr-2")), g.Text(`Service`)),
		P(g.Raw(`Hi! 🤓`)),
	)
}
