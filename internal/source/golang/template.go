package golang

// Ext is the extension a Go migration is stored under, and the one migrasyon create writes.
const Ext = ".go"

// Exts are every extension the Go runner loads, which is the one the go tool compiles.
var Exts = []string{Ext}

// SDK is the module a Go migration imports, which a new project adds with go get.
const SDK = "github.com/tosunlab/migrasyon/sdk/go"

// Template is the file migrasyon create writes, which compiles once the SDK is in go.mod.
const Template = `package migrations

import (
	"context"

	migrasyon "` + SDK + `"
)

func init() { migrasyon.Register(Migration{}) }

// Migration is what this file does. Rename it to something the next reader will recognize.
type Migration struct{}

// Up performs the change. Every operation it emits is recorded, compiled for the target
// engine, and run in one transaction:
//
//	return s.CreateTable("users").
//		Column("id", migrasyon.UUID()).
//		Column("email", migrasyon.String(255)).
//		Column("created_at", migrasyon.TimestampTZ()).Default(migrasyon.Call(migrasyon.FuncCurrentTimestamp)).
//		PrimaryKey("id").
//		Apply(ctx)
//
// s.Table(name) does the same for a table that already exists, and s.View(name) for a
// view. A column is NOT NULL unless the chain says Null.
//
// Nothing is emitted until Apply, and a chain built without it emits nothing at all --
// so migrasyon fails such a migration rather than reporting a change it never made.
func (m Migration) Up(ctx context.Context, s *migrasyon.Schema) error {
	return nil
}

// Down undoes it. Delete this method and migrasyon derives the inverse from what Up emitted.
//
// One thing has no inverse to derive: a statement this migration sends itself with
// s.Exec. That is committed on its own, before the transaction the operations run in,
// and a later failure does not undo it -- so a migration that calls s.Exec must write
// this method, or declare itself irreversible:
//
//	func (m Migration) Irreversible() string { return "the reason a rollback halts with" }
//
// A migration declares one or the other. Both is a registration error.
func (m Migration) Down(ctx context.Context, s *migrasyon.Schema) error {
	return nil
}
`
