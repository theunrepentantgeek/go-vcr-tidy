package cmd

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestAzureCleaningOptions_Options(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		all                    *bool
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
		"WithAllSetToFalse_ReturnsNoOptions": {
			all:           toPtr(false),
			expectedCount: 0,
		},
		"WithAllSetToNil_ReturnsNoOptions": {
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
				All:                    c.all,
				AsynchronousOperations: c.asynchronousOperations,
				LongRunningOperations:  c.longRunningOperations,
				ResourceModifications:  c.resourceModifications,
				ResourceDeletions:      c.resourceDeletions,
			}

			result := opt.Options()

			g.Expect(result).To(HaveLen(c.expectedCount))
		})
	}
}

//nolint:dupl // Intentional duplication across similar test methods
func TestAzureCleaningOptions_ShouldCleanLongRunningOperations(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		longRunningOperations *bool
		all                   *bool
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
			all:      toPtr(true),
			expected: true,
		},
		"WithLongRunningOperationsNilAndAllFalse_ReturnsFalse": {
			all:      toPtr(false),
			expected: false,
		},
		"WithLongRunningOperationsNilAndAllNil_ReturnsFalse": {
			expected: false,
		},
		"WithLongRunningOperationsTrueAndAllFalse_ReturnsTrueOverridingAll": {
			longRunningOperations: toPtr(true),
			all:                   toPtr(false),
			expected:              true,
		},
		"WithLongRunningOperationsFalseAndAllTrue_ReturnsFalseOverridingAll": {
			longRunningOperations: toPtr(false),
			all:                   toPtr(true),
			expected:              false,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)

			opt := &AzureCleaningOptions{
				LongRunningOperations: c.longRunningOperations,
				All:                   c.all,
			}

			result := opt.ShouldCleanLongRunningOperations()

			g.Expect(result).To(Equal(c.expected))
		})
	}
}

//nolint:dupl // Intentional duplication across similar test methods
func TestAzureCleaningOptions_ShouldCleanResourceModifications(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		resourceModifications *bool
		all                   *bool
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
			all:                   toPtr(false),
			expected:              true,
		},
		"WithResourceModificationsFalseAndAllTrue_ReturnsFalseOverridingAll": {
			resourceModifications: toPtr(false),
			all:                   toPtr(true),
			expected:              false,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)

			opt := &AzureCleaningOptions{
				ResourceModifications: c.resourceModifications,
				All:                   c.all,
			}

			result := opt.ShouldCleanResourceModifications()

			g.Expect(result).To(Equal(c.expected))
		})
	}
}

//nolint:dupl // Intentional duplication across similar test methods
func TestAzureCleaningOptions_ShouldCleanResourceDeletions(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		resourceDeletions *bool
		all               *bool
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
				All:               c.all,
			}

			result := opt.ShouldCleanResourceDeletions()

			g.Expect(result).To(Equal(c.expected))
		})
	}
}
