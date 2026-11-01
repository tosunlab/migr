package migrations

import migrasyon "github.com/tosunlab/migrasyon/sdk/go"

// init registers the expand of widening accounts.email: the column it is rebuilt in, added nullable.
func init() {
	migrasyon.Register(migrasyon.ReplaceColumn("accounts", "email", migrasyon.String(255)).With(migrasyon.String(320)).Expand())
}
