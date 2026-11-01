package migrations

import (
	"context"

	migrasyon "github.com/tosunlab/migrasyon/sdk/go"
)

func init() { migrasyon.Register(ShapePeople{}) }

// ShapePeople is the fixture that writes rows rather than schema, which irVersion 2 added.
type ShapePeople struct{}

// Up fills, rewrites and copies, through one chain per table.
func (m ShapePeople) Up(ctx context.Context, s *migrasyon.Schema) error {
	return s.Table("people").
		Backfill("nickname", migrasyon.String(64), migrasyon.Literal("anonymous")).
		Transform("email_key", migrasyon.String(255), "email", migrasyon.TransformLower).
		CopyInto("people_archive",
			migrasyon.CopyColumn{From: "id", To: "id"}, migrasyon.CopyColumn{From: "email", To: "email"}).
		Apply(ctx)
}

// Irreversible returns the reason a rollback halts with, since no data operation has an inverse.
func (m ShapePeople) Irreversible() string {
	return "every operation here writes rows, and what they replaced is recorded nowhere"
}
