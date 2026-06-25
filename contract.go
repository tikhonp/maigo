package maigo

import (
	"github.com/tikhonp/maigo/internal/api"
	"github.com/tikhonp/maigo/internal/net"
)

// UpdateCache asks Medsenger to refresh its cached data for a contract.
func (c *Client) UpdateCache(contractID int) error {
	request := c.tokenAndContractRequest(contractID)
	reqURL := c.urlAppendingPath("/api/agents/cache")
	return net.MakeRequestWithEmptyResponse(reqURL, request)
}

// SetInfoMaterials sets the information materials shown for a contract.
func (c *Client) SetInfoMaterials(contractID int, materials any) error {
	type Request struct {
		api.TokenAndContractRequest
		InfoMaterials any `json:"info_materials"`
	}
	request := Request{
		TokenAndContractRequest: c.tokenAndContractRequest(contractID),
		InfoMaterials:           materials,
	}
	reqURL := c.urlAppendingPath("/api/agents/info_materials/set")
	return net.MakeRequestWithEmptyResponse(reqURL, request)
}

// SetContractParam sets a named parameter on a contract. Pass an empty value to
// remove the parameter.
func (c *Client) SetContractParam(contractID int, name, value string) error {
	type Request struct {
		api.TokenAndContractRequest
		Name  string `json:"name"`
		Value string `json:"value"`
	}
	request := Request{
		TokenAndContractRequest: c.tokenAndContractRequest(contractID),
		Name:                    name,
		Value:                   value,
	}
	reqURL := c.urlAppendingPath("/api/agents/set_contract_param")
	return net.MakeRequestWithEmptyResponse(reqURL, request)
}
