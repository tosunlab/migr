package migrations

import (
	"context"

	migrasyon "github.com/tosunlab/migrasyon/sdk/go"
)

func init() { migrasyon.Register(CreateBooks{}) }

// CreateBooks adds the books table, each book naming the author the SQL migration before it created.
type CreateBooks struct{}

// Up creates the table with its key inline, then the key to authors.
func (CreateBooks) Up(ctx context.Context, s *migrasyon.Schema) error {
	if err := s.CreateTable("books").
		Column("id", migrasyon.Int64()).
		Column("author_id", migrasyon.Int64()).
		Column("title", migrasyon.String(200)).
		PrimaryKey("id").
		Apply(ctx); err != nil {
		return err
	}

	return s.Table("books").
		AddForeignKey("fk_books_author", []string{"author_id"}, "authors", []string{"id"}).
		Apply(ctx)
}
