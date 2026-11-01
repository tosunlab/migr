package migrasyontest

import (
	"context"
	"errors"
	"strconv"
	"testing"

	migrasyon "github.com/tosunlab/migrasyon/sdk/go"
)

// errFixture is what a fixture returns when a test told it to fail.
var errFixture = errors.New("the migration would not run")

// theTable is what every fixture here creates or changes.
const theTable = "users"

// creates is a migration that creates a table and authors the down that drops it.
type creates struct{}

// Up creates the table.
func (m creates) Up(ctx context.Context, s *migrasyon.Schema) error {
	return s.CreateTable(theTable).
		Column("id", migrasyon.Int64()).
		Column("email", migrasyon.String(255)).
		PrimaryKey("id").
		Apply(ctx)
}

// Down drops it, recording what it held so the drop inverts exactly.
func (m creates) Down(ctx context.Context, s *migrasyon.Schema) error {
	return s.Table(theTable).Drop().Was(migrasyon.TableDef{
		Columns: []migrasyon.Column{
			{Name: "id", Type: migrasyon.Int64()},
			{Name: "email", Type: migrasyon.String(255)},
		},
		PrimaryKey: []string{"id"},
	}).Apply(ctx)
}

// derives is a migration that authors no down at all, so the engine inverts its up.
type derives struct{}

// Up adds one column.
func (m derives) Up(ctx context.Context, s *migrasyon.Schema) error {
	return s.Table(theTable).AddColumn("tier", migrasyon.Int32()).Null().Apply(ctx)
}

// oneWay is a migration whose author states it cannot be undone.
type oneWay struct{}

// Up drops a column, recording nothing.
func (m oneWay) Up(ctx context.Context, s *migrasyon.Schema) error {
	return s.Table(theTable).DropColumn("legacy").Apply(ctx)
}

// Irreversible returns the reason a rollback halts with.
func (m oneWay) Irreversible() string {
	return "the dropped values are not recoverable; restore from a backup"
}

// branches is a migration that reads the database and emits a different column for each answer.
type branches struct{}

// theRead is what branches asks, which a stage has to answer for it to run at all.
const theRead = "SELECT count(*) FROM users"

// Up asks how many rows are there, and adds a nullable column when any are.
func (m branches) Up(ctx context.Context, s *migrasyon.Schema) error {
	rows, err := s.Query(ctx, theRead)
	if err != nil {
		return err
	}

	change := s.Table(theTable).AddColumn("tier", migrasyon.Int32())

	if rows.Len() > 0 && rows.Values[0][0] != nil && *rows.Values[0][0] != "0" {
		change = change.Null()
	}

	return change.Apply(ctx)
}

// mixes is a migration that sends a statement of its own beside the operation it declares.
type mixes struct{}

// theStatement is what mixes sends itself, which plan mode holds and apply mode reports performed.
const theStatement = "UPDATE users SET tier = 1"

// Up sends its own statement and then emits.
func (m mixes) Up(ctx context.Context, s *migrasyon.Schema) error {
	outcome, err := s.Exec(ctx, theStatement)
	if err != nil {
		return err
	}

	s.Log(migrasyon.LevelInfo, "backfilled", map[string]string{
		"rows": outcome.RowsAffected, "executed": strconv.FormatBool(outcome.Executed),
	})

	return s.Table(theTable).AddIndex("idx_users_tier", "tier").Apply(ctx)
}

// Down drops the index, so this migration has an authored undo despite sending a statement.
func (m mixes) Down(ctx context.Context, s *migrasyon.Schema) error {
	return s.Table(theTable).DropIndex("idx_users_tier").Was(migrasyon.Columns("tier")).Apply(ctx)
}

// derivesOverStatement sends a statement of its own and authors no down, which has no inverse.
type derivesOverStatement struct{}

// Up sends its own statement and emits, with nothing authored to undo either.
func (m derivesOverStatement) Up(ctx context.Context, s *migrasyon.Schema) error {
	if _, err := s.Exec(ctx, theStatement); err != nil {
		return err
	}

	return s.Table(theTable).AddIndex("idx_users_tier", "tier").Apply(ctx)
}

// asks is a migration that branches on what the engine is, and on what it makes of an operation.
type asks struct{}

// Up asks what this run is against and what the engine makes of the operation before emitting it.
func (m asks) Up(ctx context.Context, s *migrasyon.Schema) error {
	facts, err := s.Info(ctx)
	if err != nil {
		return err
	}

	op := migrasyon.AddIndex{Table: theTable, Name: "idx_users_email", Columns: []string{"email"}}

	verdict, err := s.Capability(ctx, op)
	if err != nil {
		return err
	}

	s.Log(migrasyon.LevelDebug, "asked", map[string]string{
		"driver": facts.Driver, "support": verdict.Support, "mode": string(facts.Mode),
	})

	return s.Apply(ctx, op)
}

// forgetful builds a change and returns without applying it, which emits nothing.
type forgetful struct{}

// Up builds a chain and never applies it.
func (m forgetful) Up(_ context.Context, s *migrasyon.Schema) error {
	s.Table(theTable).AddColumn("tier", migrasyon.Int32())

	return nil
}

// failing is a migration whose own code returns an error.
type failing struct{}

// Up fails.
func (m failing) Up(_ context.Context, _ *migrasyon.Schema) error {
	return errFixture
}

// swallows reads the database, ignores the refusal, and emits anyway.
type swallows struct{}

// Up reads, drops the error on the floor, and emits what it was going to emit regardless.
func (m swallows) Up(ctx context.Context, s *migrasyon.Schema) error {
	_, _ = s.Query(ctx, theRead)

	return s.Table(theTable).AddColumn("tier", migrasyon.Int32()).Null().Apply(ctx)
}

// theIndex is the operation several cases here expect to see emitted.
func theIndex() migrasyon.Operation {
	return migrasyon.AddIndex{Table: theTable, Name: "idx_users_tier", Columns: []string{"tier"}}
}

// theColumn is the nullable column the branching and swallowing fixtures emit.
func theColumn(t *testing.T, nullable bool) migrasyon.Operation {
	t.Helper()

	return migrasyon.AddColumn{
		Table:  theTable,
		Column: migrasyon.Column{Name: "tier", Type: migrasyon.Int32(), Nullable: nullable},
	}
}
