package migrations

import (
	"context"

	migrasyon "github.com/tosunlab/migrasyon/sdk/go"
)

func init() { migrasyon.Register(RenameUsers{}) }

// RenameUsers is the fixture for rename_table and rename_column, each its own inverse.
type RenameUsers struct{}

// Up renames the table and one of its columns.
func (m RenameUsers) Up(ctx context.Context, s *migrasyon.Schema) error {
	return s.Table("users").
		RenameTo("people").
		RenameColumn("email", "address").
		Apply(ctx)
}

// Down renames both back, in the order that undoes them.
func (m RenameUsers) Down(ctx context.Context, s *migrasyon.Schema) error {
	return s.Table("people").
		RenameColumn("address", "email").
		RenameTo("users").
		Apply(ctx)
}
