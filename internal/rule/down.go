package rule

import (
	"github.com/tosunlab/migrasyon/internal/ir"
)

// MissingDown is the id of the rule that reports a migration whose undo migrasyon had to derive.
const MissingDown ID = "missing-down"

// missingDown reports a migration that authors no down, whose rollback is therefore an inverse.
func missingDown() Rule {
	return Rule{
		ID:      MissingDown,
		Default: SeverityWarn,
		Summary: "a migration authors no down, so its rollback runs an inverse migrasyon derived",
		Judge:   judgeMissingDown,
	}
}

// judgeMissingDown reports the one answer to "what undoes this" that nobody wrote down.
func judgeMissingDown(s Subject) []Finding {
	if s.Undo != UndoDerived {
		return nil
	}

	return []Finding{{Direction: ir.Down, Detail: "this migration authors no down, so a rollback " +
		"inverts whatever its up emits when the rollback runs -- which need not be what it emitted " +
		"when it was applied: write a down, or declare the migration irreversible with a reason"}}
}
