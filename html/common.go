package html

import (
	"crypto/sha256"
	"embed"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	g "github.com/maragudk/gomponents"
	c "github.com/maragudk/gomponents/components"
	. "github.com/maragudk/gomponents/html"
)

//go:embed public
var public embed.FS

type PageProps struct {
	Title       string
	Description string
}

var hashOnce sync.Once
var appCSSPath string

func Page(p PageProps, body ...g.Node) g.Node {
	hashOnce.Do(func() {
		appCSSPath = getHashedPath(public, "public/styles/app.css")
	})

	return c.HTML5(c.HTML5Props{
		Title:       p.Title,
		Description: p.Description,
		Language:    "en",
		Head: []g.Node{
			Link(Rel("stylesheet"), Href(appCSSPath)),
		},
		Body: []g.Node{
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
	return Div(Class("prose prose-lg lg:prose-xl xl:prose-2xl prose-indigo"), g.Group(children))
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

func getHashedPath(fs embed.FS, path string) string {
	data, err := fs.ReadFile(path)
	if err != nil {
		panic(err)
	}
	path = strings.TrimPrefix(path, "public/")
	ext := filepath.Ext(path)
	if ext == "" {
		panic("no extension found")
	}
	return fmt.Sprintf("%v.%x%v", strings.TrimSuffix(path, ext), sha256.Sum256(data), ext)
}
