package cleaner

import (
	"errors"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/theunrepentantgeek/go-vcr-tidy/internal/analyzer"
	"github.com/theunrepentantgeek/go-vcr-tidy/internal/fake"
)

// Constructor Tests

func TestNewCleaner_WithNoAnalyzers_CreatesEmptyInstance(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	c := NewCleaner()

	g.Expect(c).ToNot(BeNil())
	g.Expect(c.analyzers).To(HaveLen(0))
}

func TestNewCleaner_WithAnalyzers_AddsAllToActiveSet(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	a1 := newFakeAnalyzer("analyzer1")
	a2 := newFakeAnalyzer("analyzer2")
	a3 := newFakeAnalyzer("analyzer3")

	c := NewCleaner(a1, a2, a3)

	g.Expect(c.analyzers).To(ContainElement(a1))
	g.Expect(c.analyzers).To(ContainElement(a2))
	g.Expect(c.analyzers).To(ContainElement(a3))
}

// Add Method Tests

func TestAdd_WithSingleAnalyzer_AddsToActiveSet(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	c := NewCleaner()
	a := newFakeAnalyzer("analyzer1")

	c.Add(a)

	g.Expect(c.analyzers).To(ContainElement(a))
}

func TestAdd_WithMultipleAnalyzers_AddsAllToActiveSet(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	c := NewCleaner()
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

	c := NewCleaner()
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

	a := newFakeAnalyzer("analyzer1")
	c := NewCleaner(a)

	baseURL := mustParseURL("https://api.example.com/resource/123")
	inter := fake.NewInteraction(baseURL, "GET", 200)
	err := c.analyze(inter)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(a.callCount).To(Equal(1))
	g.Expect(a.lastInteraction).To(Equal(inter))
}

func TestAnalyze_MultipleAnalyzers_AllProcessInteraction(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	a1 := newFakeAnalyzer("analyzer1")
	a2 := newFakeAnalyzer("analyzer2")
	a3 := newFakeAnalyzer("analyzer3")
	c := NewCleaner(a1, a2, a3)

	baseURL := mustParseURL("https://api.example.com/resource/123")
	inter := fake.NewInteraction(baseURL, "GET", 200)
	err := c.analyze(inter)

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

	expectedErr := errors.New("analysis failed")
	a := newFakeAnalyzer("analyzer1").withError(expectedErr)
	c := NewCleaner(a)

	baseURL := mustParseURL("https://api.example.com/resource/123")
	inter := fake.NewInteraction(baseURL, "GET", 200)
	err := c.analyze(inter)

	g.Expect(err).To(Equal(expectedErr))
}

func TestAnalyze_WhenAnalyzerFinishes_RemovesFromActiveSet(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	a := newFakeAnalyzer("analyzer1").withResult(analyzer.Finished())
	c := NewCleaner(a)

	baseURL := mustParseURL("https://api.example.com/resource/123")
	inter1 := fake.NewInteraction(baseURL, "GET", 200)
	err := c.analyze(inter1)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(a.callCount).To(Equal(1))

	g.Expect(c.analyzers).NotTo(ContainElement(a))

	// Second interaction should not be processed by the finished analyzer
	inter2 := fake.NewInteraction(baseURL, "GET", 200)
	err = c.analyze(inter2)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(a.callCount).To(Equal(1), "Finished analyzer should not process second interaction")
}

func TestAnalyze_AnalyzerSpawns_AddsNewAnalyzersToActiveSet(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	spawned := newFakeAnalyzer("spawned")
	a := newFakeAnalyzer("analyzer1").withResult(analyzer.Spawn(spawned))
	c := NewCleaner(a)

	baseURL := mustParseURL("https://api.example.com/resource/123")
	inter1 := fake.NewInteraction(baseURL, "GET", 200)
	err := c.analyze(inter1)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(c.analyzers).To(ContainElement(spawned))
}

func TestAnalyze_AnalyzerExcludesInteractions_TracksExclusions(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	baseURL := mustParseURL("https://api.example.com/resource/123")
	inter1 := fake.NewInteraction(baseURL, "GET", 200)
	inter2 := fake.NewInteraction(baseURL, "GET", 200)

	a := newFakeAnalyzer("analyzer1").
		withResult(analyzer.FinishedWithExclusions(inter1, inter2))
	c := NewCleaner(a)

	inter3 := fake.NewInteraction(baseURL, "DELETE", 200)
	err := c.analyze(inter3)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(c.interactionsToRemove).To(HaveKey(inter1.ID()))
	g.Expect(c.interactionsToRemove).To(HaveKey(inter2.ID()))
}

func TestAnalyze_AnalyzerFinishesAndSpawns_HandlesBoth(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	spawned := newFakeAnalyzer("spawned")
	result := analyzer.Result{
		Finished: true,
		Spawn:    []analyzer.Interface{spawned},
	}
	a := newFakeAnalyzer("analyzer1").withResult(result)
	c := NewCleaner(a)

	baseURL := mustParseURL("https://api.example.com/resource/123")
	inter1 := fake.NewInteraction(baseURL, "GET", 200)
	err := c.analyze(inter1)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(a.callCount).To(Equal(1))

	// Second interaction should only be processed by spawned analyzer
	inter2 := fake.NewInteraction(baseURL, "POST", 201)
	err = c.analyze(inter2)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(a.callCount).To(Equal(1), "Finished analyzer should not process second interaction")
	g.Expect(spawned.callCount).To(Equal(1), "Spawned analyzer should process second interaction")
}

func TestAnalyze_MultipleAnalyzersFinish_RemovesAll(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	a1 := newFakeAnalyzer("analyzer1").withResult(analyzer.Finished())
	a2 := newFakeAnalyzer("analyzer2").withResult(analyzer.Finished())
	a3 := newFakeAnalyzer("analyzer3").withResult(analyzer.Finished())
	c := NewCleaner(a1, a2, a3)

	baseURL := mustParseURL("https://api.example.com/resource/123")
	inter1 := fake.NewInteraction(baseURL, "GET", 200)

	g.Expect(c.analyze(inter1)).To(Succeed())

	g.Expect(c.analyzers).NotTo(ContainElement(a1))
	g.Expect(c.analyzers).NotTo(ContainElement(a2))
	g.Expect(c.analyzers).NotTo(ContainElement(a3))
}

func TestAnalyze_SpawnedAnalyzerProcessesNextInteraction_Works(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

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

	c := NewCleaner(a)

	baseURL := mustParseURL("https://api.example.com/resource/123")

	// First interaction: a spawns spawned1
	inter1 := fake.NewInteraction(baseURL, "GET", 200)
	err := c.analyze(inter1)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(c.analyzers).To(ContainElement(spawned1))

	// Second interaction: both a and spawned1 process, a finishes and spawns spawned2
	inter2 := fake.NewInteraction(baseURL, "POST", 201)
	err = c.analyze(inter2)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(c.analyzers).NotTo(ContainElement(a))
	g.Expect(c.analyzers).To(ContainElement(spawned1))
	g.Expect(c.analyzers).To(ContainElement(spawned2))
}

func TestAnalyze_ExclusionFromMultipleAnalyzers_AccumulatesAll(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	baseURL := mustParseURL("https://api.example.com/resource/123")
	inter1 := fake.NewInteraction(baseURL, "GET", 200)
	inter2 := fake.NewInteraction(baseURL, "GET", 201)
	inter3 := fake.NewInteraction(baseURL, "GET", 202)

	a1 := newFakeAnalyzer("analyzer1").withResult(analyzer.FinishedWithExclusions(inter1))
	a2 := newFakeAnalyzer("analyzer2").withResult(analyzer.FinishedWithExclusions(inter2))
	a3 := newFakeAnalyzer("analyzer3").withResult(analyzer.FinishedWithExclusions(inter3))
	c := NewCleaner(a1, a2, a3)

	inter4 := fake.NewInteraction(baseURL, "DELETE", 200)
	err := c.analyze(inter4)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(c.interactionsToRemove).To(HaveKey(inter1.ID()))
	g.Expect(c.interactionsToRemove).To(HaveKey(inter2.ID()))
	g.Expect(c.interactionsToRemove).To(HaveKey(inter3.ID()))
}

func TestAnalyze_EmptyResult_NoSideEffects(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	a := newFakeAnalyzer("analyzer1").withResult(analyzer.Result{})
	c := NewCleaner(a)

	baseURL := mustParseURL("https://api.example.com/resource/123")
	inter1 := fake.NewInteraction(baseURL, "GET", 200)
	err := c.analyze(inter1)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(a.callCount).To(Equal(1))

	// Second interaction should still be processed
	inter2 := fake.NewInteraction(baseURL, "GET", 200)
	err = c.analyze(inter2)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(a.callCount).To(Equal(2))
}

// Edge Cases

func TestAnalyze_NoAnalyzers_NoError(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	c := NewCleaner()

	baseURL := mustParseURL("https://api.example.com/resource/123")
	inter := fake.NewInteraction(baseURL, "GET", 200)
	err := c.analyze(inter)

	g.Expect(err).ToNot(HaveOccurred())
}

func TestAnalyze_AllAnalyzersFinish_LeavesEmptySet(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	a1 := newFakeAnalyzer("analyzer1").withResult(analyzer.Finished())
	a2 := newFakeAnalyzer("analyzer2").withResult(analyzer.Finished())
	c := NewCleaner(a1, a2)

	baseURL := mustParseURL("https://api.example.com/resource/123")
	inter1 := fake.NewInteraction(baseURL, "GET", 200)
	err := c.analyze(inter1)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(a1.callCount).To(Equal(1))
	g.Expect(a2.callCount).To(Equal(1))

	// After all analyzers finish, subsequent interactions should work but do nothing
	inter2 := fake.NewInteraction(baseURL, "GET", 200)
	err = c.analyze(inter2)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(a1.callCount).To(Equal(1), "Finished analyzer should not be called again")
	g.Expect(a2.callCount).To(Equal(1), "Finished analyzer should not be called again")
}
