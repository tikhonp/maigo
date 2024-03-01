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

func (c *Client) tokenAndContractRequest(contractId int) api.TokenAndContractRequest {
	return api.TokenAndContractRequest{
		TokenOnlyRequest: api.TokenOnlyRequest{ApiKey: c.apiKey},
		ContractId:       contractId,
	}
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

type emptyResponse struct{}

// GetContractInfo fetches information about contract with provided contractId.
func (c *Client) GetContractInfo(contractId int) (*api.ContractInfo, error) {
	request := c.tokenAndContractRequest(contractId)
	reqUrl := c.urlAppendingPath("/api/agents/patient/info")
	return net.MakeRequest[api.TokenAndContractRequest, api.ContractInfo](reqUrl, request)
}

// GetClinicsInfo fetches all clinics.
func (c *Client) GetClinicsInfo() (*api.Clinics, error) {
	request := api.TokenOnlyRequest{ApiKey: c.apiKey}
	reqUrl := c.urlAppendingPath("/api/agents/clinics")
	return net.MakeRequest[api.TokenOnlyRequest, api.Clinics](reqUrl, request)
}

// SendMessage sends message in contract chat.
func (c *Client) SendMessage(contractId int, text string, opts ...SendMessageOption) (msgId int, err error) {
	type Request struct {
		api.TokenAndContractRequest
		Message *sendMessageOptions `json:"message"`
	}
	type Response struct {
		State string `json:"state"`
		Id    int    `json:"id"`
	}
	request := Request{
		TokenAndContractRequest: c.tokenAndContractRequest(contractId),
		Message:                 newSendMessageOptions(text, opts...),
	}
	reqUrl := c.urlAppendingPath("/api/agents/message")
	resp, err := net.MakeRequest[Request, Response](reqUrl, request)
	return resp.Id, err
}

// OutDateMessage hides the message from a chat.
func (c *Client) OutDateMessage(contractId int, messageId int) error {
	type Request struct {
		api.TokenAndContractRequest
		MessageId int `json:"message_id"`
	}
	request := Request{TokenAndContractRequest: c.tokenAndContractRequest(contractId), MessageId: messageId}
	reqUrl := c.urlAppendingPath("/api/agents/message/outdate")
	return net.MakeRequestWithEmptyResponse(reqUrl, request)
}

// GetCategories fetches all medical records categories.
func (c *Client) GetCategories() (*api.Categories, error) {
	request := api.TokenOnlyRequest{ApiKey: c.apiKey}
	reqUrl := c.urlAppendingPath("/api/agents/records/categories")
	return net.MakeRequest[api.TokenOnlyRequest, api.Categories](reqUrl, request)
}

// GetAvailableCategories fetches all available medical records categories.
func (c *Client) GetAvailableCategories(contractId int) (*api.Categories, error) {
	request := c.tokenAndContractRequest(contractId)
	reqUrl := c.urlAppendingPath("/api/agents/records/available_categories")
	return net.MakeRequest[api.TokenAndContractRequest, api.Categories](reqUrl, request)
}

func (c *Client) GetRecords(contractId int) (*emptyResponse, error) {
	request := c.tokenAndContractRequest(contractId)
	reqUrl := c.urlAppendingPath("/api/agents/records/get")
	return net.MakeRequest[api.TokenAndContractRequest, emptyResponse](reqUrl, request)
}

func (c *Client) GetRecord(contractId int, recordId int) (*api.MedicalRecord, error) {
	type Request struct {
		api.TokenAndContractRequest
		RecordId int `json:"record_id"`
	}
	request := Request{
		TokenAndContractRequest: c.tokenAndContractRequest(contractId),
		RecordId:                recordId,
	}
	reqUrl := c.urlAppendingPath("/api/agents/records/get")
	return net.MakeRequest[Request, api.MedicalRecord](reqUrl, request)
}

func (c *Client) AddHooksForCategories(contractId int) {
	// TODO: implement it
}

func (c *Client) RemoveHooksForCategories(contractId int) {
	// TODO: implement it
}

// SendRecordAddition commit addition to a record.
func (c *Client) SendRecordAddition(contractId int, recordId int, note string) error {
	type Request struct {
		api.TokenAndContractRequest
		RecordId int    `json:"record_id"`
		Note     string `json:"addition"`
	}
	request := Request{
		TokenAndContractRequest: c.tokenAndContractRequest(contractId),
		RecordId:                recordId,
		Note:                    note,
	}
	reqUrl := c.urlAppendingPath("/api/agents/records/addition")
	return net.MakeRequestWithEmptyResponse(reqUrl, request)
}
