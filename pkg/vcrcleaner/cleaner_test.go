package vcrcleaner

import (
	"encoding/json"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/neilotoole/slogt"
	"github.com/sebdah/goldie/v2"
	"gopkg.in/dnaeon/go-vcr.v4/pkg/cassette"

	"github.com/theunrepentantgeek/go-vcr-tidy/internal/azure"
)

//nolint:funlen // length comes from multiple test cases
func TestGolden_CleanerClean_givenRecording_removesExpectedInteractions(t *testing.T) {
	t.Parallel()

	operationStatusColumn := withColumn("Operation Status", operationStatus)
	provisioningStateColumn := withColumn("Provisioning State", provisioningState)

	cases := map[string]struct {
		option        Option
		recordingPath string
		extraColumn   *cassetteColumn
	}{
		"reduce-long-running-operation-polling-sql-server": {
			option:        ReduceAzureLongRunningOperationPolling(),
			recordingPath: "Test_SQL_Server_FailoverGroup_CRUD",
			extraColumn:   operationStatusColumn,
		},
		"reduce-long-running-operation-polling-managed-cluster": {
			option:        ReduceAzureLongRunningOperationPolling(),
			recordingPath: "Test_AKS_ManagedCluster_20231001_CRUD",
			extraColumn:   operationStatusColumn,
		},
		"reduce-long-running-operation-polling-api-management": {
			option:        ReduceAzureLongRunningOperationPolling(),
			recordingPath: "Test_Apimanagement_v1api20220801_CreationAndDeletion",
			extraColumn:   operationStatusColumn,
		},
		"reduce-asynchronous-operation-polling-api-management": {
			option:        ReduceAzureAsynchronousOperationPolling(),
			recordingPath: "Test_Apimanagement_v1api20220801_CreationAndDeletion",
		},
		"reduce-azure-resource-modification-monitoring-eventhub": {
			option:        ReduceAzureResourceModificationMonitoring(),
			recordingPath: "Test_EventHub_Namespace_v20240101_CRUD",
			extraColumn:   provisioningStateColumn,
		},
		"reduce-asynchronous-operation-polling-sql": {
			option:        ReduceAzureLongRunningOperationPolling(),
			recordingPath: "Test_Sql_v1api20211101_CreationAndDeletion",
			extraColumn:   operationStatusColumn,
		},
	}

	/*

		Option{
			"reduce-delete-monitoring":                      ReduceDeleteMonitoring(),
			"reduce-long-running-operation-polling":         ReduceAzureLongRunningOperationPolling(),
			"reduce-azure-resource-modification-monitoring": ReduceAzureResourceModificationMonitoring(),
			"reduce-azure-resource-deletion-monitoring":     ReduceAzureResourceDeletionMonitoring(),
		}
	*/

	// Run each option as a golden test
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g := NewGomegaWithT(t)

			log := slogt.New(t)

			// Load the cassette from the file
			fp := filepath.Join("testdata", c.recordingPath)
			cas, err := cassette.Load(fp)
			g.Expect(err).NotTo(HaveOccurred(), "loading cassette from %s", c.recordingPath)

			// Clean it
			cleaner := New(log, c.option)

			_, err = cleaner.CleanCassette(cas)
			g.Expect(err).NotTo(HaveOccurred(), "cleaning cassette from %s", c.recordingPath)

			// Get summary for the cleaned cassette
			var cleaned string
			if c.extraColumn == nil {
				cleaned = cassetteSummary(cas)
			} else {
				cleaned = cassetteSummary(cas, c.extraColumn)
			}

			// use goldie to assert the changes made
			gold := goldie.New(t, goldie.WithTestNameForDir(true))
			gold.Assert(t, name, []byte(cleaned))
		})
	}
}

func provisioningState(i *cassette.Interaction) string {
	var resp azure.ResourceResponse

	if err := json.Unmarshal([]byte(i.Response.Body), &resp); err != nil {
		// Not a JSON payload, return empty string
		return ""
	}

	return resp.Properties.ProvisioningState
}

func operationStatus(i *cassette.Interaction) string {
	var resp struct {
		Status string `json:"status"`
	}

	if err := json.Unmarshal([]byte(i.Response.Body), &resp); err != nil {
		// Not a JSON payload, return empty string
		return ""
	}

	return resp.Status
}
