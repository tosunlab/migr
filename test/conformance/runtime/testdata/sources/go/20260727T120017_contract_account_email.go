package migrations

import migrasyon "github.com/tosunlab/migrasyon/sdk/go"

// init registers the contract of widening accounts.email, whose down copies the old values back.
func init() {
	migrasyon.Register(migrasyon.ReplaceColumn("accounts", "email", migrasyon.String(255)).With(migrasyon.String(320)).Contract())
}
