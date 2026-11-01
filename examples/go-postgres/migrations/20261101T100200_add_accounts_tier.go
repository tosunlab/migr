package migrations

import (
	"context"

	migrasyon "github.com/tosunlab/migrasyon/sdk/go"
)

func init() { migrasyon.Register(AddTier{}) }

// AddTier adds a tier, defaulting it to free only where accounts already exist.
type AddTier struct{}

// Up reads the table and decides what to emit from the answer.
func (AddTier) Up(ctx context.Context, s *migrasyon.Schema) error {
	rows, err := s.Query(ctx, "SELECT count(*) FROM accounts")
	if err != nil {
		return err
	}

	tier := s.Table("accounts").AddColumn("tier", migrasyon.String(16))
	if *rows.Values[0][0] != "0" {
		tier = tier.Default(migrasyon.Literal("free"))
	}

	return tier.Apply(ctx)
}

// Irreversible says why no down is offered.
func (AddTier) Irreversible() string {
	return "the tiers written after it ran are not recorded anywhere a rollback could restore them from"
}
