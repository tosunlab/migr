package main

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// A chunked migration killed with SIGKILL partway and carried on by the next migrasyon up, untagged
// because this engine needs no daemon. The kill is a real one, against a real process: the test
// binary re-executes itself as migrasyon, since an in-process run has no process to kill.
//
// What the rows hold afterwards is asked by a migration rather than by this file, which composes no
// SQL above a driver package (invariant #3): the last migration reads the counts through
// host.db.query, operator SQL as every fixture's is, and fails unless every row was written once.

// asBinary is the variable that makes this test binary run as migrasyon itself, for a test to kill.
const asBinary = "MIGRASYON_E2E_AS_BINARY"

// TestMain runs the binary as migrasyon where a test asked it to, and as the test suite otherwise.
func TestMain(m *testing.M) {
	if os.Getenv(asBinary) == "1" {
		os.Exit(run())
	}

	os.Exit(m.Run())
}

// chunkedRows is how many rows the killed migration pages through.
const chunkedRows = "2000"

// fillMigration creates the two tables and fills one, through the escape hatch beside an index.
const fillMigration = `package migrations

import (
	"context"

	migrasyon "github.com/tosunlab/migrasyon/sdk/go"
)

func init() { migrasyon.Register(FillPeople{}) }

// FillPeople creates a table of people and the archive they are copied into, and fills the first.
type FillPeople struct{}

// Up creates both tables, then fills people.
func (FillPeople) Up(ctx context.Context, s *migrasyon.Schema) error {
	if err := s.CreateTable("people").
		Column("id", migrasyon.Int64()).
		Column("email", migrasyon.String(255)).
		Column("nickname", migrasyon.String(64)).Null().
		PrimaryKey("id").
		Apply(ctx); err != nil {
		return err
	}

	return s.CreateTable("people_archive").
		Column("id", migrasyon.Int64()).
		Column("email", migrasyon.String(255)).
		PrimaryKey("id").
		Apply(ctx)
}

// Down drops both.
func (FillPeople) Down(ctx context.Context, s *migrasyon.Schema) error {
	if err := s.Table("people_archive").Drop().Apply(ctx); err != nil {
		return err
	}

	return s.Table("people").Drop().Apply(ctx)
}
`

// seedMigration inserts the rows, beside an index so the migration has an operation to run.
const seedMigration = `package migrations

import (
	"context"

	migrasyon "github.com/tosunlab/migrasyon/sdk/go"
)

func init() { migrasyon.Register(SeedPeople{}) }

// SeedPeople inserts the rows the chunks page through.
type SeedPeople struct{}

// Up inserts them and indexes the email.
func (SeedPeople) Up(ctx context.Context, s *migrasyon.Schema) error {
	if _, err := s.Exec(ctx, "WITH RECURSIVE g(i) AS (SELECT 1 UNION ALL SELECT i + 1 FROM g "+
		"WHERE i < ` + chunkedRows + `) INSERT INTO people (id, email) SELECT i, 'user' || i FROM g"); err != nil {
		return err
	}

	return s.Table("people").AddIndex("people_email_idx", "email").Apply(ctx)
}

// Irreversible says why.
func (SeedPeople) Irreversible() string { return "the rows it inserted are not told apart" }
`

// archiveMigration copies every row first and fills the nicknames after, all in chunks.
const archiveMigration = `package migrations

import (
	"context"

	migrasyon "github.com/tosunlab/migrasyon/sdk/go"
)

func init() { migrasyon.Register(ArchivePeople{}) }

// ArchivePeople copies people into the archive and fills their nicknames.
type ArchivePeople struct{}

// Up copies, then fills.
func (ArchivePeople) Up(ctx context.Context, s *migrasyon.Schema) error {
	return s.Table("people").
		CopyInto("people_archive",
			migrasyon.CopyColumn{From: "id", To: "id"}, migrasyon.CopyColumn{From: "email", To: "email"}).
		Backfill("nickname", migrasyon.String(64), migrasyon.Literal("anonymous")).
		Apply(ctx)
}

// Irreversible says why.
func (ArchivePeople) Irreversible() string { return "the rows these wrote are recorded nowhere" }
`

// checkMigration reads the rows the chunked migration wrote, and fails unless each was written once.
const checkMigration = `package migrations

import (
	"context"
	"fmt"

	migrasyon "github.com/tosunlab/migrasyon/sdk/go"
)

func init() { migrasyon.Register(CheckPeople{}) }

// CheckPeople asserts what the chunked migration left, then indexes the archive.
type CheckPeople struct{}

// Up fails unless the archive holds every row once and no nickname is NULL.
func (CheckPeople) Up(ctx context.Context, s *migrasyon.Schema) error {
	rows, err := s.Query(ctx, "SELECT (SELECT count(*) FROM people_archive), "+
		"(SELECT count(DISTINCT id) FROM people_archive), "+
		"(SELECT count(*) FROM people WHERE nickname IS NULL)")
	if err != nil {
		return err
	}

	got := fmt.Sprintf("%s %s %s", *rows.Values[0][0], *rows.Values[0][1], *rows.Values[0][2])
	if got != "` + chunkedRows + ` ` + chunkedRows + ` 0" {
		return fmt.Errorf("the archive and the nicknames read %q after the resumed run", got)
	}

	return s.Table("people_archive").AddIndex("people_archive_email_idx", "email").Apply(ctx)
}

// Irreversible says why.
func (CheckPeople) Irreversible() string { return "it only asserts" }
`

// chunkProject is a Go project on SQLite whose third migration runs in chunks of 50, paced at 5ms.
func chunkProject(t *testing.T) string {
	t.Helper()

	config, _ := sqliteGoProject(t)
	dir := filepath.Join(filepath.Dir(config), "migrations")

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("reading the staged migrations: %v", err)
	}

	for _, entry := range entries {
		if err := os.Remove(filepath.Join(dir, entry.Name())); err != nil {
			t.Fatalf("clearing %s: %v", entry.Name(), err)
		}
	}

	for name, body := range map[string]string{
		"20260929T170000_fill_people.go":    fillMigration,
		"20260929T170050_seed_people.go":    seedMigration,
		"20260929T170100_archive_people.go": archiveMigration,
		"20260929T170200_check_people.go":   checkMigration,
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o600); err != nil {
			t.Fatalf("staging %s: %v", name, err)
		}
	}

	appendConfig(t, config, "safety:\n  chunk_rows: 50\n  chunk_pause: 5ms\n")

	return config
}

// appendConfig adds a block to a staged migrasyon.yaml.
func appendConfig(t *testing.T, config, block string) {
	t.Helper()

	body, err := os.ReadFile(config)
	if err != nil {
		t.Fatalf("reading migrasyon.yaml: %v", err)
	}

	if err := os.WriteFile(config, append(body, []byte(block)...), 0o600); err != nil {
		t.Fatalf("writing migrasyon.yaml: %v", err)
	}
}

// killedAfter starts migrasyon up as a process of its own and kills it once enough chunks have committed.
func killedAfter(t *testing.T, config string, chunks int) []string {
	t.Helper()

	cmd := exec.Command(os.Args[0], "--config", config, "up", "-v")
	cmd.Env = append(os.Environ(), asBinary+"=1")

	stderr, err := cmd.StderrPipe()
	if err != nil {
		t.Fatalf("piping stderr: %v", err)
	}

	if err := cmd.Start(); err != nil {
		t.Fatalf("starting migrasyon up: %v", err)
	}

	defer func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	}()

	var committed []string

	lines := bufio.NewScanner(stderr)
	deadline := time.AfterFunc(2*time.Minute, func() { _ = cmd.Process.Kill() })

	defer deadline.Stop()

	for lines.Scan() {
		if !strings.Contains(lines.Text(), "committed a chunk") {
			continue
		}

		committed = append(committed, lines.Text())
		if len(committed) == chunks {
			if err := cmd.Process.Kill(); err != nil {
				t.Fatalf("killing migrasyon up: %v", err)
			}

			break
		}
	}

	return committed
}

func TestAChunkedMigrationKilledPartwayIsCarriedOnByTheNextUp(t *testing.T) {
	config := chunkProject(t)

	if err := migrasyon(t, config, "up", "--step", "2"); err != nil {
		t.Fatalf("applying the tables and their rows: %v", err)
	}

	planned := migrasyonOut(t, config, "plan", "--step", "1")
	for _, want := range []string{
		`chunks: 50 rows at a time by "people" (id), pausing 5ms between them`,
		`WHERE "people"."id" > ? AND "people"."id" <= ?`,
	} {
		if !strings.Contains(planned, want) {
			t.Errorf("migrasyon plan does not show %q:\n%s", want, planned)
		}
	}

	committed := killedAfter(t, config, 5)
	if len(committed) != 5 || !strings.Contains(committed[4], "through=250") {
		t.Fatalf("migrasyon up was killed after %d chunks, the last %q, want the fifth through key 250",
			len(committed), committed)
	}

	status := migrasyonOut(t, config, "status")
	if !strings.Contains(status, "stopped running in chunks: operation 1, 5 chunks and 250 rows "+
		"written, through key (250) -- the next migrasyon up carries on from there") {
		t.Errorf("migrasyon status does not say how far the killed run got:\n%s", status)
		t.Log("SIGKILL runs no code at all, so the dirty row and the checkpoint have to have")
		t.Log("committed with the chunk they describe -- nothing after it got the chance to.")
	}

	resumed := migrasyonOut(t, config, "up")
	if !strings.Contains(resumed, "resumed") {
		t.Errorf("the next migrasyon up did not say it carried the run on:\n%s", resumed)
	}

	t.Log("The last migration read the archive and the nicknames and passed: 2000 rows copied")
	t.Log("once each, the five chunks before the kill not copied again, and no NULL left behind.")
}
