# Deferred Creation Polling — Detector/Monitor Design

## Summary

Add a new generic detector/monitor pair (`DetectDeferredCreation` / `MonitorDeferredCreation`) to handle the case where a client polls a known URL receiving 404s while waiting for a resource to be created by another process. The monitor removes intermediate 404 GET responses, keeping only the first and last to preserve timing context.

## Motivation

In recorded HTTP interactions (cassettes), a common pattern is a client polling a URL with GET requests, receiving repeated 404 Not Found responses until the resource is eventually created and a 2xx response is returned. These intermediate 404s add bulk to cassette files without adding test value. Removing them reduces cassette size while preserving the meaningful boundary interactions.

## Design

### DetectDeferredCreation

**File:** `internal/generic/detect_deferred_creation.go`

A stateless detector that watches every interaction for a GET returning 404. When found, it spawns a `MonitorDeferredCreation` for that URL.

- Implements `analyzer.Interface`
- Never finishes (runs for the entire analysis lifetime)
- Ignores all interactions that aren't GET→404
- On GET→404: logs at Debug level, spawns `MonitorDeferredCreation(interaction)`, passing the triggering interaction so the monitor starts with it already accumulated

### MonitorDeferredCreation

**File:** `internal/generic/monitor_deferred_creation.go`

A stateful monitor that tracks a specific URL after a GET→404 is detected. It accumulates 404 GET responses and waits for either a successful GET (2xx) or an abandonment condition.

**Constructor:** `NewMonitorDeferredCreation(firstInteraction interaction.Interface)` — extracts `baseURL` from `firstInteraction.Request().BaseURL()` and seeds the accumulator with the triggering interaction.

**State:**

- `baseURL *url.URL` — derived from the first interaction
- `interactions []interaction.Interface` — seeded with the first 404 GET

**Behavior on `Analyze()`:**

| Condition | Action |
|---|---|
| Different URL | Ignore (return empty result) |
| GET → 404 | Accumulate the interaction |
| GET → 2xx | Creation confirmed — finish with exclusions |
| POST, PUT, DELETE, PATCH to URL | Abandon — `Finished()` with no exclusions |
| GET → other status (500, 301, etc.) | Abandon — `Finished()` with no exclusions |

**Exclusion logic** (on creation confirmed via GET→2xx):

- If fewer than 3 accumulated 404 GETs: finish with no exclusions (nothing worth removing)
- If 3+ accumulated: exclude `interactions[1 : len-1]` (keep first and last 404, remove all middle ones)

This mirrors the `MonitorDeletion.deletionConfirmed()` logic exactly.

### Testing

**Files:** `detect_deferred_creation_test.go` and `monitor_deferred_creation_test.go` in `internal/generic/`

Tests follow the same patterns as the deletion pair, using the existing `runAnalyzer` helper from `helpers_test.go` and fakes from `internal/fake/`.

**Detector tests:**

- GET→404 spawns a `MonitorDeferredCreation`
- Various 4xx codes other than 404 do not spawn
- Non-GET methods returning 404 do not spawn
- Successful GETs (2xx) do not spawn
- Multiple 404 GETs to different URLs spawn independent monitors
- Detector never finishes

**Monitor tests:**

- Single 404 then 2xx → finishes, no exclusions (only 1 accumulated)
- Two 404s then 2xx → finishes, no exclusions (only 2 accumulated)
- Three 404s then 2xx → finishes, middle one excluded
- Many 404s then 2xx → all middle excluded
- Different URL → ignored
- POST/PUT/DELETE/PATCH to URL → abandons
- GET returning 500/301 → abandons
- Various 2xx status codes accepted as creation confirmation
- URL with query parameters matches base URL

## Decisions

1. **Trigger:** Any GET returning 404 (simplest, consistent with deletion detector pattern; false triggers are harmless since monitors that see no polling produce no exclusions)
2. **Abandonment:** Non-GET methods or unexpected status codes cause the monitor to finish with no exclusions (mirrors deletion monitor)
3. **Exclusion strategy:** Keep first and last 404 GET, remove middle ones (preserves timing context via timestamps)
4. **Naming:** `DetectDeferredCreation` / `MonitorDeferredCreation` — "deferred" captures the intent that creation happens elsewhere
5. **Constructor:** Takes only `firstInteraction` (derives baseURL from it, avoiding redundant parameter)

## Approach

Direct mirror of the existing `DetectDeletion`/`MonitorDeletion` pair. No new abstractions. If a third pattern emerges later, that would be the right time to consider extracting shared logic.

## Files Changed

- `internal/generic/detect_deferred_creation.go` — new detector
- `internal/generic/detect_deferred_creation_test.go` — detector tests
- `internal/generic/monitor_deferred_creation.go` — new monitor
- `internal/generic/monitor_deferred_creation_test.go` — monitor tests
