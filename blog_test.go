package goblawg_test

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"

	"github.com/ejamesc/goblawg"
)

var tmpdir = os.TempDir()
var settingsJSON = fmt.Sprintf(
	`{"name": "My First Blog", 
	"out_dir": "%s", 
	"in_dir": "%s", 
	"static_dir": "%s", 
	"link": "http://elijames.org",
	"description": "Test Blog",
	"author": "Eli James",
	"email": "bob@test.com"}`, tmpdir, tmpdir, tmpdir)
var infoJSON = `{"name":"My First Blog",
"link":"http://elijames.org",
"description":"Test Blog",
"author":"Eli James",
"email":"bob@test.com",
"last_modified":"0001-01-01T00:00:00Z"}`

// Test NewBlog constructs and returns a Blog struct correctly
func TestNewBlog(t *testing.T) {
	// Setup
	settingsPath := path.Join(tmpdir, "settings.json")
	infoPath := path.Join(tmpdir, "info.json")
	ioutil.WriteFile(settingsPath, []byte(settingsJSON), 0775)
	ioutil.WriteFile(infoPath, []byte(infoJSON), 0775)

	b, err := goblawg.NewBlog(settingsPath)

	// Teardown
	defer os.Remove(settingsPath)
	defer os.Remove(infoPath)

	ok(t, err)

	equals(t, "My First Blog", b.Name)
	equals(t, "Test Blog", b.Description)
	equals(t, tmpdir, b.OutDir)
	equals(t, tmpdir, b.InDir)
	equals(t, []*goblawg.Post{}, b.Posts)
	equals(t, time.Time{}, b.LastModified)

	_, err = os.Stat(path.Join(tmpdir, "info.json"))
	ok(t, err)

	//blah, _ := ioutil.ReadFile(path.Join(tmpdir, "info.json"))
	//fmt.Println(string(blah))
}

var infoContent = fmt.Sprintf(`{"title": "Test Post", 
	"body": "%s", 
	"link": "test-post", 
	"time": "%s", 
	"is_draft": false, 
	"last_modified": "%s"}`,
	base64.StdEncoding.EncodeToString([]byte("The quick brown fox jumps over the lazy dog")),
	time.Now().Format(time.RFC3339),
	time.Now().Format(time.RFC3339))

func TestLoadPostFromFile(t *testing.T) {
	postPath := path.Join(tmpdir, "post.json")
	ioutil.WriteFile(postPath, []byte(infoContent), 0775)

	fi, _ := os.Stat(postPath)
	p, err := goblawg.NewPostFromFile(postPath, fi)

	ok(t, err)
	equals(t, "Test Post", p.Title)
	equals(t, []byte("The quick brown fox jumps over the lazy dog"), p.Body)
	equals(t, "test-post", p.Link)
	equals(t, false, p.IsDraft)
}
