package goblawg

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"time"
)

const layout = "2-Jan-2006-15-04-05"

type Blog struct {
	Name         string    `json:"name"`
	Link         string    `json:"link"`
	Description  string    `json:"description"`
	Author       string    `json:"author"`
	Email        string    `json:"email"`
	Posts        []*Post   `json:"posts"`
	InDir        string    `json:"in_dir"`
	OutDir       string    `json:"out_dir"`
	StaticDir    string    `json:"static_dir"`
	LastModified time.Time `json:"last_modified"`
}

type Post struct {
	Title        string    `json:"title"`
	Body         []byte    `json:"body"`
	Link         string    `json:"link"`
	Time         time.Time `json:"time"`
	IsDraft      bool      `json:"is_draft"`
	LastModified time.Time `json:"last_modified"`
}

// Generator
func NewBlog(settingsFile string) (*Blog, error) {
	settingsJSON, err := ioutil.ReadFile(settingsFile)
	if err != nil {
		fmt.Printf("Dude, unable to read settings.json: %v\n", err)
		return nil, err
	}

	var b *Blog
	err = json.Unmarshal(settingsJSON, &b)
	if err != nil {
		fmt.Printf("Unmarshalling of settings file failed: %v\n", err)
		return nil, err
	}

	b.Posts, err = loadPostsFromDir(path.Join(b.InDir, "posts"))
	if err != nil {
		fmt.Printf("Couldn't load posts from %v: %v\n", b.InDir, err)
		b.Posts = []*Post{}
	}

	type timeDecode struct {
		LastGen time.Time `json:"last_gen"`
	}
	var c timeDecode
	lastModifiedParseIsSuccess := true

	publicJSON, err := ioutil.ReadFile(path.Join(b.OutDir, "info.json"))
	if err != nil {
		fmt.Printf("No info.json generated: %v\n", err)
		lastModifiedParseIsSuccess = false
		b.CreateInfoJSON()
	} else {
		err = json.Unmarshal(publicJSON, &c)
		if err != nil {
			fmt.Println("Unmashalling of info.json error: %v\n", err)
			lastModifiedParseIsSuccess = false
		}
		b.LastModified = c.LastGen
	}

	if !lastModifiedParseIsSuccess {
		b.LastModified = time.Time{}
	}

	return b, nil
}

type ByTime []*Post

func (t ByTime) Len() int           { return len(t) }
func (t ByTime) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t ByTime) Less(i, j int) bool { return t[i].Time.Before(t[j].Time) }

// Return all published posts, sorted in reverse chronological order
func (b *Blog) GetPublishedPosts() []*Post {
	return b.getPostsWithDraft(false)
}

func (b *Blog) GetAllPosts() []*Post {
	return b.getPostsWithDraft(true)
}

// Return all posts, sorted in reverse chronological order
func (b *Blog) getPostsWithDraft(drafts bool) []*Post {
	ps := []*Post{}
	if !drafts {
		for _, p := range b.Posts {
			if !p.IsDraft {
				ps = append(ps, p)
			}
		}
	}
	sort.Sort(sort.Reverse(ByTime(ps)))
	return ps
}

// Inner type, used to omit fields from JSON marshalling
type omit *struct{}

// Creates an info.json to be placed in the public dir
// This info.json is part of the blog's public API, and
// is where the LastModified timestamp is taken from
func (b *Blog) CreateInfoJSON() {
	toWrite, _ := json.Marshal(struct {
		*Blog
		InDir     omit `json:"in_dir,omitempty"`
		OutDir    omit `json:"out_dir,omitempty"`
		StaticDir omit `json:"static_dir,omitempty"`
		Posts     omit `json:"posts,omitempty"`
	}{Blog: b})
	_ = ioutil.WriteFile(path.Join(b.OutDir, "info.json"), toWrite, 0775)
}

func (b *Blog) GetPostByLink(link string) *Post {
	for _, p := range b.Posts {
		if p.Link == link {
			return p
		}
	}
	return nil
}

/* Helpers */
func loadPostsFromDir(dir string) ([]*Post, error) {
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("%v does not exist, yo!", dir)
	}

	listFileInfo, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var jsonFileList []os.FileInfo
	for _, entry := range listFileInfo {
		if isJSONFile(entry.Name()) {
			jsonFileList = append(jsonFileList, entry)
		}
	}

	posts := []*Post{}

	for _, entry := range jsonFileList {
		fpath := path.Join(dir, entry.Name())

		p, err := NewPostFromFile(fpath, entry)
		if err != nil {
			fmt.Printf("Unable to create post: %v\n", err)
			continue
		}
		posts = append(posts, p)
	}

	return posts, nil
}

func NewPostFromFile(path string, fi os.FileInfo) (*Post, error) {
	if !isJSONFile(path) {
		return nil, fmt.Errorf("%s does not have a JSON file extension", path)
	}

	fc, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var p *Post
	err = json.Unmarshal(fc, &p)

	if err != nil {
		fmt.Printf("Unmarshalling error: %v\n", err)
		return nil, err
	}
	p.LastModified = fi.ModTime()

	return p, nil
}

func isJSONFile(n string) bool {
	ext := path.Ext(n)
	if ext == ".json" {
		return true
	} else {
		return false
	}
}
