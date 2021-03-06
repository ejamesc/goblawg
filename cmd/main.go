package main

import (
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"path"
	"time"

	"github.com/codegangsta/negroni"
	"github.com/ejamesc/goblawg"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/kardianos/osext"
	"github.com/unrolled/render"
)

var extDir, _ = osext.ExecutableFolder()
var currDir = path.Join(extDir, "../src/github.com/ejamesc/goblawg")

type App struct {
	rndr   *render.Render
	router *mux.Router
	store  *sessions.CookieStore
	blog   *goblawg.Blog
	*log.Logger
}

func main() {
	// We'll use random numbers to get random posts
	rand.Seed(time.Now().UTC().UnixNano())

	b, err := goblawg.NewBlog(path.Join(currDir, "settings.json"))
	if err != nil {
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
				"fdate":        goblawg.DateFmt,
				"md":           goblawg.Markdown,
				"subslice":     goblawg.Subslice,
				"is_even":      goblawg.IsEven,
				"random_posts": goblawg.RandomPosts,
			},
		},
	})

	a := &App{
		rndr,
		r,
		store,
		b,
		b.Logger,
	}

	admin := mux.NewRouter().StrictSlash(true)
	admin.HandleFunc("/admin", a.adminPageHandler).Name("admin")

	r.PathPrefix("/admin").Handler(
		negroni.New(
			negroni.HandlerFunc(a.authMiddleware(store, r)),
			negroni.Wrap(admin),
		))

	ar := admin.PathPrefix("/admin").Subrouter()
	ar.HandleFunc("/new", a.newPostHandler).Methods("POST")
	ar.HandleFunc("/new", a.newPostDisplayHandler).Methods("GET")
	ar.HandleFunc("/edit/{link}", a.editPostDisplayHandler).Methods("GET")
	ar.HandleFunc("/edit/{link}", a.editPostHandler).Methods("POST")
	ar.HandleFunc("/delete/{link}", a.deletePostHandler).Methods("DELETE")
	ar.HandleFunc("/regen", a.generateHandler).Methods("POST")

	r.HandleFunc("/", a.loginHandler).Methods("GET").Name("login")
	r.HandleFunc("/", a.loginPostHandler).Methods("POST").Name("login")
	r.HandleFunc("/logout", a.logoutHandler).Methods("POST").Name("logout")

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/",
		http.FileServer(http.Dir(path.Join(currDir, "static")))))

	n := standardMiddleware()
	n.UseHandler(r)
	n.Run(":3000")
}

func (a *App) adminPageHandler(rw http.ResponseWriter, req *http.Request) {
	fs := a.getFlashes(rw, req)
	presenter := struct {
		*goblawg.Blog
		Posts   []*goblawg.Post
		Flashes []interface{}
	}{a.blog, a.blog.GetAllPosts(), fs}
	a.rndr.HTML(rw, http.StatusOK, "admin", presenter)
}

/* Middleware
* */
func standardMiddleware() *negroni.Negroni {
	return negroni.New(
		negroni.NewRecovery(),
		negroni.NewLogger())
}
