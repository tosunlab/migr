package migrations

import migrasyon "github.com/tosunlab/migrasyon/sdk/go"

// init registers the swap of widening accounts.email: both columns change nullability, then names.
func init() {
	migrasyon.Register(migrasyon.ReplaceColumn("accounts", "email", migrasyon.String(255)).With(migrasyon.String(320)).Swap())
}
