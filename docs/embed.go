package docs

import "embed"

// Assets contains embedded documentation assets (images, etc.)
// referenced by the project README.
//
//go:embed *.gif *.png
var Assets embed.FS
