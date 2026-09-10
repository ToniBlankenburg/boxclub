// Package templates bündelt die html/template-Fragmente, die die Handler in app/
// an das htmx-Frontend ausliefern. Die Dateien liegen als eigenes Package vor,
// damit go:embed sie erreicht und in das Binary aufnimmt.
package templates

import "embed"

// FS enthält alle Fragment-Templates.
//
//go:embed *.html
var FS embed.FS
