package cleaner

import (
	"errors"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/go-logr/logr/testr"

	"github.com/theunrepentantgeek/go-vcr-tidy/internal/analyzer"
	"github.com/theunrepentantgeek/go-vcr-tidy/internal/fake"
)

// Constructor Tests

func TestNewCleaner_WithNoAnalyzers_CreatesEmptyInstance(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	c := New()

	g.Expect(c).ToNot(BeNil())
	g.Expect(c.analyzers).To(BeEmpty())
}

func TestNewCleaner_WithAnalyzers_AddsAllToActiveSet(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	a1 := newFakeAnalyzer("analyzer1")
	a2 := newFakeAnalyzer("analyzer2")
	a3 := newFakeAnalyzer("analyzer3")

	c := New(a1, a2, a3)

	g.Expect(c.analyzers).To(ContainElement(a1))
	g.Expect(c.analyzers).To(ContainElement(a2))
	g.Expect(c.analyzers).To(ContainElement(a3))
}

// Add Method Tests

func TestAdd_WithSingleAnalyzer_AddsToActiveSet(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	c := New()
	a := newFakeAnalyzer("analyzer1")

	c.Add(a)

	g.Expect(c.analyzers).To(ContainElement(a))
}

func TestAdd_WithMultipleAnalyzers_AddsAllToActiveSet(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	c := New()
	a1 := newFakeAnalyzer("analyzer1")
	a2 := newFakeAnalyzer("analyzer2")
	a3 := newFakeAnalyzer("analyzer3")

	c.Add(a1, a2, a3)

	g.Expect(c.analyzers).To(ContainElement(a1))
	g.Expect(c.analyzers).To(ContainElement(a2))
	g.Expect(c.analyzers).To(ContainElement(a3))
}

func TestAdd_WhenCalledMultipleTimes_AccumulatesAnalyzers(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	c := New()
	a1 := newFakeAnalyzer("analyzer1")
	a2 := newFakeAnalyzer("analyzer2")
	a3 := newFakeAnalyzer("analyzer3")

	c.Add(a1)
	c.Add(a2)
	c.Add(a3)

	g.Expect(c.analyzers).To(ContainElement(a1))
	g.Expect(c.analyzers).To(ContainElement(a2))
	g.Expect(c.analyzers).To(ContainElement(a3))
}

// Analyze Method Tests

func TestAnalyze_SingleAnalyzer_ProcessesInteraction(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)
	log := testr.NewWithOptions(t, testr.Options{Verbosity: 1})

	a := newFakeAnalyzer("analyzer1")
	c := New(a)

	baseURL := mustParseURL("https://api.example.com/resource/123")
	inter := fake.Interaction(baseURL, "GET", 200)
	err := c.Analyze(log, inter)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(a.callCount).To(Equal(1))
	g.Expect(a.lastInteraction).To(Equal(inter))
}

func TestAnalyze_MultipleAnalyzers_AllProcessInteraction(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)
	log := testr.NewWithOptions(t, testr.Options{Verbosity: 1})

	a1 := newFakeAnalyzer("analyzer1")
	a2 := newFakeAnalyzer("analyzer2")
	a3 := newFakeAnalyzer("analyzer3")
	c := New(a1, a2, a3)

	baseURL := mustParseURL("https://api.example.com/resource/123")
	inter := fake.Interaction(baseURL, "GET", 200)
	err := c.Analyze(log, inter)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(a1.callCount).To(Equal(1))
	g.Expect(a2.callCount).To(Equal(1))
	g.Expect(a3.callCount).To(Equal(1))
	g.Expect(a1.lastInteraction).To(Equal(inter))
	g.Expect(a2.lastInteraction).To(Equal(inter))
	g.Expect(a3.lastInteraction).To(Equal(inter))
}

func TestAnalyze_WhenAnalyzerReturnsError_PropagatesError(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)
	log := testr.NewWithOptions(t, testr.Options{Verbosity: 1})

	expectedErr := errors.New("analysis failed")
	a := newFakeAnalyzer("analyzer1").withError(expectedErr)
	c := New(a)

	baseURL := mustParseURL("https://api.example.com/resource/123")
	inter := fake.Interaction(baseURL, "GET", 200)
	err := c.Analyze(log, inter)

	g.Expect(err).To(MatchError(ContainSubstring(expectedErr.Error())))
}

func TestAnalyze_WhenAnalyzerFinishes_RemovesFromActiveSet(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)
	log := testr.NewWithOptions(t, testr.Options{Verbosity: 1})

	a := newFakeAnalyzer("analyzer1").withResult(analyzer.Finished())
	c := New(a)

	baseURL := mustParseURL("https://api.example.com/resource/123")
	inter1 := fake.Interaction(baseURL, "GET", 200)
	err := c.Analyze(log, inter1)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(a.callCount).To(Equal(1))

	g.Expect(c.analyzers).NotTo(ContainElement(a))

	// Second interaction should not be processed by the finished analyzer
	inter2 := fake.Interaction(baseURL, "GET", 200)
	err = c.Analyze(log, inter2)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(a.callCount).To(Equal(1), "Finished analyzer should not process second interaction")
}

func TestAnalyze_AnalyzerSpawns_AddsNewAnalyzersToActiveSet(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)
	log := testr.NewWithOptions(t, testr.Options{Verbosity: 1})

	spawned := newFakeAnalyzer("spawned")
	a := newFakeAnalyzer("analyzer1").withResult(analyzer.Spawn(spawned))
	c := New(a)

	baseURL := mustParseURL("https://api.example.com/resource/123")
	inter1 := fake.Interaction(baseURL, "GET", 200)
	err := c.Analyze(log, inter1)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(c.analyzers).To(ContainElement(spawned))
}

func TestAnalyze_AnalyzerExcludesInteractions_TracksExclusions(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)
	log := testr.NewWithOptions(t, testr.Options{Verbosity: 1})

	baseURL := mustParseURL("https://api.example.com/resource/123")
	inter1 := fake.Interaction(baseURL, "GET", 200)
	inter2 := fake.Interaction(baseURL, "GET", 200)

	a := newFakeAnalyzer("analyzer1").
		withResult(analyzer.FinishedWithExclusions(inter1, inter2))
	c := New(a)

	inter3 := fake.Interaction(baseURL, "DELETE", 200)
	err := c.Analyze(log, inter3)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(c.interactionsToRemove).To(HaveKey(inter1.ID()))
	g.Expect(c.interactionsToRemove).To(HaveKey(inter2.ID()))
}

func TestAnalyze_AnalyzerFinishesAndSpawns_HandlesBoth(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)
	log := testr.NewWithOptions(t, testr.Options{Verbosity: 1})

	spawned := newFakeAnalyzer("spawned")
	result := analyzer.Result{
		Finished: true,
		Spawn:    []analyzer.Interface{spawned},
	}
	a := newFakeAnalyzer("analyzer1").withResult(result)
	c := New(a)

	baseURL := mustParseURL("https://api.example.com/resource/123")
	inter1 := fake.Interaction(baseURL, "GET", 200)
	err := c.Analyze(log, inter1)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(a.callCount).To(Equal(1))

	// Second interaction should only be processed by spawned analyzer
	inter2 := fake.Interaction(baseURL, "POST", 201)
	err = c.Analyze(log, inter2)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(a.callCount).To(Equal(1), "Finished analyzer should not process second interaction")
	g.Expect(spawned.callCount).To(Equal(1), "Spawned analyzer should process second interaction")
}

func TestAnalyze_MultipleAnalyzersFinish_RemovesAll(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)
	log := testr.NewWithOptions(t, testr.Options{Verbosity: 1})

	a1 := newFakeAnalyzer("analyzer1").withResult(analyzer.Finished())
	a2 := newFakeAnalyzer("analyzer2").withResult(analyzer.Finished())
	a3 := newFakeAnalyzer("analyzer3").withResult(analyzer.Finished())
	c := New(a1, a2, a3)

	baseURL := mustParseURL("https://api.example.com/resource/123")
	inter1 := fake.Interaction(baseURL, "GET", 200)

	g.Expect(c.Analyze(log, inter1)).To(Succeed())

	g.Expect(c.analyzers).NotTo(ContainElement(a1))
	g.Expect(c.analyzers).NotTo(ContainElement(a2))
	g.Expect(c.analyzers).NotTo(ContainElement(a3))
}

func TestAnalyze_SpawnedAnalyzerProcessesNextInteraction_Works(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)
	log := testr.NewWithOptions(t, testr.Options{Verbosity: 1})

	spawned1 := newFakeAnalyzer("spawned1")
	spawned2 := newFakeAnalyzer("spawned2")

	// First analyzer spawns on first interaction, finishes on second
	a := newFakeAnalyzer("analyzer1").withResults(
		analyzer.Spawn(spawned1),
		analyzer.Result{
			Finished: true,
			Spawn:    []analyzer.Interface{spawned2},
		},
	)

	c := New(a)

	baseURL := mustParseURL("https://api.example.com/resource/123")

	// First interaction: a spawns spawned1
	inter1 := fake.Interaction(baseURL, "GET", 200)
	err := c.Analyze(log, inter1)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(c.analyzers).To(ContainElement(spawned1))

	// Second interaction: both a and spawned1 process, a finishes and spawns spawned2
	inter2 := fake.Interaction(baseURL, "POST", 201)
	err = c.Analyze(log, inter2)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(c.analyzers).NotTo(ContainElement(a))
	g.Expect(c.analyzers).To(ContainElement(spawned1))
	g.Expect(c.analyzers).To(ContainElement(spawned2))
}

func TestAnalyze_ExclusionFromMultipleAnalyzers_AccumulatesAll(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)
	log := testr.NewWithOptions(t, testr.Options{Verbosity: 1})

	baseURL := mustParseURL("https://api.example.com/resource/123")
	inter1 := fake.Interaction(baseURL, "GET", 200)
	inter2 := fake.Interaction(baseURL, "GET", 201)
	inter3 := fake.Interaction(baseURL, "GET", 202)

	a1 := newFakeAnalyzer("analyzer1").withResult(analyzer.FinishedWithExclusions(inter1))
	a2 := newFakeAnalyzer("analyzer2").withResult(analyzer.FinishedWithExclusions(inter2))
	a3 := newFakeAnalyzer("analyzer3").withResult(analyzer.FinishedWithExclusions(inter3))
	c := New(a1, a2, a3)

	inter4 := fake.Interaction(baseURL, "DELETE", 200)
	err := c.Analyze(log, inter4)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(c.interactionsToRemove).To(HaveKey(inter1.ID()))
	g.Expect(c.interactionsToRemove).To(HaveKey(inter2.ID()))
	g.Expect(c.interactionsToRemove).To(HaveKey(inter3.ID()))
}

func TestAnalyze_EmptyResult_NoSideEffects(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)
	log := testr.NewWithOptions(t, testr.Options{Verbosity: 1})

	a := newFakeAnalyzer("analyzer1").withResult(analyzer.Result{})
	c := New(a)

	baseURL := mustParseURL("https://api.example.com/resource/123")
	inter1 := fake.Interaction(baseURL, "GET", 200)
	err := c.Analyze(log, inter1)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(a.callCount).To(Equal(1))

	// Second interaction should still be processed
	inter2 := fake.Interaction(baseURL, "GET", 200)
	err = c.Analyze(log, inter2)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(a.callCount).To(Equal(2))
}

// Edge Cases

func TestAnalyze_NoAnalyzers_NoError(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)
	log := testr.NewWithOptions(t, testr.Options{Verbosity: 1})

	c := New()

	baseURL := mustParseURL("https://api.example.com/resource/123")
	inter := fake.Interaction(baseURL, "GET", 200)
	err := c.Analyze(log, inter)

	g.Expect(err).ToNot(HaveOccurred())
}

func TestAnalyze_AllAnalyzersFinish_LeavesEmptySet(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)
	log := testr.NewWithOptions(t, testr.Options{Verbosity: 1})

	a1 := newFakeAnalyzer("analyzer1").withResult(analyzer.Finished())
	a2 := newFakeAnalyzer("analyzer2").withResult(analyzer.Finished())
	c := New(a1, a2)

	baseURL := mustParseURL("https://api.example.com/resource/123")
	inter1 := fake.Interaction(baseURL, "GET", 200)
	err := c.Analyze(log, inter1)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(a1.callCount).To(Equal(1))
	g.Expect(a2.callCount).To(Equal(1))

	// After all analyzers finish, subsequent interactions should work but do nothing
	inter2 := fake.Interaction(baseURL, "GET", 200)
	err = c.Analyze(log, inter2)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(a1.callCount).To(Equal(1), "Finished analyzer should not be called again")
	g.Expect(a2.callCount).To(Equal(1), "Finished analyzer should not be called again")
}
