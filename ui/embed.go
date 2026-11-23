// Package ui embeds the HTML templates and static assets.
package ui

import "embed"

// FS contains templates/ and static/.
//
//go:embed templates static
var FS embed.FS
