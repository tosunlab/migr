package migrationtest

import (
	"slices"
	"testing"

	migrasyon "github.com/tosunlab/migrasyon/sdk/go"
	"github.com/tosunlab/migrasyon/sdk/go/migrasyontest"

	"example.com/go-postgres/migrations"
)

func TestCreateAccountsHasItsDownDerived(t *testing.T) {
	if r := migrasyontest.Reversal(t, migrations.CreateAccounts{}); r.Undo != migrasyon.UndoDerived {
		t.Errorf("Undo = %s, want derived", r.Undo)
	}
}

func TestAddNicknameAddsAColumnAndItsIndex(t *testing.T) {
	got := migrasyontest.Up(t, migrations.AddNickname{}).Kinds()

	if !slices.Equal(got, []migrasyon.Kind{migrasyon.KindAddColumn, migrasyon.KindAddIndex}) {
		t.Errorf("Up emitted %v", got)
	}
}

func TestAddNicknameAuthorsItsDown(t *testing.T) {
	got := migrasyontest.Down(t, migrations.AddNickname{}).Kinds()

	if !slices.Equal(got, []migrasyon.Kind{migrasyon.KindDropIndex, migrasyon.KindDropColumn}) {
		t.Errorf("Down emitted %v", got)
	}
}

// withAccounts answers AddTier's one read as a table already holding three accounts.
func withAccounts() migrasyontest.Stage {
	return migrasyontest.Stage{Reads: map[string]migrasyon.Rows{
		"SELECT count(*) FROM accounts": migrasyontest.Answer([]string{"count"}, []string{"3"}),
	}}
}

// empty answers AddTier's one read as a table holding no accounts.
func empty() migrasyontest.Stage {
	return migrasyontest.Stage{Reads: map[string]migrasyon.Rows{
		"SELECT count(*) FROM accounts": migrasyontest.Answer([]string{"count"}, []string{"0"}),
	}}
}

func TestAddTierDefaultsWhereAccountsExist(t *testing.T) {
	withAccounts().Up(t, migrations.AddTier{}).Is(t, migrasyon.AddColumn{Table: "accounts", Column: migrasyon.Column{
		Name: "tier", Type: migrasyon.String(16), Default: migrasyon.Literal("free"),
	}})
}

func TestAddTierNeedsNoDefaultOnAnEmptyTable(t *testing.T) {
	empty().Up(t, migrations.AddTier{}).Is(t, migrasyon.AddColumn{Table: "accounts", Column: migrasyon.Column{
		Name: "tier", Type: migrasyon.String(16),
	}})
}

func TestAddTierCannotBeUndone(t *testing.T) {
	if r := withAccounts().Reversal(t, migrations.AddTier{}); r.Undo != migrasyon.UndoDeclared {
		t.Errorf("Undo = %s, want declared", r.Undo)
	}
}

func TestTheMigrationsApplyAndRefuseToRollBackPastTheTier(t *testing.T) {
	project := migrasyontest.New(t, migrasyontest.Project{Dir: "../migrations"})

	project.Validate(t).NoProblems(t)

	if got := project.Up(t).Versions(); len(got) != 3 {
		t.Fatalf("migrasyon up applied %v, want all three", got)
	}

	if refusal := project.Refused(t, "down"); !refusal.Says("cannot be undone") {
		t.Errorf("migrasyon down failed with %q, want the tier's declared reason", refusal.Error)
	}
}
