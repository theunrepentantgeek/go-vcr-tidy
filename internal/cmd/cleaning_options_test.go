package cmd

import (
	"testing"

	. "github.com/onsi/gomega"
)

// CleaningOptions.Options Tests

func TestCleaningOptions_Options(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		deletes       *bool
		azureAll      *bool
		expectedCount int
	}{
		"WithNoOptionsSet_ReturnsEmptySlice": {
			expectedCount: 0,
		},
		"WithOnlyDeletesSet_ReturnsOneDeleteOption": {
			deletes:       toPtr(true),
			expectedCount: 1,
		},
		"WithOnlyAzureAllSet_ReturnsFourAzureOptions": {
			azureAll:      toPtr(true),
			expectedCount: 4,
		},
		"WithDeletesAndAzureAll_ReturnsFourOptions": {
			deletes:       toPtr(true),
			azureAll:      toPtr(true),
			expectedCount: 5,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)

			opt := &CleaningOptions{
				Deletes: c.deletes,
				Azure: AzureCleaningOptions{
					All: c.azureAll,
				},
			}

			result := opt.Options()

			g.Expect(result).To(HaveLen(c.expectedCount))
		})
	}
}

// CleaningOptions.ShouldCleanDeletes Tests

func TestCleaningOptions_ShouldCleanDeletes(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		deletes  *bool
		expected bool
	}{
		"WithDeletesTrue_ReturnsTrue": {
			deletes:  toPtr(true),
			expected: true,
		},
		"WithDeletesFalse_ReturnsFalse": {
			deletes:  toPtr(false),
			expected: false,
		},
		"WithDeletesNil_ReturnsFalse": {
			expected: false,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)

			opt := &CleaningOptions{
				Deletes: c.deletes,
			}

			result := opt.ShouldCleanDeletes()

			g.Expect(result).To(Equal(c.expected))
		})
	}
}
