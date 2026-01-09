package fake

import (
	"log/slog"

	"github.com/theunrepentantgeek/go-vcr-tidy/internal/analyzer"
	"github.com/theunrepentantgeek/go-vcr-tidy/internal/interaction"
)

// TestAnalyzer is a mock analyzer for testing purposes.
type TestAnalyzer struct {
	name            string
	analyzeFunc     func(*slog.Logger, interaction.Interface) (analyzer.Result, error)
	CallCount       int
	LastInteraction interaction.Interface
}

// Analyzer creates a new TestAnalyzer with the given name.
func Analyzer(name string) *TestAnalyzer {
	return &TestAnalyzer{
		name: name,
		analyzeFunc: func(*slog.Logger, interaction.Interface) (analyzer.Result, error) {
			return analyzer.Result{}, nil
		},
	}
}

// Analyze processes an interaction and tracks call count and last interaction.
func (f *TestAnalyzer) Analyze(log *slog.Logger, inter interaction.Interface) (analyzer.Result, error) {
	f.CallCount++
	f.LastInteraction = inter

	return f.analyzeFunc(log, inter)
}

// WithResult configures the analyzer to return the specified result.
func (f *TestAnalyzer) WithResult(result analyzer.Result) *TestAnalyzer {
	f.analyzeFunc = func(*slog.Logger, interaction.Interface) (analyzer.Result, error) {
		return result, nil
	}

	return f
}

// WithResults configures the analyzer to return a sequence of results.
func (f *TestAnalyzer) WithResults(results ...analyzer.Result) *TestAnalyzer {
	index := 0
	f.analyzeFunc = func(*slog.Logger, interaction.Interface) (analyzer.Result, error) {
		if index >= len(results) {
			return results[len(results)-1], nil
		}

		result := results[index]
		index++

		return result, nil
	}

	return f
}

// WithError configures the analyzer to return an error.
func (f *TestAnalyzer) WithError(err error) *TestAnalyzer {
	f.analyzeFunc = func(*slog.Logger, interaction.Interface) (analyzer.Result, error) {
		return analyzer.Result{}, err
	}

	return f
}
