package main

import (
	"fmt"
	"html/template"
	"net/http"
	"path"
	"time"

	"github.com/codegangsta/negroni"
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
	r := mux.NewRouter().StrictSlash(true)
	rndr := render.New(render.Options{
		Directory:  path.Join(currDir, "templates"),
		Extensions: []string{".html"},
		Layout:     "base",
		Funcs: []template.FuncMap{
			template.FuncMap{
				"fdate": dateFmt,
				"md":    markdown,
			},
		},
	})

	a := &App{
		rndr,
		r,
		store,
		b,
	}

	admin := mux.NewRouter().StrictSlash(true)
	admin.HandleFunc("/admin", a.adminPageHandler).Methods("GET").Name("admin-home")
	r.PathPrefix("/admin").Handler(
		negroni.New(
			negroni.HandlerFunc(authMiddleware(store, r)),
			negroni.Wrap(admin),
		))

	r.HandleFunc("/", a.loginHandler).Methods("GET").Name("login")
	r.HandleFunc("/logout", a.logoutHandler).Methods("POST").Name("logout")

	n := standardMiddleware()
	n.UseHandler(r)
	n.Run(":3000")
}

func (a *App) adminPageHandler(w http.ResponseWriter, req *http.Request) {
	a.rndr.HTML(w, http.StatusOK, "admin-front", a.blog)
}

/* Template functions
* */
func markdown(input []byte) string {
	output := blackfriday.MarkdownCommon(input)
	return string(output)
}

func dateFmt(tt time.Time) string {
	const layout = "3:04pm, 2 January 2006"
	return tt.Format(layout)
}

/* Middleware
* */
func standardMiddleware() *negroni.Negroni {
	return negroni.New(
		negroni.NewRecovery(),
		negroni.NewLogger())
}
