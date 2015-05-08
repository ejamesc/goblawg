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
	"github.com/kardianos/osext"
)

/* Contains all the generator functions for Blog */
func (b *Blog) GenerateSite() []error {
	errors := []error{}

	pErr := b.GeneratePostsWithTemplate("essay.html")
	if pErr != nil {
		errors = append(errors, pErr)
	}

	rssErr := b.GenerateRSS()
	if rssErr != nil {
		errors = append(errors, rssErr)
	}

	pageErrors := b.GenerateSitePages()
	if pageErrors != nil {
		errors = append(errors, pageErrors...)
	}

	b.WriteInfoJSON()

	return errors
}

func (b *Blog) GeneratePostsWithTemplate(tmpl string) error {

	funcMap := template.FuncMap{
		"fdate": DateFmt,
		"md":    Markdown,
	}

	extDir, _ := osext.ExecutableFolder()
	tmpDir := path.Join(extDir, "../src/github.com/ejamesc/goblawg", "templates")
	t := template.Must(template.New("essay.html").Funcs(funcMap).ParseFiles(path.Join(tmpDir, tmpl)))

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

			err = t.Execute(file, post)
			if err != nil {
				return err
			}
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
func (b *Blog) GenerateSitePages() []error {
	errors := []error{}
	fil, err := ioutil.ReadDir(b.InDir)
	if err != nil {
		errors = append(errors, err)
		return errors
	}

	for _, fi := range fil {
		if path.Ext(fi.Name()) == ".html" {
			t, err := template.ParseFiles(path.Join(b.InDir, fi.Name()))
			if err != nil {
				errors = append(errors, err)
			}

			name := strings.Split(fi.Name(), ".")
			if len(name) != 2 {
				errors = append(errors, fmt.Errorf("%s is a bad filename, expected x.html", fi.Name()))
			}

			oPath := path.Join(b.OutDir, name[0], "index.html")
			f, err := os.OpenFile(oPath, os.O_RDWR|os.O_CREATE, 0776)
			if err != nil {
				errors = append(errors, err)
			}

			t.Execute(f, b)
		}

		return errors
	}

	b.LastModified = time.Now()
	return nil
}
