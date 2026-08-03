// Package migrations provide embedded SQL schema migration files
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
