package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ejamesc/goblawg"
	"github.com/gorilla/mux"
)

func (a *App) newPostDisplayHandler(rw http.ResponseWriter, req *http.Request) {
	a.rndr.HTML(rw, http.StatusOK, "newpost", a.blog)
}

func (a *App) newPostHandler(rw http.ResponseWriter, req *http.Request) {
	post := &goblawg.Post{}
	post.Body = []byte(req.FormValue("body"))
	post.Title = req.FormValue("title")
	post.Link = goblawg.LinkifyTitle(post.Title)

	isDraft := false
	if req.FormValue("draft") == "true" {
		isDraft = true
	}
	post.IsDraft = isDraft

	// Change this later
	post.Time = time.Now()

	post.LastModified = time.Now()

	err := a.blog.SavePost(post)
	// TODO: Change to session to display error.
	if err != nil {
		session, _ := a.store.Get(req, "session")
		session.AddFlash(fmt.Sprintf("Unable to save post: %s", err))
		session.Save(req, rw)

	}

	http.Redirect(rw, req, "/admin", 302)
}

func (a *App) editPostDisplayHandler(rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	link := vars["link"]
	post := a.blog.GetPostByLink(link)

	presenter := struct {
		Name         string
		BlogLink     string
		Title        string
		Body         string
		Link         string
		Time         time.Time
		IsDraft      bool
		LastModified time.Time
	}{
		a.blog.Name,
		a.blog.Link,
		post.Title,
		string(post.Body),
		post.Link,
		post.Time,
		post.IsDraft,
		post.LastModified,
	}

	a.rndr.HTML(rw, http.StatusOK, "edit", presenter)
}

func (a *App) editPostHandler(rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	link := vars["link"]
	post := a.blog.GetPostByLink(link)

	fmt.Println(post)
}
