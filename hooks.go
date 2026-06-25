package maigo

import (
	"github.com/tikhonp/maigo/internal/api"
	"github.com/tikhonp/maigo/internal/net"
)

// AddHooks subscribes the agent to updates for the given record categories on a
// contract. The agent is notified whenever a record in one of these categories
// is added.
func (c *Client) AddHooks(contractID int, categories []string) error {
	type Request struct {
		api.TokenAndContractRequest
		Categories []string `json:"categories"`
	}
	request := Request{
		TokenAndContractRequest: c.tokenAndContractRequest(contractID),
		Categories:              categories,
	}
	reqURL := c.urlAppendingPath("/api/agents/hooks/add")
	return net.MakeRequestWithEmptyResponse(reqURL, request)
}

// RemoveHooks unsubscribes the agent from updates for the given record
// categories on a contract.
func (c *Client) RemoveHooks(contractID int, categories []string) error {
	type Request struct {
		api.TokenAndContractRequest
		Categories []string `json:"categories"`
	}
	request := Request{
		TokenAndContractRequest: c.tokenAndContractRequest(contractID),
		Categories:              categories,
	}
	reqURL := c.urlAppendingPath("/api/agents/hooks/remove")
	return net.MakeRequestWithEmptyResponse(reqURL, request)
}
