package migrations

import migrasyon "github.com/tosunlab/migrasyon/sdk/go"

// init registers the backfill of widening accounts.email, whose down has nothing of its own to undo.
func init() {
	migrasyon.Register(migrasyon.ReplaceColumn("accounts", "email", migrasyon.String(255)).With(migrasyon.String(320)).Backfill())
}
