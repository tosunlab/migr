package sql

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/tosunlab/migrasyon/internal/engine"
	"github.com/tosunlab/migrasyon/internal/ir"
)

const concurrently = `-- migrasyon:up
CREATE TABLE users (id int PRIMARY KEY, email text);

-- migrasyon:no-transaction
CREATE INDEX CONCURRENTLY users_email ON users (email);
INSERT INTO users VALUES (1, 'a');

-- migrasyon:down
-- migrasyon:no-transaction
DROP INDEX CONCURRENTLY users_email;
DROP TABLE users;
`

// standaloneOf reports which of a direction's statements declared they cannot run inside a transaction.
func standaloneOf(t *testing.T, ops ir.Operations) []bool {
	t.Helper()

	out := make([]bool, len(ops))

	for i, op := range ops {
		raw, ok := op.(ir.Raw)
		if !ok {
			t.Fatalf("operation %d is a %s, and a .sql file emits raw alone", i+1, op.Kind())
		}

		out[i] = raw.NoTransaction
	}

	return out
}

func TestTheMarkerDeclaresTheOneStatementBelowIt(t *testing.T) {
	source := sourceOf(t, concurrently)

	for dir, want := range map[ir.Direction][]bool{
		ir.Up:   {false, true, false},
		ir.Down: {true, false},
	} {
		got := standaloneOf(t, emit(t, source, dir))
		if !slices.Equal(got, want) {
			t.Errorf("%s declares %v, want %v", dir, got, want)
		}
	}

	for _, op := range emit(t, source, ir.Up) {
		if sql := op.(ir.Raw).SQL; strings.Contains(sql, MarkerNoTransaction) {
			t.Errorf("the statement sent still carries the marker: %q", sql)
		}
	}

	t.Log("A marker declares a statement rather than a file, so one migration can create a table")
	t.Log("in a transaction and build its index concurrently outside one.")
}

func TestAFileWithNoMarkerEmitsTheBytesItAlwaysDid(t *testing.T) {
	ops := emit(t, sourceOf(t, "-- migrasyon:up\nCREATE TABLE t (id int);\n"), ir.Up)

	if got := ir.VersionOf(ops); got != 1 {
		t.Errorf("a .sql migration with no marker states irVersion %d, want 1", got)
		t.Log("Its checksum covers those bytes, so a field that appeared unasked would read as")
		t.Log("drift on every .sql migration applied before this one existed.")
	}
}

func TestAMarkerInsideQuotingOrAStatementIsNotAMarker(t *testing.T) {
	quoted := "-- migrasyon:up\nCREATE FUNCTION f() RETURNS int AS $$\n-- migrasyon:no-transaction\nSELECT 1;\n" +
		"$$ LANGUAGE sql;\n"
	if got := standaloneOf(t, emit(t, sourceOf(t, quoted), ir.Up)); !slices.Equal(got, []bool{false}) {
		t.Errorf("a marker inside a function body declared %v", got)
	}

	inline := "-- migrasyon:up\nCREATE INDEX i ON t (c); -- migrasyon:no-transaction\nSELECT 1;\n"
	if got := standaloneOf(t, emit(t, sourceOf(t, inline), ir.Up)); !slices.Equal(got, []bool{false, false}) {
		t.Errorf("a marker trailing a statement on its own line declared %v", got)
	}
}

func TestAMarkerOutOfPlaceIsRefusedNamingItsLine(t *testing.T) {
	for name, c := range map[string]struct {
		body string
		line string
	}{
		"inside a statement": {
			body: "-- migrasyon:up\nCREATE TABLE t (\n-- migrasyon:no-transaction\n  id int);\n",
			line: "line 3",
		},
		"twice in a row": {
			body: "-- migrasyon:up\n-- migrasyon:no-transaction\n-- migrasyon:no-transaction\nVACUUM;\n",
			line: "line 3",
		},
		"at the end of a section": {
			body: "-- migrasyon:up\nVACUUM;\n-- migrasyon:no-transaction\n-- migrasyon:down\nSELECT 1;\n",
			line: "line 3",
		},
		"above the up marker": {
			body: "-- migrasyon:no-transaction\n-- migrasyon:up\nVACUUM;\n",
			line: "line 1",
		},
		"below a declaration": {
			body: "-- migrasyon:up\nVACUUM;\n-- migrasyon:irreversible: gone\n-- migrasyon:no-transaction\n",
			line: "line 4",
		},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := sourceOf(t, c.body).Emit(t.Context(), file, ir.Up, engine.ModeApply)

			if !errors.Is(err, ErrMalformed) {
				t.Fatalf("Emit returned %v, want the file refused", err)
			}

			for _, want := range []string{c.line, MarkerNoTransaction} {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("the refusal does not name %q: %v", want, err)
				}
			}
		})
	}

	t.Log("A marker that declares nothing is a migration whose author believed a statement runs")
	t.Log("outside a transaction when it does not -- read in silence, CIC fails at the server.")
}
