# Integrate DetectDeferredCreation Analyzer — Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Wire the existing `DetectDeferredCreation` analyzer into the CLI, options, and public API layers, following the `DetectDeletion` integration pattern.

**Architecture:** Four files are modified across two packages. The public API gets a new `ReduceDeferredCreationMonitoring()` option function. The CLI layer gets a new `--clean-deferred-creations` flag, a `ShouldCleanDeferredCreations()` method, and wiring in `Options()`. Tests are updated at both layers.

**Tech Stack:** Go, gomega (assertions), kong (CLI parsing)

**Spec:** `docs/superpowers/specs/2026-03-16-integrate-deferred-creation-design.md`

---

## Chunk 1: Integration

### Task 1: Add `ReduceDeferredCreationMonitoring` to public API

**Files:**
- Modify: `pkg/vcrcleaner/options.go`

- [ ] **Step 1: Add `ReduceDeferredCreationMonitoring` function before `ReduceDeleteMonitoring`**

Insert before the existing `ReduceDeleteMonitoring` function:

```go
// ReduceDeferredCreationMonitoring adds an analyzer that reduces deferred creation monitoring noise.
func ReduceDeferredCreationMonitoring() Option {
	return func(c *cleaner.Cleaner) {
		c.AddAnalyzers(generic.NewDetectDeferredCreation())
	}
}
```

- [ ] **Step 2: Verify compilation**

Run: `go build ./...`
Expected: success (no compilation errors)

- [ ] **Step 3: Commit**

```bash
git add pkg/vcrcleaner/options.go
git commit -m "feat: add ReduceDeferredCreationMonitoring option function"
```

---

### Task 2: Add CLI flag and wiring for deferred creations

**Files:**
- Modify: `internal/cmd/cleaning_options.go`

- [ ] **Step 1: Add `DeferredCreations` field to `CleaningOptions` struct**

The struct should become (note lifecycle ordering — creation before deletion):

```go
type CleaningOptions struct {
	All               *bool `help:"Clean all supported interaction types."`
	DeferredCreations *bool `help:"Clean deferred creation interactions."`
	Deletes           *bool `help:"Clean delete interactions."`

	Azure AzureCleaningOptions `embed:"" prefix:"azure-"`
}
```

- [ ] **Step 2: Add `ShouldCleanDeferredCreations` method**

Add after `ShouldCleanDeletes`, following the same pattern:

```go
func (opt *CleaningOptions) ShouldCleanDeferredCreations() bool {
	return opt.coalesce(
		opt.DeferredCreations,
		opt.All)
}
```

- [ ] **Step 3: Wire deferred creations into `Options()` before deletes**

Update the `Options()` method so deferred creations is checked first (lifecycle ordering):

```go
func (opt *CleaningOptions) Options() []vcrcleaner.Option {
	var result []vcrcleaner.Option
	if opt.ShouldCleanDeferredCreations() {
		result = append(result, vcrcleaner.ReduceDeferredCreationMonitoring())
	}

	if opt.ShouldCleanDeletes() {
		result = append(result, vcrcleaner.ReduceDeleteMonitoring())
	}

	azureOptions := opt.Azure.Options(opt.All)
	result = append(result, azureOptions...)

	return result
}
```

- [ ] **Step 4: Verify compilation**

Run: `go build ./...`
Expected: success

- [ ] **Step 5: Commit**

```bash
git add internal/cmd/cleaning_options.go
git commit -m "feat: add --clean-deferred-creations CLI flag and wiring"
```

---

### Task 3: Add tests for `ShouldCleanDeferredCreations`

**Files:**
- Modify: `internal/cmd/cleaning_options_test.go`

- [ ] **Step 1: Write `TestCleaningOptions_ShouldCleanDeferredCreations` table test**

Add after the existing `TestCleaningOptions_ShouldCleanDeletes` test, mirroring its structure:

```go
// CleaningOptions.ShouldCleanDeferredCreations Tests

func TestCleaningOptions_ShouldCleanDeferredCreations(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		deferredCreations *bool
		expected          bool
	}{
		"WithDeferredCreationsTrue_ReturnsTrue": {
			deferredCreations: toPtr(true),
			expected:          true,
		},
		"WithDeferredCreationsFalse_ReturnsFalse": {
			deferredCreations: toPtr(false),
			expected:          false,
		},
		"WithDeferredCreationsNil_ReturnsFalse": {
			expected: false,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)

			opt := &CleaningOptions{
				DeferredCreations: c.deferredCreations,
			}

			result := opt.ShouldCleanDeferredCreations()

			g.Expect(result).To(Equal(c.expected))
		})
	}
}
```

- [ ] **Step 2: Run the new test to verify it passes**

Run: `go test ./internal/cmd/ -run TestCleaningOptions_ShouldCleanDeferredCreations -v`
Expected: all 3 subtests PASS

- [ ] **Step 3: Commit**

```bash
git add internal/cmd/cleaning_options_test.go
git commit -m "test: add ShouldCleanDeferredCreations tests"
```

---

### Task 4: Update `TestCleaningOptions_Options` test cases

**Files:**
- Modify: `internal/cmd/cleaning_options_test.go`

- [ ] **Step 1: Add `deferredCreations` field to the test struct and add new cases**

Update the test struct to include `deferredCreations *bool`, and add new test cases. The full updated test function:

```go
func TestCleaningOptions_Options(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		deferredCreations *bool
		deletes           *bool
		azureAll          *bool
		expectedCount     int
	}{
		"WithNoOptionsSet_ReturnsEmptySlice": {
			expectedCount: 0,
		},
		"WithOnlyDeferredCreationsSet_ReturnsOneDeferredCreationOption": {
			deferredCreations: toPtr(true),
			expectedCount:     1,
		},
		"WithOnlyDeletesSet_ReturnsOneDeleteOption": {
			deletes:       toPtr(true),
			expectedCount: 1,
		},
		"WithDeferredCreationsAndDeletes_ReturnsTwoOptions": {
			deferredCreations: toPtr(true),
			deletes:           toPtr(true),
			expectedCount:     2,
		},
		"WithOnlyAzureAllSet_ReturnsFourAzureOptions": {
			azureAll:      toPtr(true),
			expectedCount: 4,
		},
		"WithDeletesAndAzureAll_ReturnsFiveOptions": {
			deletes:       toPtr(true),
			azureAll:      toPtr(true),
			expectedCount: 5,
		},
		"WithDeferredCreationsAndAzureAll_ReturnsFiveOptions": {
			deferredCreations: toPtr(true),
			azureAll:          toPtr(true),
			expectedCount:     5,
		},
		"WithAllOptions_ReturnsSixOptions": {
			deferredCreations: toPtr(true),
			deletes:           toPtr(true),
			azureAll:          toPtr(true),
			expectedCount:     6,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)

			opt := &CleaningOptions{
				DeferredCreations: c.deferredCreations,
				Deletes:           c.deletes,
				Azure: AzureCleaningOptions{
					All: c.azureAll,
				},
			}

			result := opt.Options()

			g.Expect(result).To(HaveLen(c.expectedCount))
		})
	}
}
```

Note: The `deferredCreations` field appears first in the struct literal (lifecycle ordering), and the existing "WithDeletesAndAzureAll" case is renamed to "WithDeletesAndAzureAll_ReturnsFiveOptions" to fix the count in the name.

- [ ] **Step 2: Run the updated test to verify all cases pass**

Run: `go test ./internal/cmd/ -run TestCleaningOptions_Options -v`
Expected: all 8 subtests PASS

- [ ] **Step 3: Commit**

```bash
git add internal/cmd/cleaning_options_test.go
git commit -m "test: update Options test cases for deferred creation"
```

---

### Task 5: Update `TestBuildOptions` test cases

**Files:**
- Modify: `internal/cmd/clean_command_test.go`

- [ ] **Step 1: Add `deferredCreations` field and new test cases**

Update the test struct to include `deferredCreations *bool`, wire it in the test body, and add new cases. The full updated test function:

```go
//nolint:funlen // Table test cases are extensive but clear
func TestBuildOptions(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		deferredCreations      *bool
		deletes                *bool
		longRunningOperations  *bool
		resourceModifications  *bool
		resourceDeletions      *bool
		expectedOptionsCount   int
		expectedErrorSubstring string
	}{
		"WithNoOptionsSpecified_ReturnsError": {
			expectedErrorSubstring: "no cleaning options specified",
		},
		"WithOnlyDeferredCreationsSet_ReturnsDeferredCreationOption": {
			deferredCreations:    toPtr(true),
			expectedOptionsCount: 1,
		},
		"WithOnlyDeletesSet_ReturnsDeleteOption": {
			deletes:              toPtr(true),
			expectedOptionsCount: 1,
		},
		"WithOnlyLongRunningOperationsSet_ReturnsLROOption": {
			longRunningOperations: toPtr(true),
			expectedOptionsCount:  1,
		},
		"WithOnlyResourceModificationsSet_ReturnsResourceModificationOption": {
			resourceModifications: toPtr(true),
			expectedOptionsCount:  1,
		},
		"WithOnlyResourceDeletionsSet_ReturnsResourceDeletionOption": {
			resourceDeletions:    toPtr(true),
			expectedOptionsCount: 1,
		},
		"WithDeferredCreationsAndDeletes_ReturnsTwoOptions": {
			deferredCreations:    toPtr(true),
			deletes:              toPtr(true),
			expectedOptionsCount: 2,
		},
		"WithDeletesAndLongRunningOperations_ReturnsTwoOptions": {
			deletes:               toPtr(true),
			longRunningOperations: toPtr(true),
			expectedOptionsCount:  2,
		},
		"WithAllOptionsSet_ReturnsAllOptions": {
			deferredCreations:     toPtr(true),
			deletes:               toPtr(true),
			longRunningOperations: toPtr(true),
			resourceModifications: toPtr(true),
			resourceDeletions:     toPtr(true),
			expectedOptionsCount:  5,
		},
		"WithDeferredCreationsSetToFalse_ReturnsError": {
			deferredCreations:      toPtr(false),
			expectedErrorSubstring: "no cleaning options specified",
		},
		"WithDeletesSetToFalse_ReturnsError": {
			deletes:                toPtr(false),
			expectedErrorSubstring: "no cleaning options specified",
		},
		"WithLongRunningOperationsSetToFalse_ReturnsError": {
			longRunningOperations:  toPtr(false),
			expectedErrorSubstring: "no cleaning options specified",
		},
		"WithResourceModificationsSetToFalse_ReturnsError": {
			resourceModifications:  toPtr(false),
			expectedErrorSubstring: "no cleaning options specified",
		},
		"WithResourceDeletionsSetToFalse_ReturnsError": {
			resourceDeletions:      toPtr(false),
			expectedErrorSubstring: "no cleaning options specified",
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)

			cmd := &CleanCommand{}
			cmd.Clean.DeferredCreations = c.deferredCreations
			cmd.Clean.Deletes = c.deletes
			cmd.Clean.Azure.LongRunningOperations = c.longRunningOperations
			cmd.Clean.Azure.ResourceModifications = c.resourceModifications
			cmd.Clean.Azure.ResourceDeletions = c.resourceDeletions

			options, err := cmd.buildOptions()

			if c.expectedErrorSubstring != "" {
				g.Expect(err).To(MatchError(ContainSubstring(c.expectedErrorSubstring)))
				g.Expect(options).To(BeNil())
			} else {
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(options).To(HaveLen(c.expectedOptionsCount))
			}
		})
	}
}
```

Note: `deferredCreations` field appears first in the struct (lifecycle ordering). The existing "WithBothOptionsSet" is renamed to "WithDeletesAndLongRunningOperations" for clarity. The "WithAllOptionsSet" count changes from 4 to 5.

- [ ] **Step 2: Run the updated test to verify all cases pass**

Run: `go test ./internal/cmd/ -run TestBuildOptions -v`
Expected: all 14 subtests PASS

- [ ] **Step 3: Run the full test suite to verify nothing is broken**

Run: `go test ./...`
Expected: all tests PASS

- [ ] **Step 4: Commit**

```bash
git add internal/cmd/clean_command_test.go
git commit -m "test: update buildOptions test cases for deferred creation"
```

---

### Task 6: Final verification

- [ ] **Step 1: Run linter**

Run: `task lint`
Expected: no lint errors

- [ ] **Step 2: Run full test suite**

Run: `task test`
Expected: all tests pass

- [ ] **Step 3: Build**

Run: `task build`
Expected: builds successfully
