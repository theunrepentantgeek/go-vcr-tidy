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
		c.Add(generic.NewDetectDeletion())
	}
}

func ReduceAzureLongRunningOperationPolling() Option {
	return func(c *cleaner.Cleaner) {
		c.Add(azure.NewDetectAzureLongRunningOperation())
	}
}
