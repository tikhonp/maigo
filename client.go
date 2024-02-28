package maigo

import (
	"github.com/TikhonP/maigo/internal/api"
	"github.com/TikhonP/maigo/internal/net"
	"net/url"
)

// Client encapsulates a range of functionality related to
// actions for Medsenger AI actions.
type Client struct {
	apiKey string // Secret assigned to agent.
	host   string // Medsenger service target hostname.
}

// urlAppendingPath generates *url.URL based on Client.host and provided path.
func (c *Client) urlAppendingPath(path string) *url.URL {
	return &url.URL{Scheme: "https", Host: c.host, Path: path}
}

// Init creates Medsenger AI Client with provided apiKey.
//
// Default host is "medsenger.ru". Host can be modified using Client.UpdateHost method.
func Init(apiKey string) *Client {
	return &Client{apiKey: apiKey, host: "medsenger.ru"}
}

// UpdateHost modifies host for all Client requests.
func (c *Client) UpdateHost(host string) *Client {
	c.host = host
	return c
}

// GetContractInfo fetches information about contract with provided contractId.
func (c *Client) GetContractInfo(contractId int) (*api.ContractInfo, error) {
	type Request struct {
		ContractId int    `json:"contract_id"`
		ApiKey     string `json:"api_key"`
	}
	reqUrl := c.urlAppendingPath("/api/agents/patient/info")
	return net.MakeRequest[Request, api.ContractInfo](reqUrl, Request{contractId, c.apiKey})
}

// GetClinicsInfo fetches all clinics.
func (c *Client) GetClinicsInfo() (*api.Clinics, error) {
	type Request struct {
		ApiKey string `json:"api_key"`
	}
	request := Request{ApiKey: c.apiKey}
	reqUrl := c.urlAppendingPath("/api/agents/clinics")
	return net.MakeRequest[Request, api.Clinics](reqUrl, request)
}

// SendMessage sends message in contract chat.
func (c *Client) SendMessage(contractId int, text string, opts ...SendMessageOption) (msgId int, err error) {
	type Request struct {
		ContractId int                 `json:"contract_id"`
		ApiKey     string              `json:"api_key"`
		Message    *sendMessageOptions `json:"message"`
	}
	type Response struct {
		State string `json:"state"`
		Id    int    `json:"id"`
	}

	message := newSendMessageOptions(text, opts...)
	request := Request{ContractId: contractId, ApiKey: c.apiKey, Message: message}
	reqUrl := c.urlAppendingPath("/api/agents/message")

	resp, err := net.MakeRequest[Request, Response](reqUrl, request)
	return resp.Id, err
}

// OutDateMessage hides the message from a chat.
func (c *Client) OutDateMessage(contractId int, messageId int) error {
	type Request struct {
		ContractId int    `json:"contract_id"`
		ApiKey     string `json:"api_key"`
		MessageId  int    `json:"message_id"`
	}
	type response struct{}

	request := Request{ContractId: contractId, ApiKey: c.apiKey, MessageId: messageId}
	reqUrl := c.urlAppendingPath("/api/agents/message/outdate")
	_, err := net.MakeRequest[Request, response](reqUrl, request)
	return err
}
