---
title: Go
---

# Writing migrations in Go

A Go migration is a type with an `Up` method that describes the change through a schema builder.
migrasyon compiles what it describes for whichever database it runs against, and derives the down from it
unless you write one.

## Set up

Migrations live in a package of their own inside your module, and the SDK is a separate module, so
it brings no database driver into your application:

```console
$ go get github.com/tosunlab/migrasyon/sdk/go
```

```yaml
# migrasyon.yaml, beside go.mod or anywhere inside the module
version: 1
migrations:
  dir: ./migrations
  language: go
databases:
  default:
    driver: postgres
    dsn: ${DATABASE_URL}
```

```console
$ migrasyon create create_accounts
created migrations/20261023T120000_create_accounts.go
```

The template explains itself. A migration registers itself from `init`, and its version comes from
its file name:

```go
package migrations

import (
	"context"

	migrasyon "github.com/tosunlab/migrasyon/sdk/go"
)

func init() { migrasyon.Register(CreateAccounts{}) }

// CreateAccounts adds the accounts table.
type CreateAccounts struct{}

// Up creates the table with its key inline.
func (CreateAccounts) Up(ctx context.Context, s *migrasyon.Schema) error {
	return s.CreateTable("accounts").
		Column("id", migrasyon.Int64()).
		Column("email", migrasyon.String(255)).
		Column("created_at", migrasyon.TimestampTZ()).Default(migrasyon.Call(migrasyon.FuncCurrentTimestamp)).
		PrimaryKey("id").
		Apply(ctx)
}
```

**How migrasyon runs it:** it generates a small runner in `.migrasyon/go/`, builds it with your Go toolchain,
and speaks to it over a pipe. The build is cached by content. The machine running `migrasyon` therefore
needs Go, and the `.migrasyon/` directory ignores itself, so it never shows up in version control.

## Describing a change

- **`s.CreateTable(name)`** builds a new table, with its primary key stated inline.
- **`s.Table(name)`** changes one that exists. Its chain adds, drops, renames and alters columns,
  indexes, unique constraints, checks, foreign keys and the primary key.
- **`s.View(name)`** creates, replaces or drops a view.

Two conventions are built into the API, so getting them wrong isn't possible:
- **a column is `NOT NULL` unless the chain says `Null()`;**
- **nothing is emitted until `Apply`.** A chain built without it fails the migration, so a migration
  can't report a change it never made.

Types are logical, and each driver maps them to native ones: `String(n)`, `Text()`, `Bool()`,
`Int16()`, `Int32()`, `Int64()`, `Float32()`, `Float64()`, `Decimal(p, s)`, `UUID()`, `JSON()`,
`Timestamp()`, `TimestampTZ()`, `Date()`, `Time()`, `Binary()`, `Enum(values...)` and `Array(of)`.

```go
func (AddNickname) Up(ctx context.Context, s *migrasyon.Schema) error {
	return s.Table("accounts").
		AddColumn("nickname", migrasyon.String(64)).Null().
		AddIndex("idx_accounts_nickname", "nickname").
		Apply(ctx)
}
```

## The down

There are three ways a migration can be undone, and it uses exactly one of them:

- **Derived.** Write no `Down`, and migrasyon inverts what `Up` emitted. `migrasyon validate` warns
  (`missing-down`) because the inverse is computed when the rollback runs, from what `Up` emits then.
- **Authored.** Write a `Down`. A drop records what it removes with `Was`, so the drop can itself be
  reversed:

  ```go
  func (AddNickname) Down(ctx context.Context, s *migrasyon.Schema) error {
  	return s.Table("accounts").
  		DropIndex("idx_accounts_nickname").Was(migrasyon.Columns("nickname")).
  		DropColumn("nickname").Was(migrasyon.Column{Name: "nickname", Type: migrasyon.String(64), Nullable: true}).
  		Apply(ctx)
  }
  ```

- **Declared irreversible.** A rollback halts before its first statement and prints your reason:

  ```go
  func (AddTier) Irreversible() string {
  	return "the tiers written after it ran are not recorded anywhere a rollback could restore them from"
  }
  ```

  Writing both `Down` and `Irreversible` is a registration error.

## Asking the database

A migration can read the database and decide what to emit from the answer:

```go
func (AddTier) Up(ctx context.Context, s *migrasyon.Schema) error {
	rows, err := s.Query(ctx, "SELECT count(*) FROM accounts")
	if err != nil {
		return err
	}

	tier := s.Table("accounts").AddColumn("tier", migrasyon.String(16))
	if *rows.Values[0][0] != "0" {
		tier = tier.Default(migrasyon.Literal("free"))
	}

	return tier.Apply(ctx)
}
```

- **Every value is the server's own text**, or `nil` for a null.
- **A read runs against the database as it is now.** So `migrasyon plan` over a queue where an earlier
  pending migration creates the table fails on the read. Apply up to that point first
  (`migrasyon up --to <version>`).
- **The placeholder is the engine's own:** `$1` on PostgreSQL, `?` on MySQL and SQLite.
  `s.Info(ctx)` says which driver this is.
- **A read makes the plan depend on the data.** A preview and the run that follows it can take
  different branches if the data changed between them.

`s.Exec(ctx, sql)` sends a statement the builder can't express. It's the last resort:
- it runs **outside the migration's transaction**, so a later failure doesn't undo it;
- `migrasyon plan` **records it instead of running it**;
- a migration that calls it must write its down, since a statement has no inverse to derive.

## Data migrations

`Table.Backfill`, `Table.CopyInto` and `Table.Transform` write rows. A migration holding nothing
else runs them **in chunks** of `safety.chunk_rows`, each committed on its own, and a stopped run
carries on from its last chunk. A migration that also changes the schema runs them as single
statements inside its transaction.

To change a column's type with no downtime, declare the change once and register one phase per file.
Each phase is its own migration, applied a deploy apart:

```go
func init() {
	migrasyon.Register(migrasyon.ReplaceColumn("users", "email", migrasyon.String(255)).
		With(migrasyon.String(320)).Expand())
}
```

The phases are `Expand`, `Backfill`, `Swap` and `Contract`. The
[zero-downtime guide](../zero-downtime.md#_6-changing-a-column-across-deploys) explains the deploy
sequence and `--release`.

## Large tables

A migration implementing `Online` is performed through a shadow copy of its table, kept current from
the server's replication log and swapped in once you confirm it with `migrasyon cutover`:

```go
func (IndexPeople) Online() migrasyon.Strategy { return migrasyon.ShadowTable }
```

The [driver pages](../drivers/postgres.md#shadow-tables) list what each server needs, and the
[zero-downtime guide](../zero-downtime.md#_4-shadow-table-migrations) covers the whole workflow.

## Testing

`github.com/tosunlab/migrasyon/sdk/go/migrasyontest` has three helpers:
- **`Up` and `Down`** capture what a half emits;
- **`Reversal`** reports how a migration is undone;
- **`New`** builds a throwaway project and drives the real `migrasyon` binary against it.

**Tests go in a package of their own, not in the migrations directory.** Discovery accepts only
migrations there, so a `migrations_test.go` beside them fails every command:

```go
package migrationtest

import (
	"slices"
	"testing"

	migrasyon "github.com/tosunlab/migrasyon/sdk/go"
	"github.com/tosunlab/migrasyon/sdk/go/migrasyontest"

	"example.com/app/migrations"
)

func TestAddNicknameAddsAColumnAndItsIndex(t *testing.T) {
	got := migrasyontest.Up(t, migrations.AddNickname{}).Kinds()

	if !slices.Equal(got, []migrasyon.Kind{migrasyon.KindAddColumn, migrasyon.KindAddIndex}) {
		t.Errorf("Up emitted %v", got)
	}
}

// stage answers AddTier's one read as a table already holding three accounts.
func stage() migrasyontest.Stage {
	return migrasyontest.Stage{Reads: map[string]migrasyon.Rows{
		"SELECT count(*) FROM accounts": migrasyontest.Answer([]string{"count"}, []string{"3"}),
	}}
}

func TestAddTierDefaultsOnlyWhereAccountsExist(t *testing.T) {
	stage().Up(t, migrations.AddTier{}).Is(t, migrasyon.AddColumn{Table: "accounts", Column: migrasyon.Column{
		Name: "tier", Type: migrasyon.String(16), Default: migrasyon.Literal("free"),
	}})
}

func TestAddTierCannotBeUndone(t *testing.T) {
	if r := stage().Reversal(t, migrations.AddTier{}); r.Undo != migrasyon.UndoDeclared {
		t.Errorf("Undo = %s, want declared", r.Undo)
	}
}
```

A read the stage doesn't answer is refused rather than answered empty, so a test never takes a branch
by accident. `Reversal` runs `Up` too, so it needs the same stage.

## What `migrasyon validate` checks

`migrasyon validate` builds every migration, compiles both directions, and checks that each one can be
undone. It opens no database, so a migration that reads is reported **unchecked** rather than
failed. `migrasyon validate --connect` also asks the target server what it refuses, and checks every
reference a migration makes. `--require-safe` gates each up half on its lock cost.

## Example projects

[`examples/go-postgres`](../../examples/go-postgres/README.md) is a Go project on PostgreSQL with
the three kinds of down and tests using every `migrasyontest` helper.
[`examples/zero-downtime`](../../examples/zero-downtime/README.md) replaces a column with
`ReplaceColumn` across five deploys while an app keeps serving.
