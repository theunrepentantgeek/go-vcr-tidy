package cmd

import "github.com/theunrepentantgeek/go-vcr-tidy/pkg/vcrcleaner"

type AzureCleaningOptions struct {
	All                    *bool `help:"Clean all Azure-related monitoring interactions."`
	AsynchronousOperations *bool `help:"Clean Azure asynchronous operation monitoring interactions."`
	LongRunningOperations  *bool `help:"Clean Azure long-running operation interactions."`
	ResourceModifications  *bool `help:"Clean Azure resource modification (PUT/PATCH) monitoring interactions."`
	ResourceDeletions      *bool `help:"Clean Azure resource deletion monitoring interactions."`
}

// Options builds the vcrcleaner options based on the Azure cleaning options.
// all specifies the general 'All' option from the parent CleaningOptions.
func (opt *AzureCleaningOptions) Options(
	all *bool,
) []vcrcleaner.Option {
	var result []vcrcleaner.Option

	if opt.ShouldCleanLongRunningOperations(all) {
		result = append(result, vcrcleaner.ReduceAzureLongRunningOperationPolling())
	}

	if opt.ShouldCleanResourceModifications(all) {
		result = append(result, vcrcleaner.ReduceAzureResourceModificationMonitoring())
	}

	if opt.ShouldCleanResourceDeletions(all) {
		result = append(result, vcrcleaner.ReduceAzureResourceDeletionMonitoring())
	}

	if opt.ShouldCleanAsynchronousOperations(all) {
		result = append(result, vcrcleaner.ReduceAzureAsynchronousOperationMonitoring())
	}

	return result
}

// ShouldCleanLongRunningOperations indicates whether long-running operation monitoring should be cleaned.
// More specific options override the general 'All' option.
// all specifies the general 'All' option from the parent CleaningOptions.
func (opt *AzureCleaningOptions) ShouldCleanLongRunningOperations(all *bool) bool {
	return opt.coalesce(
		opt.LongRunningOperations,
		opt.All,
		all)
}

// ShouldCleanResourceModifications indicates whether resource modification monitoring should be cleaned.
// More specific options override the general 'All' option.
// all specifies the general 'All' option from the parent CleaningOptions.
func (opt *AzureCleaningOptions) ShouldCleanResourceModifications(all *bool) bool {
	return opt.coalesce(
		opt.ResourceModifications,
		opt.All,
		all)
}

// ShouldCleanResourceDeletions indicates whether resource deletion monitoring should be cleaned.
// More specific options override the general 'All' option.
// all specifies the general 'All' option from the parent CleaningOptions.
func (opt *AzureCleaningOptions) ShouldCleanResourceDeletions(all *bool) bool {
	return opt.coalesce(
		opt.ResourceDeletions,
		opt.All,
		all)
}

// ShouldCleanAsynchronousOperations indicates whether asynchronous operation monitoring should be cleaned.
// More specific options override the general 'All' option.
// all specifies the general 'All' option from the parent CleaningOptions.
func (opt *AzureCleaningOptions) ShouldCleanAsynchronousOperations(all *bool) bool {
	return opt.coalesce(
		opt.AsynchronousOperations,
		opt.All,
		all)
}

func (*AzureCleaningOptions) coalesce(opts ...*bool) bool {
	for _, o := range opts {
		if o != nil {
			return *o
		}
	}

	return false
}
