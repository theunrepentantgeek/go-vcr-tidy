package analyzer

import "github.com/theunrepentantgeek/go-vcr-tidy/internal/interaction"

// Result is returned by an analyzer after processing an interaction.
type Result struct {
	// Finished indicates the analyzer has completed its work, and can be discarded from the set of active analyzers.
	Finished bool
	// Spawn contains any new analyzers that should be added to the active set, allowing one analyzer to create others.
	Spawn []Interface
	// Excluded lists interactions that should be excluded from the final output.
	Excluded []interaction.Interface
}

// Finished creates a Result indicating the analyzer is finished.
func Finished() Result {
	return Result{
		Finished: true,
	}
}

// Finished with Exclusions creates a Result indicating the analyzer is finished and listing interactions to exclude.
func FinishedWithExclusions(excluded []interaction.Interface) Result {
	return Result{
		Finished: true,
		Excluded: excluded,
	}
}
