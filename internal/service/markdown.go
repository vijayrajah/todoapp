package service

import (
	"bytes"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

// Raw HTML in the source is escaped (goldmark's default), so the output is safe to inject.
var markdown = goldmark.New(goldmark.WithExtensions(extension.GFM))

// RenderMarkdown converts markdown source to safe HTML.
func RenderMarkdown(src string) string {
	if src == "" {
		return ""
	}
	var buf bytes.Buffer
	if err := markdown.Convert([]byte(src), &buf); err != nil {
		return ""
	}
	return buf.String()
}
