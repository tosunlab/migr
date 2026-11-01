package migrations

import (
	"context"

	migrasyon "github.com/tosunlab/migrasyon/sdk/go"
)

func init() { migrasyon.Register(AddNickname{}) }

// AddNickname adds an optional nickname and an index over it, and authors its own down.
type AddNickname struct{}

// Up adds the column and the index.
func (AddNickname) Up(ctx context.Context, s *migrasyon.Schema) error {
	return s.Table("accounts").
		AddColumn("nickname", migrasyon.String(64)).Null().
		AddIndex("idx_accounts_nickname", "nickname").
		Apply(ctx)
}

// Down drops both, recording what each was so the drop can itself be reversed.
func (AddNickname) Down(ctx context.Context, s *migrasyon.Schema) error {
	return s.Table("accounts").
		DropIndex("idx_accounts_nickname").Was(migrasyon.Columns("nickname")).
		DropColumn("nickname").Was(migrasyon.Column{Name: "nickname", Type: migrasyon.String(64), Nullable: true}).
		Apply(ctx)
}
