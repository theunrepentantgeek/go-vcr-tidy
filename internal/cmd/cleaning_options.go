package cmd

import "github.com/theunrepentantgeek/go-vcr-tidy/pkg/vcrcleaner"

type CleaningOptions struct {
	All     *bool `help:"Clean all supported interaction types."`
	Deletes *bool `help:"Clean delete interactions."`

	Azure AzureCleaningOptions `embed:"" prefix:"azure-"`
}

func (opt *CleaningOptions) Options() []vcrcleaner.Option {
	var result []vcrcleaner.Option
	if opt.ShouldCleanDeletes(opt.All) {
		result = append(result, vcrcleaner.ReduceDeleteMonitoring())
	}

	azureOptions := opt.Azure.Options(opt.All)
	result = append(result, azureOptions...)

	return result
}

func (opt *CleaningOptions) ShouldCleanDeletes(all *bool) bool {
	return opt.coalesce(
		opt.Deletes,
		opt.All,
		all,
	)
}

func (*CleaningOptions) coalesce(opts ...*bool) bool {
	for _, o := range opts {
		if o != nil {
			return *o
		}
	}

	return false
}
