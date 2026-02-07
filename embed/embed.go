package embed

import "embed"

// Assets contains all embedded static assets (CSS, JS).
// Files are populated by `make deps`.
//
//go:embed *.js *.css
var Assets embed.FS
