package web

import "embed"

// Files contains the local Klip web interface.
//
//go:embed index.html assets
var Files embed.FS
