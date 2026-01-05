package cmd

import "github.com/theunrepentantgeek/go-vcr-tidy/pkg/vcrcleaner"

type CleaningOptions struct {
	Deletes *bool                `help:"Clean delete interactions."`
	Azure   AzureCleaningOptions `embed:""                          prefix:"azure-"`
}

func (opt *CleaningOptions) Options() []vcrcleaner.Option {
	var result []vcrcleaner.Option
	if opt.ShouldCleanDeletes() {
		result = append(result, vcrcleaner.ReduceDeleteMonitoring())
	}

	azureOptions := opt.Azure.Options()
	result = append(result, azureOptions...)

	return result
}

func (opt *CleaningOptions) ShouldCleanDeletes() bool {
	if opt.Deletes != nil {
		return *opt.Deletes
	}

	return false
}
