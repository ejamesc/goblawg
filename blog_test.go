package goblawg_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"

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

// Test NewBlog constructs and returns a Blog struct correctly
func TestNewBlog(t *testing.T) {
	// Setup
	settingsPath := path.Join(tmpdir, "settings.json")
	ioutil.WriteFile(settingsPath, []byte(settingsJSON), 0775)

	b, err := goblawg.NewBlog(settingsPath)

	// Teardown
	defer os.Remove(settingsPath)
	defer os.Remove(path.Join(tmpdir, "info.json"))

	ok(t, err)

	equals(t, b.Name, "My First Blog")
	equals(t, b.Description, "Test Blog")
	equals(t, b.OutDir, tmpdir)
	equals(t, b.InDir, tmpdir)
	equals(t, b.Posts, []*goblawg.Post{})

	_, err = os.Stat(path.Join(tmpdir, "info.json"))
	ok(t, err)

	blah, _ := ioutil.ReadFile(path.Join(tmpdir, "info.json"))
	fmt.Println(string(blah))
}
