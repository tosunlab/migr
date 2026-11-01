// Package migrations is the conformance fixture set expressed in Go.
package migrations

import (
	"context"

	migrasyon "github.com/tosunlab/migrasyon/sdk/go"
)

func init() { migrasyon.Register(CreateUsers{}) }

// CreateUsers is the fixture for create_table and drop_table.
type CreateUsers struct{}

// Up creates the table the rest of the fixture set builds on.
func (m CreateUsers) Up(ctx context.Context, s *migrasyon.Schema) error {
	return s.CreateTable("users").
		Column("id", migrasyon.UUID()).
		Column("email", migrasyon.String(255)).
		Column("created_at", migrasyon.TimestampTZ()).Null().
		Default(migrasyon.Call(migrasyon.FuncCurrentTimestamp)).
		PrimaryKey("id").
		Apply(ctx)
}

// Down drops it, recording what it held so that the drop inverts exactly.
func (m CreateUsers) Down(ctx context.Context, s *migrasyon.Schema) error {
	return s.Table("users").Drop().Was(migrasyon.TableDef{
		Columns: []migrasyon.Column{
			{Name: "id", Type: migrasyon.UUID()},
			{Name: "email", Type: migrasyon.String(255)},
			{
				Name: "created_at", Type: migrasyon.TimestampTZ(),
				Nullable: true, Default: migrasyon.Call(migrasyon.FuncCurrentTimestamp),
			},
		},
		PrimaryKey: []string{"id"},
	}).Apply(ctx)
}
