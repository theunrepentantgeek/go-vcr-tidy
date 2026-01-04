package vcrcleaner

import (
	"github.com/theunrepentantgeek/go-vcr-tidy/internal/azure"
	"github.com/theunrepentantgeek/go-vcr-tidy/internal/cleaner"
	"github.com/theunrepentantgeek/go-vcr-tidy/internal/generic"
)

// Option represents a configuration option for the Cleaner.
type Option func(*cleaner.Cleaner)

// ReduceDeleteMonitoring adds an analyzer that reduces delete monitoring noise.
func ReduceDeleteMonitoring() Option {
	return func(c *cleaner.Cleaner) {
		c.AddAnalyzers(generic.NewDetectDeletion())
	}
}

func ReduceAzureLongRunningOperationPolling() Option {
	return func(c *cleaner.Cleaner) {
		c.AddAnalyzers(azure.NewDetectAzureLongRunningOperation())
	}
}

func ReduceAzureAsynchronousOperationPolling() Option {
	return func(c *cleaner.Cleaner) {
		c.AddAnalyzers(azure.NewDetectAzureAsynchronousOperation())
	}
}

// ReduceAzureResourceModificationMonitoring adds an analyzer that reduces Azure resource modification monitoring.
// This analyzer watches for PUT and PATCH requests and monitors subsequent GET requests for Creating/Updating states.
func ReduceAzureResourceModificationMonitoring() Option {
	return func(c *cleaner.Cleaner) {
		c.AddAnalyzers(azure.NewDetectResourceModification())
	}
}

// ReduceAzureResourceDeletionMonitoring adds an analyzer that reduces Azure resource deletion monitoring.
// This analyzer watches for DELETE requests and monitors subsequent GET requests for Deleting state.
func ReduceAzureResourceDeletionMonitoring() Option {
	return func(c *cleaner.Cleaner) {
		c.AddAnalyzers(azure.NewDetectResourceDeletion())
	}
}

// ReduceAzureAsynchronousOperationMonitoring adds an analyzer that reduces Azure asynchronous operation monitoring.
// This analyzer watches for asynchronous operation polling interactions.
func ReduceAzureAsynchronousOperationMonitoring() Option {
	return func(c *cleaner.Cleaner) {
		c.AddAnalyzers(azure.NewDetectAzureAsynchronousOperation())
	}
}
