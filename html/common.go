package html

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	g "github.com/maragudk/gomponents"
	c "github.com/maragudk/gomponents/components"
	. "github.com/maragudk/gomponents/html"
)

type PageProps struct {
	Title       string
	Description string
}

var hashOnce sync.Once
var appCSSPath string

func Page(p PageProps, body ...g.Node) g.Node {
	hashOnce.Do(func() {
		appCSSPath = getHashedPath("public/styles/app.css")
	})

	return c.HTML5(c.HTML5Props{
		Title:       p.Title,
		Description: p.Description,
		Language:    "en",
		Head: []g.Node{
			Link(Rel("stylesheet"), Href(appCSSPath)),
		},
		Body: []g.Node{Class("dark:bg-gray-900"),
			Container(true,
				Prose(
					g.Group(body),
				),
			),
		},
	})
}

func Container(padY bool, children ...g.Node) g.Node {
	return Div(
		c.Classes{
			"max-w-7xl mx-auto px-4 sm:px-6 lg:px-8": true,
			"py-4 sm:py-6 lg:py-8":                   padY,
		},
		g.Group(children),
	)
}

func Prose(children ...g.Node) g.Node {
	return Div(Class("prose prose-lg lg:prose-xl xl:prose-2xl dark:prose-invert"), g.Group(children))
}

func ErrorPage() g.Node {
	return Page(PageProps{Title: "Something went wrong", Description: "Oh no! 😵"},
		H1(g.Text("Something went wrong")),
		P(g.Text("Oh no! 😵")),
		P(A(Href("/"), g.Text("Back to front."))),
	)
}

func NotFoundPage() g.Node {
	return Page(PageProps{Title: "There's nothing here! 💨", Description: "Just the void."},
		H1(g.Text("There's nothing here! 💨")),
		P(A(Href("/"), g.Text("Back to front."))),
	)
}

func getHashedPath(path string) string {
	externalPath := strings.TrimPrefix(path, "public")
	ext := filepath.Ext(path)
	if ext == "" {
		panic("no extension found")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Sprintf("%v.x%v", strings.TrimSuffix(externalPath, ext), ext)
	}

	return fmt.Sprintf("%v.%x%v", strings.TrimSuffix(externalPath, ext), sha256.Sum256(data), ext)
}
