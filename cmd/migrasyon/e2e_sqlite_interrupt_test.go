package main

import (
	"bufio"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// endlessMigration is a statement that would run for minutes, so only an interrupt ends it in a test.
const endlessMigration = "-- migrasyon:up\nCREATE TABLE e2e_big AS WITH RECURSIVE c(x) AS (SELECT 1 UNION ALL " +
	"SELECT x + 1 FROM c) SELECT x FROM c LIMIT 1000000000;\n\n-- migrasyon:down\nDROP TABLE e2e_big;\n"

// quickMigration is what the endless one is edited into once it has been interrupted.
const quickMigration = "-- migrasyon:up\nCREATE TABLE e2e_big (x integer);\n\n-- migrasyon:down\nDROP TABLE e2e_big;\n"

// endlessName is the endless migration's file, the third in the staged project.
const endlessName = "20260929T120200_fill_e2e_big.sql"

// ending is how an interrupted migrasyon up came to an end: what it wrote, how it exited, and whether it did.
type ending struct {
	stderr string
	err    error
	exited bool
}

// interrupted starts migrasyon up as a process of its own and sends it SIGINT once it is applying.
func interrupted(t *testing.T, config string) ending {
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

	defer func() { _ = cmd.Process.Kill() }()

	lines := bufio.NewScanner(stderr)
	for lines.Scan() {
		if strings.Contains(lines.Text(), "connected") {
			break
		}
	}

	time.Sleep(500 * time.Millisecond)

	if err := cmd.Process.Signal(os.Interrupt); err != nil {
		t.Fatalf("interrupting migrasyon up: %v", err)
	}

	var rest strings.Builder

	read := make(chan struct{})

	go func() {
		defer close(read)

		for lines.Scan() {
			rest.WriteString(lines.Text() + "\n")
		}
	}()

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	select {
	case err := <-done:
		<-read

		return ending{stderr: rest.String(), err: err, exited: true}
	case <-time.After(15 * time.Second):
		return ending{}
	}
}

func TestAnInterruptedRunStopsItsStatementAndLeavesTheDatabaseUsable(t *testing.T) {
	config, _ := sqliteProject(t)
	dir := filepath.Dir(config)

	if err := migrasyon(t, config, "up"); err != nil {
		t.Fatalf("applying the first two migrations: %v", err)
	}

	endless := filepath.Join(dir, "migrations", endlessName)
	if err := os.WriteFile(endless, []byte(endlessMigration), 0o600); err != nil {
		t.Fatalf("staging the endless migration: %v", err)
	}

	end := interrupted(t, config)
	if !end.exited {
		t.Fatal("migrasyon up was still running 15s after SIGINT, so the statement was not interrupted")
	}

	var exit *exec.ExitError
	if !errors.As(end.err, &exit) || exit.ExitCode() == 0 {
		t.Fatalf("migrasyon up exited with %v after SIGINT, want a failure", end.err)
	}

	if !strings.Contains(end.stderr, "context canceled") || !strings.Contains(end.stderr, endlessName) {
		t.Errorf("migrasyon up did not fail on the interrupt in the endless migration:\n%s", end.stderr)
	}

	status := migrasyonOut(t, config, "status")
	if got := stateOf(t, status, "fill_e2e_big"); got != "pending" {
		t.Errorf("the interrupted migration is %q, want pending -- its rollback took what it wrote", got)
	}

	if err := os.WriteFile(endless, []byte(quickMigration), 0o600); err != nil {
		t.Fatalf("editing the migration: %v", err)
	}

	if err := migrasyon(t, config, "up"); err != nil {
		t.Fatalf("the next migrasyon up over the interrupted database: %v", err)
	}

	t.Log("Ctrl-C is the everyday cancel. The statement stopped inside the engine, the transaction")
	t.Log("rolled back, the run recorded its failure and released its lock -- so the next migrasyon up")
	t.Log("ran at once rather than waiting out a lock a dead run still held.")
}
