package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ejamesc/goblawg"
	"github.com/gorilla/mux"
)

type postPresenter struct {
	*goblawg.Blog
	TitleValue string
	BodyValue  string
	Flashes    []interface{}
}

func (a *App) newPostDisplayHandler(rw http.ResponseWriter, req *http.Request) {
	presenter := &postPresenter{
		a.blog,
		"",
		"",
		nil,
	}
	a.rndr.HTML(rw, http.StatusOK, "newpost", presenter)
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
		fs := session.Flashes()
		session.Save(req, rw)

		presenter := &postPresenter{
			a.blog,
			string(post.Body),
			post.Title,
			fs,
		}
		a.rndr.HTML(rw, http.StatusOK, "newpost", presenter)
		return
	}

	http.Redirect(rw, req, "/admin", http.StatusFound)
}

func (a *App) editPostDisplayHandler(rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	link := vars["link"]
	post := a.blog.GetPostByLink(link)

	if post == nil {
		http.Error(rw, "No such post exists", http.StatusNotFound)
		return
	}

	presenter := struct {
		*goblawg.Post
		Body     string
		BlogLink string
		Name     string
	}{
		post,
		string(post.Body),
		a.blog.Link,
		a.blog.Name,
	}

	a.rndr.HTML(rw, http.StatusOK, "edit", presenter)
}

func (a *App) editPostHandler(rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	link := vars["link"]
	post := a.blog.GetPostByLink(link)

	if post == nil {
		http.Error(rw, "No such post exists", http.StatusNotFound)
		return
	}

	post.Body = []byte(req.FormValue("body"))
	post.Title = req.FormValue("title")

	isDraft := false
	if req.FormValue("draft") == "true" {
		isDraft = true
	}
	post.IsDraft = isDraft

	// Change this later, to check from form
	// post.Time = time.Now()

	post.LastModified = time.Now()

	err := a.blog.EditPost(post)
	session, _ := a.store.Get(req, "session")
	if err != nil {
		session.AddFlash(fmt.Sprintf("Unable to save post: %s", err))
		fs := session.Flashes()
		session.Save(req, rw)

		presenter := struct {
			*goblawg.Post
			*goblawg.Blog
			Flashes []interface{}
		}{
			post,
			a.blog,
			fs,
		}

		a.rndr.HTML(rw, http.StatusOK, "edit", presenter)
		return
	}

	session.AddFlash(fmt.Sprintf("%s was successfully edited.", post.Title))
	session.Save(req, rw)
	http.Redirect(rw, req, "/admin", http.StatusFound)
}
