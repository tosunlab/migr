package migrations

import (
	"context"

	migrasyon "github.com/tosunlab/migrasyon/sdk/go"
)

func init() { migrasyon.Register(AddProfileColumns{}) }

// AddProfileColumns is the fixture for add_column, drop_column, and alter_column.
type AddProfileColumns struct{}

// Dependencies declares the migration that creates the table this one changes.
func (m AddProfileColumns) Dependencies() []string {
	return []string{"20260727T120001"}
}

// Up adds two columns and widens a third, which is the one change that may lose data.
func (m AddProfileColumns) Up(ctx context.Context, s *migrasyon.Schema) error {
	return s.Table("users").
		AddColumn("bio", migrasyon.Text()).Null().
		AddColumn("tier", migrasyon.String(16)).Default(migrasyon.Literal("free")).
		AlterColumn("email").Type(migrasyon.String(320)).WasType(migrasyon.String(255)).
		Apply(ctx)
}

// Down narrows the column back and drops the two, recording what each one was.
func (m AddProfileColumns) Down(ctx context.Context, s *migrasyon.Schema) error {
	return s.Table("users").
		AlterColumn("email").Type(migrasyon.String(255)).WasType(migrasyon.String(320)).
		DropColumn("tier").Was(migrasyon.Column{
		Name: "tier", Type: migrasyon.String(16), Default: migrasyon.Literal("free"),
	}).
		DropColumn("bio").Was(migrasyon.Column{Name: "bio", Type: migrasyon.Text(), Nullable: true}).
		Apply(ctx)
}
