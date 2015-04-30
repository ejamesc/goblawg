package main

import "net/http"

/* Contains all the helpers
* */
func (a *App) checkLogin(req *http.Request) bool {
	session, _ := a.store.Get(req, "session")
	usr := session.Values["username"]
	if usr == username {
		return true
	} else {
		return false
	}
}

func (a *App) getFlashes(rw http.ResponseWriter, req *http.Request) []interface{} {
	session, _ := a.store.Get(req, "session")
	fs := session.Flashes()
	session.Save(req, rw)
	return fs
}
