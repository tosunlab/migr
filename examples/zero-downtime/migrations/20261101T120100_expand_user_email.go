package migrations

import migrasyon "github.com/tosunlab/migrasyon/sdk/go"

func init() {
	migrasyon.Register(migrasyon.ReplaceColumn("users", "email", migrasyon.String(255)).
		With(migrasyon.String(320)).RenameTo("email_address").Expand())
}
