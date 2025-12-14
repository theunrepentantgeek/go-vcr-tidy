# Development guidelines for go-vcr-tidy

# Abstractions

Key abstractions - including `Analyzer` and `Interaction` are declared in their own packages. 

## Testing

We use [gomega](https://github.com/onsi/gomega) for unit test assertions, and [goldie](github.com/sebdah/goldie/v2) for 
golden tests, where required.

Test cases are ordered in each test file, with later tests able to assume that system properties asserted by earlier 
tests are held. This helps to narrow the focus of each test. As a direct corolloary of this, when diagnosing test 
failures, the earliest failing test in a file is a good place to start.

Table tests use `cases := map[string]struct{...}` to capture test cases, with the name of the test as the map key. Test 
case iteration uses `for name, c := range cases`.

* All tests are marked with `t.Parallel()` unless the test cannot run in parallel.
* Helper methods are always marked with `t.Helper()`.
