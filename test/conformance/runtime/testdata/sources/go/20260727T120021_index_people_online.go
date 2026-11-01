package migrations

import (
	"context"

	migrasyon "github.com/tosunlab/migrasyon/sdk/go"
)

func init() { migrasyon.Register(IndexPeopleOnline{}) }

// IndexPeopleOnline is the fixture for the online declaration, which every SDK reports identically.
type IndexPeopleOnline struct{}

// Online declares the index is built through a shadow table rather than by a plain ALTER.
func (m IndexPeopleOnline) Online() migrasyon.Strategy { return migrasyon.ShadowTable }

// Up indexes one column.
func (m IndexPeopleOnline) Up(ctx context.Context, s *migrasyon.Schema) error {
	return s.Table("people").AddIndex("idx_people_region", "region").Apply(ctx)
}

// Down removes it, recording which column it covered.
func (m IndexPeopleOnline) Down(ctx context.Context, s *migrasyon.Schema) error {
	return s.Table("people").DropIndex("idx_people_region").Was(migrasyon.Columns("region")).Apply(ctx)
}
