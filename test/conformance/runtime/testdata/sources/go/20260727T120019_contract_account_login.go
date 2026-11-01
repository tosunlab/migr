package migrations

import migrasyon "github.com/tosunlab/migrasyon/sdk/go"

// init registers the contract of renaming accounts.handle, irreversible after a lowercasing backfill.
func init() {
	migrasyon.Register(migrasyon.ReplaceColumn("accounts", "handle", migrasyon.String(64)).WasNull().
		WasDefault(migrasyon.Literal("anon")).RenameTo("login").Default(migrasyon.Literal("guest")).
		Through(migrasyon.TransformLower).Contract())
}
