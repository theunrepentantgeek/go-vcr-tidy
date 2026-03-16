# Integrate DetectDeferredCreation Analyzer

## Summary

Wire the existing `DetectDeferredCreation` and `MonitorDeferredCreation` analyzers (in `internal/generic/`) into the application's integration chain, following the established pattern used by `DetectDeletion`.

## Decision Record

- **CLI flag placement**: Top-level on `CleaningOptions` (not nested under Azure), since the analyzer is in `internal/generic/`.
- **Naming**: `--clean-deferred-creations` flag, `ShouldCleanDeferredCreations()` method, `ReduceDeferredCreationMonitoring()` public API function.
- **`--clean-all` behavior**: Includes deferred creation cleaning.
- **Lifecycle ordering**: Creation options appear before deletion options throughout (struct fields, `Options()` method, public API file, tests).

## Changes

### 1. Public API: `pkg/vcrcleaner/options.go`

Add `ReduceDeferredCreationMonitoring()` before `ReduceDeleteMonitoring()`:

```go
// ReduceDeferredCreationMonitoring adds an analyzer that reduces deferred creation monitoring noise.
func ReduceDeferredCreationMonitoring() Option {
    return func(c *cleaner.Cleaner) {
        c.AddAnalyzers(generic.NewDetectDeferredCreation())
    }
}
```

### 2. CLI Options: `internal/cmd/cleaning_options.go`

Add `DeferredCreations *bool` field before `Deletes` in `CleaningOptions`:

```go
type CleaningOptions struct {
    All               *bool `help:"Clean all supported interaction types."`
    DeferredCreations *bool `help:"Clean deferred creation interactions."`
    Deletes           *bool `help:"Clean delete interactions."`
    Azure AzureCleaningOptions `embed:"" prefix:"azure-"`
}
```

Add `ShouldCleanDeferredCreations()` method:

```go
func (opt *CleaningOptions) ShouldCleanDeferredCreations() bool {
    return opt.coalesce(opt.DeferredCreations, opt.All)
}
```

Wire into `Options()`, before deletes:

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

### 3. Tests: `internal/cmd/cleaning_options_test.go`

Update `TestCleaningOptions_Options`:
- Add `deferredCreations *bool` field to test struct.
- Add cases: deferred creations alone (count 1), combined with deletes (count 2), combined with azure-all (count 5), all three (count 6).
- Update existing combined case counts where `All` is involved.

Add `TestCleaningOptions_ShouldCleanDeferredCreations` table test mirroring `ShouldCleanDeletes` tests:
- True → true, false → false, nil → false.

### 4. Tests: `internal/cmd/clean_command_test.go`

Update `TestBuildOptions`:
- Add `deferredCreations *bool` field to test struct.
- Add cases: deferred creations alone (count 1), deferred creations false (error), combined cases.
- Update "all options" count from 4 to 5.
- Wire `c.deferredCreations` → `cmd.Clean.DeferredCreations` in test body.

### What's NOT changing

- `internal/generic/` — analyzers and their tests are already complete.
- `internal/cleaner/` — the generic analyzer pipeline already handles any analyzer.
- `pkg/vcrcleaner/cleaner.go` / `cleaner_test.go` — no new golden test cassettes needed.
- `AzureCleaningOptions` — unchanged.
- `main.go` — unchanged (CLI struct embedding handles the new flag automatically).
