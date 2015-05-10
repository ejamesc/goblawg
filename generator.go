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
	"github.com/termie/go-shutil"
)

var funcMap = template.FuncMap{
	"fdate":    DateFmt,
	"md":       Markdown,
	"is_even":  IsEven,
	"subslice": Subslice,
}

/* Contains all the generator functions for Blog */
func (b *Blog) GenerateSite() []error {
	errors := []error{}

	pErrs := b.GeneratePostsWithTemplate("essay.html")
	if len(pErrs) > 0 {
		errors = append(errors, pErrs...)
	}

	rssErr := b.GenerateRSS()
	if rssErr != nil {
		errors = append(errors, rssErr)
	}

	pageErrors := b.GenerateSitePages()
	if len(pageErrors) > 0 {
		errors = append(errors, pageErrors...)
	}

	cpErr := b.CopyStatic()
	if cpErr != nil {
		errors = append(errors, cpErr)
	}

	b.WriteInfoJSON()

	return errors
}

func (b *Blog) GeneratePostsWithTemplate(tmpl string) []error {
	errors := []error{}
	extDir, _ := osext.ExecutableFolder()
	tmpDir := path.Join(extDir, "../src/github.com/ejamesc/goblawg", "templates")
	t := template.Must(template.New(tmpl).Funcs(funcMap).ParseFiles(path.Join(tmpDir, tmpl)))

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
		dirErr := createDirIfNotExist(filepath)
		if dirErr != nil {
			errors = append(errors, dirErr)
			continue
		}

		// Generate the HTML and write to file
		if b.LastModified.Before(post.LastModified) || b.LastModified.Equal(post.LastModified) {
			filepath = path.Join(filepath, "index.html")
			file, err := os.OpenFile(filepath, os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0755)
			defer file.Close()
			if err != nil {
				errors = append(errors, err)
				continue
			}

			bp := struct {
				*Post
				*Blog
			}{post, b}
			err = t.ExecuteTemplate(file, tmpl, bp)
			if err != nil {
				errors = append(errors, err)
				continue
			}
		}
	}

	b.LastModified = time.Now()
	return errors
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

	err = ioutil.WriteFile(path.Join(b.OutDir, "feed.rss"), []byte(rss), 0755)
	if err != nil {
		return err
	}

	b.LastModified = time.Now()
	return nil
}

// Generate the rest of the templates that isn't the blog
// TODO: Test the fuck out of this.
func (b *Blog) GenerateSitePages() []error {
	errors := []error{}
	fil, err := ioutil.ReadDir(b.InDir)
	if err != nil {
		errors = append(errors, err)
		return errors
	}

	for _, fi := range fil {
		if path.Ext(fi.Name()) == ".html" {
			isIndex := false
			if fi.Name() == "index.html" {
				isIndex = true
			}

			t := template.New(fi.Name()).Funcs(funcMap)
			t, err = t.ParseFiles(path.Join(b.InDir, fi.Name()))
			if err != nil {
				errors = append(errors, err)
				continue
			}

			name := strings.Split(fi.Name(), ".")
			if len(name) != 2 {
				errors = append(errors, fmt.Errorf("%s is a bad filename, expected x.html", fi.Name()))
				continue
			}

			folder := ""
			if !isIndex {
				folder = name[0]
				err = createDirIfNotExist(path.Join(b.OutDir, folder))
				if err != nil {
					errors = append(errors, err)
					continue
				}
			}

			oPath := path.Join(b.OutDir, folder, "index.html")
			f, err := os.OpenFile(oPath, os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0755)
			if err != nil {
				errors = append(errors, err)
				continue
			}

			pb := struct {
				*Blog
				Posts []*Post
			}{b, b.GetAllPosts()}
			err = t.ExecuteTemplate(f, fi.Name(), pb)
			if err != nil {
				errors = append(errors, err)
				continue
			}
		}

	}

	b.LastModified = time.Now()
	return errors
}

func (b *Blog) CopyStatic() error {
	extDir, _ := osext.ExecutableFolder()
	staticPath := path.Join(extDir, "../src/github.com/ejamesc/goblawg", "static")

	// CopyTree demands that the destination folder not exist
	// If it does, we delete it
	outDir := path.Join(b.OutDir, "static")
	_, err := os.Stat(outDir)
	if err == nil {
		err = os.RemoveAll(outDir)
		if err != nil {
			return err
		}
	} else if err != nil && !os.IsNotExist(err) {
		return err
	}

	// CopyTree options:
	// Symlinks - if true, symbolic links copied, if false symlinked files copied
	// IgnoreDanglingSymlinks - supress error thrown when symlink links to missing file
	// Optional CopyFunction
	// Optional Ignore function
	options := &shutil.CopyTreeOptions{
		Symlinks:               false,
		IgnoreDanglingSymlinks: true,
		CopyFunction:           shutil.Copy,
		Ignore:                 nil,
	}
	err = shutil.CopyTree(staticPath, outDir, options)
	if err != nil {
		return err
	}
	return nil
}

// Helpers

func createDirIfNotExist(dirpath string) error {
	_, err := os.Stat(dirpath)
	// The directory doesn't yet exist
	if err != nil && os.IsNotExist(err) {
		dirErr := os.Mkdir(dirpath, 0755)
		if dirErr != nil {
			return dirErr
		}
	}
	return nil
}
