package migrations

import (
	"context"

	migrasyon "github.com/tosunlab/migrasyon/sdk/go"
)

func init() { migrasyon.Register(AddNote{}) }

// AddNote is the fixture that authors no undo at all, which the engine derives one for.
type AddNote struct{}

// Up adds one column and writes no Down, so the engine is asked to invert what it emitted.
func (m AddNote) Up(ctx context.Context, s *migrasyon.Schema) error {
	return s.Table("people").
		AddColumn("note", migrasyon.Text()).Null().
		Apply(ctx)
}
