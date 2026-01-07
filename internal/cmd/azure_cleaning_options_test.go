package cmd

import (
	"testing"

	. "github.com/onsi/gomega"
)

//nolint:funlen // Length comes from the number of test cases
func TestAzureCleaningOptions_Options(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		all                    *bool
		azureAll               *bool
		asynchronousOperations *bool
		longRunningOperations  *bool
		resourceModifications  *bool
		resourceDeletions      *bool
		expectedCount          int
	}{
		"WithAllSetToTrue_ReturnsAllOptions": {
			all:           toPtr(true),
			expectedCount: 4,
		},
		"WithAzureAllSetToTrue_ReturnsAllOptions": {
			azureAll:      toPtr(true),
			expectedCount: 4,
		},
		"WithAzureAllSetToFalse_ReturnsNoOptions": {
			azureAll:      toPtr(false),
			expectedCount: 0,
		},
		"WithAzureAllSetToNil_ReturnsNoOptions": {
			expectedCount: 0,
		},
		"WithOnlyLongRunningOperationsSet_ReturnsOneLROOption": {
			longRunningOperations: toPtr(true),
			expectedCount:         1,
		},
		"WithOnlyResourceModificationsSet_ReturnsOneResourceModificationOption": {
			resourceModifications: toPtr(true),
			expectedCount:         1,
		},
		"WithOnlyResourceDeletionsSet_ReturnsOneResourceDeletionOption": {
			resourceDeletions: toPtr(true),
			expectedCount:     1,
		},
		"WithOnlyAsynchronousOperationsSet_ReturnsOneAsyncOperationOption": {
			asynchronousOperations: toPtr(true),
			expectedCount:          1,
		},
		"WithAllSpecificOptionsSet_ReturnsAllOptions": {
			longRunningOperations:  toPtr(true),
			resourceModifications:  toPtr(true),
			resourceDeletions:      toPtr(true),
			asynchronousOperations: toPtr(true),
			expectedCount:          4,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)

			opt := &AzureCleaningOptions{
				All:                    c.azureAll,
				AsynchronousOperations: c.asynchronousOperations,
				LongRunningOperations:  c.longRunningOperations,
				ResourceModifications:  c.resourceModifications,
				ResourceDeletions:      c.resourceDeletions,
			}

			result := opt.Options(c.all)

			g.Expect(result).To(HaveLen(c.expectedCount))
		})
	}
}

func TestAzureCleaningOptions_ShouldCleanLongRunningOperations(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		all                   *bool
		azureAll              *bool
		longRunningOperations *bool
		expected              bool
	}{
		"WithLongRunningOperationsTrue_ReturnsTrue": {
			longRunningOperations: toPtr(true),
			expected:              true,
		},
		"WithLongRunningOperationsFalse_ReturnsFalse": {
			longRunningOperations: toPtr(false),
			expected:              false,
		},
		"WithLongRunningOperationsNilAndAllTrue_ReturnsTrue": {
			azureAll: toPtr(true),
			expected: true,
		},
		"WithLongRunningOperationsNilAndAllFalse_ReturnsFalse": {
			azureAll: toPtr(false),
			expected: false,
		},
		"WithLongRunningOperationsNilAndAllNil_ReturnsFalse": {
			expected: false,
		},
		"WithLongRunningOperationsTrueAndAllFalse_ReturnsTrueOverridingAll": {
			longRunningOperations: toPtr(true),
			azureAll:              toPtr(false),
			expected:              true,
		},
		"WithLongRunningOperationsFalseAndAllTrue_ReturnsFalseOverridingAll": {
			longRunningOperations: toPtr(false),
			azureAll:              toPtr(true),
			expected:              false,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)

			opt := &AzureCleaningOptions{
				LongRunningOperations: c.longRunningOperations,
				All:                   c.azureAll,
			}

			result := opt.ShouldCleanLongRunningOperations(nil)

			g.Expect(result).To(Equal(c.expected))
		})
	}
}

//nolint:dupl,funlen // Intentional duplication & length from test cases
func TestAzureCleaningOptions_ShouldCleanResourceModifications(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		all                   *bool
		azureAll              *bool
		resourceModifications *bool
		expected              bool
	}{
		"WithResourceModificationsTrue_ReturnsTrue": {
			resourceModifications: toPtr(true),
			expected:              true,
		},
		"WithResourceModificationsFalse_ReturnsFalse": {
			resourceModifications: toPtr(false),
			expected:              false,
		},
		"WithResourceModificationsNilAndAllAzureTrue_ReturnsTrue": {
			azureAll: toPtr(true),
			expected: true,
		},
		"WithResourceModificationsNilAndAllAzureFalse_ReturnsFalse": {
			azureAll: toPtr(false),
			expected: false,
		},
		"WithResourceModificationsNilAndAllTrue_ReturnsTrue": {
			all:      toPtr(true),
			expected: true,
		},
		"WithResourceModificationsNilAndAllFalse_ReturnsFalse": {
			all:      toPtr(false),
			expected: false,
		},
		"WithResourceModificationsNilAndAllNil_ReturnsFalse": {
			expected: false,
		},
		"WithResourceModificationsTrueAndAllFalse_ReturnsTrueOverridingAll": {
			resourceModifications: toPtr(true),
			azureAll:              toPtr(false),
			expected:              true,
		},
		"WithResourceModificationsFalseAndAllTrue_ReturnsFalseOverridingAll": {
			resourceModifications: toPtr(false),
			azureAll:              toPtr(true),
			expected:              false,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)

			opt := &AzureCleaningOptions{
				ResourceModifications: c.resourceModifications,
				All:                   c.azureAll,
			}

			result := opt.ShouldCleanResourceModifications(c.all)

			g.Expect(result).To(Equal(c.expected))
		})
	}
}

//nolint:dupl,funlen // Intentional duplication & length from test cases
func TestAzureCleaningOptions_ShouldCleanResourceDeletions(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		all               *bool
		azureAll          *bool
		resourceDeletions *bool
		expected          bool
	}{
		"WithResourceDeletionsTrue_ReturnsTrue": {
			resourceDeletions: toPtr(true),
			expected:          true,
		},
		"WithResourceDeletionsFalse_ReturnsFalse": {
			resourceDeletions: toPtr(false),
			expected:          false,
		},
		"WithResourceDeletionsNilAndAzureAllTrue_ReturnsTrue": {
			azureAll: toPtr(true),
			expected: true,
		},
		"WithResourceDeletionsNilAndAzureAllFalse_ReturnsFalse": {
			azureAll: toPtr(false),
			expected: false,
		},
		"WithResourceDeletionsNilAndAllTrue_ReturnsTrue": {
			all:      toPtr(true),
			expected: true,
		},
		"WithResourceDeletionsNilAndAllFalse_ReturnsFalse": {
			all:      toPtr(false),
			expected: false,
		},
		"WithResourceDeletionsNilAndAllNil_ReturnsFalse": {
			expected: false,
		},
		"WithResourceDeletionsTrueAndAllFalse_ReturnsTrueOverridingAll": {
			resourceDeletions: toPtr(true),
			all:               toPtr(false),
			expected:          true,
		},
		"WithResourceDeletionsFalseAndAllTrue_ReturnsFalseOverridingAll": {
			resourceDeletions: toPtr(false),
			all:               toPtr(true),
			expected:          false,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)

			opt := &AzureCleaningOptions{
				ResourceDeletions: c.resourceDeletions,
				All:               c.azureAll,
			}

			result := opt.ShouldCleanResourceDeletions(c.all)

			g.Expect(result).To(Equal(c.expected))
		})
	}
}
