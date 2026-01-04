package cmd

import "github.com/theunrepentantgeek/go-vcr-tidy/pkg/vcrcleaner"

type AzureCleaningOptions struct {
	All                    *bool `help:"Clean all Azure-related monitoring interactions."`
	AsynchronousOperations *bool `help:"Clean Azure asynchronous operation monitoring interactions."`
	LongRunningOperations  *bool `help:"Clean Azure long-running operation interactions."`
	ResourceModifications  *bool `help:"Clean Azure resource modification (PUT/PATCH) monitoring interactions."`
	ResourceDeletions      *bool `help:"Clean Azure resource deletion monitoring interactions."`
}

func (opt *AzureCleaningOptions) Options() []vcrcleaner.Option {
	var result []vcrcleaner.Option

	if opt.ShouldCleanLongRunningOperations() {
		result = append(result, vcrcleaner.ReduceAzureLongRunningOperationPolling())
	}

	if opt.ShouldCleanResourceModifications() {
		result = append(result, vcrcleaner.ReduceAzureResourceModificationMonitoring())
	}

	if opt.ShouldCleanResourceDeletions() {
		result = append(result, vcrcleaner.ReduceAzureResourceDeletionMonitoring())
	}

	if opt.ShouldCleanAsynchronousOperations() {
		result = append(result, vcrcleaner.ReduceAzureAsynchronousOperationMonitoring())
	}

	return result
}

// ShouldCleanLongRunningOperations indicates whether long-running operation monitoring should be cleaned.
// More specific options override the general 'All' option.
func (opt *AzureCleaningOptions) ShouldCleanLongRunningOperations() bool {
	if opt.LongRunningOperations != nil {
		return *opt.LongRunningOperations
	}

	if opt.All != nil {
		return *opt.All
	}

	return false
}

// ShouldCleanResourceModifications indicates whether resource modification monitoring should be cleaned.
// More specific options override the general 'All' option.
func (opt *AzureCleaningOptions) ShouldCleanResourceModifications() bool {
	if opt.ResourceModifications != nil {
		return *opt.ResourceModifications
	}

	if opt.All != nil {
		return *opt.All
	}

	return false
}

// ShouldCleanResourceDeletions indicates whether resource deletion monitoring should be cleaned.
// More specific options override the general 'All' option.
func (opt *AzureCleaningOptions) ShouldCleanResourceDeletions() bool {
	if opt.ResourceDeletions != nil {
		return *opt.ResourceDeletions
	}

	if opt.All != nil {
		return *opt.All
	}

	return false
}

func (opt *AzureCleaningOptions) ShouldCleanAsynchronousOperations() bool {
	if opt.AsynchronousOperations != nil {
		return *opt.AsynchronousOperations
	}

	if opt.All != nil {
		return *opt.All
	}

	return false
}
