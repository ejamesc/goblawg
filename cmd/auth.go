package main

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

func (a *App) loginHandler(w http.ResponseWriter, req *http.Request) {
	presenter := struct {
		Name    string
		Flashes []string
	}{a.blog.Name, []string{}}
	a.rndr.HTML(w, http.StatusOK, "login", presenter)
}

func (a *App) loginPostHandler(w http.ResponseWriter, req *http.Request) {
	email := req.FormValue("email")
	pass := req.FormValue("password")
	session, _ := a.store.Get(req, "session")
	redirURL, err := a.router.Get("login").URL()
	if err != nil {
		a.blog.Printf("Problem generating link, %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	redirectTarget := redirURL.String()

	if email == username && pass == password {
		session.Values["username"] = username
		session.Save(req, w)
		redir, err := a.router.Get("admin-front").URL()
		if err != nil {
			a.blog.Printf("Problem generating link, %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		redirectTarget = redir.String()
	} else {
		session.AddFlash("Wrong username or password!")
		session.Save(req, w)
	}
	http.Redirect(w, req, redirectTarget, http.StatusFound)
}

func (a *App) logoutHandler(w http.ResponseWriter, req *http.Request) {
	session, _ := a.store.Get(req, "session")
	delete(session.Values, "username")
	redirectTarget, err := a.router.Get("admin-front").URL()
	if err != nil {
		a.blog.Printf("Problem generating link, %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, req, redirectTarget.String(), http.StatusFound)
}

// Auth Middleware
func authMiddleware(store *sessions.CookieStore, r *mux.Router) func(http.ResponseWriter, *http.Request, http.HandlerFunc) {
	return func(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
		session, _ := store.Get(req, "session")
		username := session.Values["username"]
		if username == "ejames" {
			next(w, req)
		} else {
			r, _ := r.Get("admin-front").URL()
			http.Redirect(w, req, r.String(), http.StatusFound)
		}
	}
}
