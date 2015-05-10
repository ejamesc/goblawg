package goblawg

import (
	"time"

	"github.com/russross/blackfriday"
)

/* Template functions
* */
func Markdown(input []byte) string {
	output := blackfriday.MarkdownCommon(input)
	return string(output)
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
