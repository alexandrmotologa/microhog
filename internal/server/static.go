package server

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed all:dist
var embeddedDistFS embed.FS

// GetEmbeddedFS returns an http.FileSystem rooted at the embedded dist directory.
func GetEmbeddedFS() http.FileSystem {
	sub, err := fs.Sub(embeddedDistFS, "dist")
	if err != nil {
		return nil
	}
	return http.FS(sub)
}
