package migrations

import (
	"context"

	migrasyon "github.com/tosunlab/migrasyon/sdk/go"
)

func init() { migrasyon.Register(ActivePeopleView{}) }

// ActivePeopleView is the fixture for create_view and drop_view, whose queries are engine SQL.
type ActivePeopleView struct{}

// Up defines one view and replaces another, which is the case that must record what it overwrote.
func (m ActivePeopleView) Up(ctx context.Context, s *migrasyon.Schema) error {
	if err := s.View("active_people").
		Create("SELECT id, address FROM people WHERE tier = 'paid'").
		Apply(ctx); err != nil {
		return err
	}

	return s.View("recent_people").
		Replace("SELECT id FROM people ORDER BY created_at DESC").
		Was("SELECT id FROM people").
		Apply(ctx)
}

// Down drops both, recording the query each was defined by.
func (m ActivePeopleView) Down(ctx context.Context, s *migrasyon.Schema) error {
	if err := s.View("recent_people").
		Drop().Was("SELECT id FROM people ORDER BY created_at DESC").
		Apply(ctx); err != nil {
		return err
	}

	return s.View("active_people").
		Drop().Was("SELECT id, address FROM people WHERE tier = 'paid'").
		Apply(ctx)
}
