package azure

// ResourceProperties represents the properties section of an Azure resource response.
type ResourceProperties struct {
	ProvisioningState string `json:"provisioningState"`
}

// ResourceResponse represents the structure of an Azure resource response.
type ResourceResponse struct {
	Properties ResourceProperties `json:"properties"`
}
