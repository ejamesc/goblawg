package goblawg

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"

	"github.com/gorilla/feeds"
)

/* Contains all the generator functions for Blog */
func (b *Blog) GenerateSite() []error {
	errors := []error{nil, nil, nil}

	pErr := b.GeneratePostsWithTemplate("essay.html")
	if pErr != nil {
		errors[0] = pErr
	}

	rssErr := b.GenerateRSS()
	if rssErr != nil {
		errors[1] = rssErr
	}

	pageErr := b.GenerateSitePages()
	if pageErr != nil {
		errors[2] = pageErr
	}

	b.WriteInfoJSON()

	return errors
}

func (b *Blog) GeneratePostsWithTemplate(tmpl string) error {

	for _, post := range b.Posts {
		filepath := post.Link
		filepath = path.Join(b.OutDir, filepath)

		_, err := os.Stat(filepath)
		// If it's now a draft, and the generated post exists, delete it
		if post.IsDraft {
			if err == nil {
				os.RemoveAll(filepath)
			}
			continue
		}

		// The directory doesn't yet exist
		if err != nil && os.IsNotExist(err) {
			dirErr := os.Mkdir(filepath, 0776)
			if dirErr != nil {
				return dirErr
			}
		}

		// Generate the HTML and write to file
		if b.LastModified.Before(post.LastModified) || b.LastModified.Equal(post.LastModified) {
			filepath = path.Join(filepath, "index.html")
			file, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE, 0776)
			defer file.Close()
			if err != nil {
				return err
			}

			t, err := template.ParseFiles(tmpl)
			if err != nil {
				return err
			}

			pr := struct {
				Title string
				Body  template.HTML
				Time  time.Time
			}{post.Title, template.HTML(post.Body), post.Time}

			t.Execute(file, pr)
		}
	}

	b.LastModified = time.Now()
	return nil
}

// Generate the RSS feed
func (b *Blog) GenerateRSS() error {
	feed := &feeds.Feed{
		Title:       b.Name,
		Link:        &feeds.Link{Href: b.Link},
		Description: b.Description,
		Author:      &feeds.Author{b.Author, b.Email},
		Created:     time.Now(),
	}

	feed.Items = []*feeds.Item{}
	for _, p := range b.GetPublishedPosts() {
		desc := string(p.Body)
		if len(desc) > 120 {
			desc = desc[:120] + "..."
		}
		f := &feeds.Item{
			Title:       p.Title,
			Link:        &feeds.Link{Href: b.Link + "/" + p.Link + "/"},
			Description: desc,
			Created:     p.Time,
		}
		feed.Items = append(feed.Items, f)
	}

	rss, err := feed.ToRss()
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path.Join(b.OutDir, "feed.rss"), []byte(rss), 0776)
	if err != nil {
		return err
	}

	b.LastModified = time.Now()
	return nil
}

// Generate the rest of the templates that isn't the blog
func (b *Blog) GenerateSitePages() error {
	fil, err := ioutil.ReadDir(b.InDir)
	if err != nil {
		return err
	}

	for _, fi := range fil {
		if path.Ext(fi.Name()) == ".html" {
			t, err := template.ParseFiles(path.Join(b.InDir, fi.Name()))
			if err != nil {
				return err
			}

			name := strings.Split(fi.Name(), ".")
			if len(name) != 2 {
				return fmt.Errorf("%s is a bad filename, expected x.html", fi.Name())
			}

			oPath := path.Join(b.OutDir, name[0], "index.html")
			f, err := os.OpenFile(oPath, os.O_RDWR|os.O_CREATE, 0776)
			if err != nil {
				return err
			}

			t.Execute(f, b)
		}
	}

	b.LastModified = time.Now()
	return nil
}
