package migrations

import (
	"context"

	migrasyon "github.com/tosunlab/migrasyon/sdk/go"
)

func init() { migrasyon.Register(ConstrainOrders{}) }

// ConstrainOrders is the fixture for add_check, drop_check, add_foreign_key, and drop_foreign_key.
type ConstrainOrders struct{}

// Up constrains a column and references another table, which is a rule rather than a row.
func (m ConstrainOrders) Up(ctx context.Context, s *migrasyon.Schema) error {
	return s.Table("orders").
		AddCheck("ck_orders_total", "total >= 0").
		AddForeignKey("fk_orders_person", []string{"person_id"}, "people", []string{"id"}).
		OnDelete(migrasyon.ActionCascade).OnUpdate(migrasyon.ActionRestrict).
		Apply(ctx)
}

// Down removes both, recording what each one said.
func (m ConstrainOrders) Down(ctx context.Context, s *migrasyon.Schema) error {
	return s.Table("orders").
		DropForeignKey("fk_orders_person").Was(migrasyon.ForeignKeyDef{
		Columns:         []string{"person_id"},
		ReferencesTable: "people", ReferencesColumns: []string{"id"},
		OnDelete: migrasyon.ActionCascade, OnUpdate: migrasyon.ActionRestrict,
	}).
		DropCheck("ck_orders_total").Was(migrasyon.CheckDef{Expression: "total >= 0"}).
		Apply(ctx)
}
