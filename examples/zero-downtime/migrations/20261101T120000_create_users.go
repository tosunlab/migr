package migrations

import (
	"context"

	migrasyon "github.com/tosunlab/migrasyon/sdk/go"
)

func init() { migrasyon.Register(CreateUsers{}) }

// CreateUsers adds the users table the app is serving from before the change begins.
type CreateUsers struct{}

// Up creates the table with its key inline.
func (CreateUsers) Up(ctx context.Context, s *migrasyon.Schema) error {
	return s.CreateTable("users").
		Column("id", migrasyon.Int64()).
		Column("name", migrasyon.String(100)).
		Column("email", migrasyon.String(255)).
		PrimaryKey("id").
		Apply(ctx)
}
