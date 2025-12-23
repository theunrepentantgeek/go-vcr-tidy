# Copilot Instructions for go-vcr-tidy

## Project Overview

go-vcr-tidy provides extensions for [go-vcr](https://github.com/dnaeon/go-vcr) to reduce the size of cassette files used in testing. The library focuses on tools that elide selected HTTP interactions from recordings, particularly for Azure services but with potential for other patterns.

Key planned tools include:
- Monitoring for resource creation/deletion in Azure
- Azure long-running operations (LRO) optimization
- Reducing polling sequences from recordings

## Project Structure

- `/internal/analyzer` - Core abstractions including `Analyzer` and `Interaction` interfaces
- `/internal/azure` - Azure-specific analyzers and cleaning tools
- `/internal/generic` - Generic analyzers for common patterns
- `/internal/cmd` - CLI command implementations
- `/pkg` - Public-facing packages
- `/tools` - Development tools and utilities

## Coding Standards

### General Go Guidelines

- Follow standard Go conventions and idioms
- Use `gofumpt` for code formatting
- Run `golangci-lint` to catch linting issues
- Keep code simple and readable

### Testing Practices

We use [gomega](https://github.com/onsi/gomega) for unit test assertions and [goldie](https://github.com/sebdah/goldie/v2) for golden tests.

**Test Structure:**
- Tests are ordered within each file, with later tests able to assume properties asserted by earlier tests
- When diagnosing failures, start with the earliest failing test in a file
- Use table tests with `cases := map[string]struct{...}` to capture test cases
- Test case iteration uses `for name, c := range cases`

**Test Markers:**
- Mark all tests with `t.Parallel()` unless the test cannot run in parallel
- Mark helper methods with `t.Helper()`

**Test Packages:**
- Only use a test package suffix (e.g., `package foo_test`) if needed to avoid circular imports
- Otherwise, keep tests in the same package as the code they test

**Error Assertions:**
- **ALWAYS** use `MatchError()` with a nested matcher for error checks
- **NEVER** assert simply `err != nil` - this allows ANY error to pass
- Prefer `MatchError(ContainSubstring("expected text"))` for flexible error matching
- Avoid brittle `err.Error() == "exact string"` patterns

**Test Helpers:**
- Use generic helper functions where possible (e.g., `toPtr[T any](v T) *T`)
- Create helper functions to reduce test duplication
- Store reusable test data in `testdata/` directories rather than inline strings

**Test Data:**
- Use `testdata/` directories for sample files and golden test data
- Avoid inline strings for test fixtures - extract to files
- Helper functions should load from `testdata/` when needed

**Code Style in Tests:**
- Omit default values (nil, false, zero) from test case structs for clarity
- Use struct literals with helper functions like `toPtr()` instead of separate variable declarations
- Infer boolean flags from other fields when possible (e.g., `expectError` from `expectedErrorSubstring != ""`)

## Build and Test

### Building

```bash
# Build the binary
task build

# Build with SBOM generation
task build.sbom
```

### Testing

```bash
# Run all tests
task test

# Run only unit tests
task unit-test

# Run linter
task lint

# Run all CI tasks
task ci
```

### Code Formatting and Cleanup

```bash
# Tidy everything
task tidy

# Run gofumpt
task tidy.gofumpt

# Run go mod tidy
task tidy.mod

# Run golangci-lint in fix mode
task tidy.lint
```

### Mutation Testing

```bash
# Check test quality with gremlins
task gremlins
```

## Key Abstractions

### Analyzer

The `Analyzer` interface is the core abstraction for analyzing cassette recordings. Analyzers examine sequences of HTTP interactions and identify patterns that can be optimized.

### Interaction

Represents a single HTTP request/response pair in a cassette recording. This abstraction allows analyzers to inspect and make decisions about which interactions to keep or remove.

## Development Workflow

1. **Before making changes:**
   - Review existing code structure and patterns
   - Check `DEVELOPMENT.md` for project-specific guidelines
   - Run existing tests to establish a baseline: `task test`

2. **While developing:**
   - Write tests following the conventions above
   - Use table-driven tests for multiple similar scenarios
   - Keep commits focused and atomic
   - Run `task tidy` to format code before committing

3. **Before submitting:**
   - Ensure all tests pass: `task test`
   - Run the linter: `task lint`
   - Verify builds succeed: `task build`
   - Consider running mutation tests: `task gremlins`

## Pull Request Guidelines

- Reference related issues in PR descriptions
- Keep PRs focused on a single concern
- Ensure CI passes before requesting review
- Update documentation if adding new features or changing behavior
- Add tests for new functionality

## Common Patterns

### Creating Generic Helpers

Instead of type-specific helpers like `boolPtr()`, create generic versions:

```go
func toPtr[T any](v T) *T {
    return &v
}
```

### Table-Driven Tests

```go
func TestFeature(t *testing.T) {
    t.Parallel()
    
    cases := map[string]struct {
        input          string
        expectedOutput string
        expectedError  string  // Use for error substring matching
    }{
        "DescriptiveTestName": {
            input:          "test input",
            expectedOutput: "expected output",
        },
        "ErrorCase": {
            input:         "bad input",
            expectedError: "substring of expected error",
        },
    }
    
    for name, c := range cases {
        t.Run(name, func(t *testing.T) {
            t.Parallel()
            g := NewWithT(t)
            
            result, err := FeatureUnderTest(c.input)
            
            if c.expectedError != "" {
                g.Expect(err).To(MatchError(ContainSubstring(c.expectedError)))
            } else {
                g.Expect(err).ToNot(HaveOccurred())
                g.Expect(result).To(Equal(c.expectedOutput))
            }
        })
    }
}
```

## Notes

- The initial focus is on Azure patterns, but the architecture supports adding analyzers for other cloud providers or HTTP interaction patterns
- Tests should be thorough but not brittle - use substring matching for errors rather than exact string comparisons
- When in doubt, favor simplicity and readability over clever optimizations
