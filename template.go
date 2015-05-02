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
