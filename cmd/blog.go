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
	ShortValue string
	IsDraft    bool
	Flashes    []interface{}
}

func (a *App) newPostDisplayHandler(rw http.ResponseWriter, req *http.Request) {
	presenter := &postPresenter{
		a.blog,
		"",
		"",
		"",
		false,
		nil,
	}
	a.rndr.HTML(rw, http.StatusOK, "newpost", presenter)
}

func (a *App) newPostHandler(rw http.ResponseWriter, req *http.Request) {
	session, _ := a.store.Get(req, "session")
	title, body, short := req.FormValue("title"), req.FormValue("body"), req.FormValue("short")
	if title == "" || body == "" {
		session.AddFlash("Post title or body can't be left empty!")
		session.Save(req, rw)

		fs := a.getFlashes(rw, req)
		presenter := &postPresenter{
			a.blog,
			title,
			body,
			short,
			false,
			fs,
		}
		a.rndr.HTML(rw, http.StatusOK, "newpost", presenter)
	}
	post := a.blog.NewPost(title,
		[]byte(body),
		short,
		false,
		time.Now(),
		time.Now())

	if req.FormValue("isdraft") == "on" {
		post.IsDraft = true
	}

	// Change this later
	post.Time = time.Now()

	post.LastModified = time.Now()

	err := a.blog.SavePost(post)
	if err != nil {
		session.AddFlash(fmt.Sprintf("Unable to save post: %s", err))
		fs := session.Flashes()
		session.Save(req, rw)

		presenter := &postPresenter{
			a.blog,
			post.Title,
			string(post.Body),
			post.Short,
			post.IsDraft,
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

	fs := a.getFlashes(rw, req)

	presenter := struct {
		*goblawg.Post
		Body     string
		BlogLink string
		Name     string
		Flashes  []interface{}
	}{
		post,
		string(post.Body),
		a.blog.Link,
		a.blog.Name,
		fs,
	}

	a.rndr.HTML(rw, http.StatusOK, "edit", presenter)
}

func (a *App) editPostHandler(rw http.ResponseWriter, req *http.Request) {
	session, _ := a.store.Get(req, "session")
	vars := mux.Vars(req)
	link := vars["link"]
	oldPost := a.blog.GetPostByLink(link)

	if oldPost == nil {
		http.Error(rw, "No such post exists", http.StatusNotFound)
		return
	}

	title, body, short := req.FormValue("title"), req.FormValue("body"), req.FormValue("short")
	if title == "" || body == "" {
		session.AddFlash("Post title or body can't be left empty!")
		session.Save(req, rw)

		fs := a.getFlashes(rw, req)

		presenter := struct {
			*goblawg.Post
			Title    string
			Body     string
			Short    string
			BlogLink string
			Name     string
			Flashes  []interface{}
		}{
			oldPost,
			title,
			body,
			short,
			a.blog.Link,
			a.blog.Name,
			fs,
		}

		a.rndr.HTML(rw, http.StatusOK, "edit", presenter)
		return
	}
	newPost := a.blog.NewPost(title, []byte(body), short, false, oldPost.Time, time.Now())
	if req.FormValue("isdraft") == "on" {
		newPost.IsDraft = true
	}
	// TODO: Change timestamp later, to check from form

	err := a.blog.EditPost(oldPost, newPost)
	if err != nil {
		session.AddFlash(fmt.Sprintf("Unable to edit post: %s", err))
		session.Save(req, rw)

		http.Redirect(rw, req, "/admin", http.StatusFound)
		return
	}

	session.AddFlash(fmt.Sprintf("%s was successfully edited.", oldPost.Title))
	session.Save(req, rw)
	http.Redirect(rw, req, "/admin", http.StatusFound)
}

func (a *App) deletePostHandler(rw http.ResponseWriter, req *http.Request) {
	link := mux.Vars(req)["link"]
	post := a.blog.GetPostByLink(link)
	err := a.blog.DeletePost(post)

	if err != nil {
		session, _ := a.store.Get(req, "session")
		session.AddFlash(fmt.Sprintf("Unable to delete post: %S", err))
		session.Save(req, rw)
	}

	http.Redirect(rw, req, "/admin", http.StatusFound) // TODO: right status?
}

func (a *App) generateHandler(rw http.ResponseWriter, req *http.Request) {
	session, _ := a.store.Get(req, "session")
	errors := a.blog.GenerateSite()
	noErrors := true
	for _, e := range errors {
		if e == nil {
			continue
		} else {
			session.AddFlash(fmt.Sprintf("Generate site error: %v", e))
			session.Save(req, rw)
			noErrors = false
		}
	}

	if noErrors {
		session.AddFlash("Blog successfully regenerated!")
		session.Save(req, rw)
	}
	http.Redirect(rw, req, "/admin", http.StatusFound)
}
