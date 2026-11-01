package migrations

import migrasyon "github.com/tosunlab/migrasyon/sdk/go"

// init registers the expand of renaming accounts.handle to login, which carries a default from the start.
func init() {
	migrasyon.Register(migrasyon.ReplaceColumn("accounts", "handle", migrasyon.String(64)).WasNull().
		WasDefault(migrasyon.Literal("anon")).RenameTo("login").Default(migrasyon.Literal("guest")).
		Through(migrasyon.TransformLower).Expand())
}
