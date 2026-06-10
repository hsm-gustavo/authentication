package ui

import (
	"embed"
	"io/fs"
)

// prefixo all garante q retorna até arquivos ocultos
//go:embed all:dist
var FS embed.FS

func GetFS() fs.FS {
	f, err := fs.Sub(FS, "dist")
	if err != nil {
		panic(err)
	}
	return f
}