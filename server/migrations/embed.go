package migrations

import "embed"

// FS содержит SQL-миграции, встроенные в бинарник.
//
//go:embed *.sql
var FS embed.FS
