package azure

// azureResourceProperties represents the properties section of an Azure resource response.
type azureResourceProperties struct {
	ProvisioningState string `json:"provisioningState"`
}

// azureResourceResponse represents the structure of an Azure resource response.
type azureResourceResponse struct {
	Properties azureResourceProperties `json:"properties"`
}
