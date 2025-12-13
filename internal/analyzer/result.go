package analyzer

// Result is returned by an analyzer after processing an interaction.
type Result struct {
	// Finished indicates the analyzer has completed its work, and can be discarded from the set of active analyzers.
	Finished bool
	// Spawn contains any new analyzers that should be added to the active set, allowing one analyzer to create others.
	Spawn []Interface
}
