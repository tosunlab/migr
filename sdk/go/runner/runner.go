package runner

import (
	"os"

	migrasyon "github.com/tosunlab/migrasyon/sdk/go"
)

// Main serves this program as a migrasyon runtime and exits with the status docs/protocol.md §9 gives.
func Main() {
	os.Exit(migrasyon.Run())
}
