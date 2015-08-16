package goblawg

import (
	"html/template"
	"math/rand"
	"time"

	"github.com/russross/blackfriday"
)

/* Template functions
* */
func Markdown(input []byte) template.HTML {
	output := blackfriday.MarkdownCommon(input)
	return template.HTML(string(output))
}

func DateFmt(tt time.Time) string {
	const layout = "3:04pm, 2 January 2006"
	return tt.Format(layout)
}

func IsEven(i int) bool {
	if i%2 == 0 {
		return true
	} else {
		return false
	}
}

func Subslice(a []*Post, start, end int) []*Post {
	ln := end - start
	if len(a) < ln || ln < 0 {
		return a
	} else {
		return a[start:end]
	}
}

func RandomPosts(a []*Post, currPostURL string, numPosts int) []*Post {
	perm := rand.Perm(len(a))
	res := []*Post{}
	for i := range perm {
		v := perm[i]
		if a[v].Link != currPostURL {
			res = append(res, a[v])
		}
		if len(res) >= numPosts {
			break
		}
	}

	return res
}
