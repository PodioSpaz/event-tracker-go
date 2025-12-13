package migrations

import (
	"embed"
	"io/fs"
)

// SQL contains all embedded SQL migration files
//
//go:embed *.sql
var SQL embed.FS

// FS returns a filesystem interface for the migrations
func FS() fs.FS {
	return SQL
}
