package goblawg

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
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
		LastGen string `json:"last_gen"`
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
		tts, err := time.Parse(layout, c.LastGen)
		if err != nil {
			fmt.Printf("Error with parsing lastgen timestamp: %v\n", err)
			lastModifiedParseIsSuccess = false
		}
		b.LastModified = tts
	}

	if !lastModifiedParseIsSuccess {
		b.LastModified = time.Time{}
	}

	return b, nil
}

// Creates an info.json to be placed in the public dir
func (b *Blog) CreateInfoJSON() {
	lmt := b.LastModified.Format(layout)
	type omit *struct{}
	toWrite, _ := json.Marshal(struct {
		*Blog
		InDir        omit   `json:"in_dir,omitempty"`
		OutDir       omit   `json:"out_dir,omitempty"`
		StaticDir    omit   `json:"static_dir,omitempty"`
		Posts        omit   `json:"posts,omitempty"`
		LastModified string `json:"last_modified"`
	}{Blog: b, LastModified: lmt})
	_ = ioutil.WriteFile(path.Join(b.OutDir, "info.json"), toWrite, 0775)
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

	var markdownFileList []os.FileInfo
	for _, entry := range listFileInfo {
		if isJSONFile(entry.Name()) {
			markdownFileList = append(markdownFileList, entry)
		}
	}

	posts := make([]*Post, len(markdownFileList))

	for i, entry := range markdownFileList {
		fpath := path.Join(dir, entry.Name())

		p, err := newPostFromFile(fpath, entry)
		if err != nil {
			fmt.Println(err)
			continue
		}
		posts[i] = p
	}

	return posts, nil
}

func newPostFromFile(path string, fi os.FileInfo) (*Post, error) {
	if !isJSONFile(path) {
		return nil, fmt.Errorf("%s does not have a markdown or text file extension", path)
	}

	p := &Post{}
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
