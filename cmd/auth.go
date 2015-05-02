package main

import (
	"net/http"

	"github.com/ejamesc/goblawg"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

func (a *App) loginHandler(rw http.ResponseWriter, req *http.Request) {
	if a.checkLogin(req) {
		http.Redirect(rw, req, "admin", http.StatusFound)
		return
	}

	session, _ := a.store.Get(req, "session")
	fs := session.Flashes()
	session.Save(req, rw)

	presenter := struct {
		*goblawg.Blog
		Flashes []interface{}
	}{
		a.blog,
		fs,
	}
	a.rndr.HTML(rw, http.StatusOK, "login", presenter)
}

func (a *App) loginPostHandler(rw http.ResponseWriter, req *http.Request) {
	name := req.FormValue("name")
	pass := req.FormValue("password")
	session, _ := a.store.Get(req, "session")

	redirURL, err := a.router.Get("login").URL()
	if err != nil {
		a.Printf("Problem generating link, %v", err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
	redirectTarget := redirURL.String()

	if name == username && pass == password {
		session.Values["username"] = username
		session.Save(req, rw)
		redirectTarget = "/admin"
	} else {
		session.AddFlash("Wrong username or password!")
		session.Save(req, rw)
	}
	http.Redirect(rw, req, redirectTarget, http.StatusFound)
}

func (a *App) logoutHandler(rw http.ResponseWriter, req *http.Request) {
	session, _ := a.store.Get(req, "session")
	delete(session.Values, "username")
	session.Save(req, rw)
	redirectTarget, err := a.router.Get("login").URL()
	if err != nil {
		a.Printf("Problem generating link, %v", err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(rw, req, redirectTarget.String(), http.StatusFound)
}

// Auth Middleware
func (a *App) authMiddleware(store *sessions.CookieStore, r *mux.Router) func(http.ResponseWriter, *http.Request, http.HandlerFunc) {
	return func(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
		if a.checkLogin(req) {
			next(w, req)
		} else {
			r, _ := r.Get("login").URL()
			http.Redirect(w, req, r.String(), http.StatusFound)
		}
	}
}
