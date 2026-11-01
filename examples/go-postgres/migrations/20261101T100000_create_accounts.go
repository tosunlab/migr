package migrations

import (
	"context"

	migrasyon "github.com/tosunlab/migrasyon/sdk/go"
)

func init() { migrasyon.Register(CreateAccounts{}) }

// CreateAccounts adds the accounts table, whose down migrasyon derives by inverting the up.
type CreateAccounts struct{}

// Up creates the table with its key inline.
func (CreateAccounts) Up(ctx context.Context, s *migrasyon.Schema) error {
	return s.CreateTable("accounts").
		Column("id", migrasyon.Int64()).
		Column("email", migrasyon.String(255)).
		Column("created_at", migrasyon.TimestampTZ()).Default(migrasyon.Call(migrasyon.FuncCurrentTimestamp)).
		PrimaryKey("id").
		Apply(ctx)
}
