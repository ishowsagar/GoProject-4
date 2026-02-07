package migrations

import (
	"embed"
)

//go:embed *.sql
var FS embed.FS // created variable with type being this embed.fs
