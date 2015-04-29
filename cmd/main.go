package main

import (
	"fmt"
	"html/template"
	"path"

	"github.com/ejamesc/goblawg"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/kardianos/osext"
	"github.com/russross/blackfriday"
	"github.com/unrolled/render"
)

var extDir, _ = osext.ExecutableFolder()
var currDir = path.Join(extDir, "../src/github.com/ejamesc/metacog")

type App struct {
	rndr   *render.Render
	router *mux.Router
	store  *sessions.CookieStore
	blog   *goblawg.Blog
}

func main() {
	b, err := goblawg.NewBlog(path.Join(currDir, "settings.json"))
	if err != nil {
		fmt.Println("An error with settings.json")
		panic(err)
	}

	store := sessions.NewCookieStore([]byte(secret_key))
	r := mux.NewRouter()
	rndr := render.New(render.Options{
		Directory:  path.Join(currDir, "templates"),
		Extensions: []string{".html"},
		Layout:     "base",
		Funcs: []template.FuncMap{
			template.FuncMap{
				"md": markdown,
			},
		},
	})

	a := &App{
		rndr,
		r,
		store,
		b,
	}

	fmt.Println(a)
}

// Template functions
func markdown(input []byte) string {
	output := blackfriday.MarkdownCommon(input)
	return string(output)
}
