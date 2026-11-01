package migrations

import (
	"context"

	migrasyon "github.com/tosunlab/migrasyon/sdk/go"
)

func init() { migrasyon.Register(IndexPeopleConcurrently{}) }

// IndexPeopleConcurrently is the fixture for raw's no_transaction, which irVersion 3 carries.
type IndexPeopleConcurrently struct{}

// Up builds an index outside any transaction, then runs a statement that may share one.
func (m IndexPeopleConcurrently) Up(ctx context.Context, s *migrasyon.Schema) error {
	return s.Raw("postgres", "CREATE INDEX CONCURRENTLY idx_people_nickname ON people (nickname)").
		NoTransaction().
		Raw("postgres", "ANALYZE people").
		Apply(ctx)
}

// Down drops the index outside any transaction, as it was built.
func (m IndexPeopleConcurrently) Down(ctx context.Context, s *migrasyon.Schema) error {
	return s.Raw("postgres", "DROP INDEX CONCURRENTLY idx_people_nickname").NoTransaction().Apply(ctx)
}
