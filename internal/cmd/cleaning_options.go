package cmd

import "github.com/theunrepentantgeek/go-vcr-tidy/pkg/vcrcleaner"

type CleaningOptions struct {
	All               *bool `help:"Clean all supported interaction types."`
	DeferredCreations *bool `help:"Clean deferred creation interactions."`
	Deletes           *bool `help:"Clean delete interactions."`

	Azure AzureCleaningOptions `embed:"" prefix:"azure-"`
}

func (opt *CleaningOptions) Options() []vcrcleaner.Option {
	var result []vcrcleaner.Option
	if opt.ShouldCleanDeferredCreations() {
		result = append(result, vcrcleaner.ReduceDeferredCreationMonitoring())
	}

	if opt.ShouldCleanDeletes() {
		result = append(result, vcrcleaner.ReduceDeleteMonitoring())
	}

	azureOptions := opt.Azure.Options(opt.All)
	result = append(result, azureOptions...)

	return result
}

func (opt *CleaningOptions) ShouldCleanDeletes() bool {
	return opt.coalesce(
		opt.Deletes,
		opt.All)
}

func (opt *CleaningOptions) ShouldCleanDeferredCreations() bool {
	return opt.coalesce(
		opt.DeferredCreations,
		opt.All)
}

func (*CleaningOptions) coalesce(opts ...*bool) bool {
	for _, o := range opts {
		if o != nil {
			return *o
		}
	}

	return false
}
