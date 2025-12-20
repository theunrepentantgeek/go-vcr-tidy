package cleaner

import (
	"github.com/go-logr/logr"

	"github.com/theunrepentantgeek/go-vcr-tidy/internal/analyzer"
	"github.com/theunrepentantgeek/go-vcr-tidy/internal/interaction"
)

// fakeAnalyzer is a mock analyzer for testing purposes.
type fakeAnalyzer struct {
	name            string
	analyzeFunc     func(logr.Logger, interaction.Interface) (analyzer.Result, error)
	callCount       int
	lastInteraction interaction.Interface
}

func newFakeAnalyzer(name string) *fakeAnalyzer {
	return &fakeAnalyzer{
		name: name,
		analyzeFunc: func(logr.Logger, interaction.Interface) (analyzer.Result, error) {
			return analyzer.Result{}, nil
		},
	}
}

func (f *fakeAnalyzer) Analyze(log logr.Logger, inter interaction.Interface) (analyzer.Result, error) {
	f.callCount++
	f.lastInteraction = inter

	return f.analyzeFunc(log, inter)
}

func (f *fakeAnalyzer) withResult(result analyzer.Result) *fakeAnalyzer {
	f.analyzeFunc = func(logr.Logger, interaction.Interface) (analyzer.Result, error) {
		return result, nil
	}

	return f
}

func (f *fakeAnalyzer) withResults(results ...analyzer.Result) *fakeAnalyzer {
	index := 0
	f.analyzeFunc = func(logr.Logger, interaction.Interface) (analyzer.Result, error) {
		if index >= len(results) {
			return results[len(results)-1], nil
		}

		result := results[index]
		index++

		return result, nil
	}

	return f
}

func (f *fakeAnalyzer) withError(err error) *fakeAnalyzer {
	f.analyzeFunc = func(logr.Logger, interaction.Interface) (analyzer.Result, error) {
		return analyzer.Result{}, err
	}

	return f
}
