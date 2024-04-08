package maigo

import (
	"encoding/json"
	"net/url"
	"time"

	"github.com/TikhonP/maigo/internal/api"
	pjson "github.com/TikhonP/maigo/internal/json"
	"github.com/TikhonP/maigo/internal/net"
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
func (c *Client) GetContractInfo(contractId int) (*ContractInfo, error) {
	request := c.tokenAndContractRequest(contractId)
	reqUrl := c.urlAppendingPath("/api/agents/patient/info")
	return net.MakeRequest[api.TokenAndContractRequest, ContractInfo](reqUrl, request)
}

// GetClinicsInfo fetches all clinics.
func (c *Client) GetClinicsInfo() (*Clinics, error) {
	request := api.TokenOnlyRequest{ApiKey: c.apiKey}
	reqUrl := c.urlAppendingPath("/api/agents/clinics")
	return net.MakeRequest[api.TokenOnlyRequest, Clinics](reqUrl, request)
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
func (c *Client) GetCategories() (*Categories, error) {
	request := api.TokenOnlyRequest{ApiKey: c.apiKey}
	reqUrl := c.urlAppendingPath("/api/agents/records/categories")
	return net.MakeRequest[api.TokenOnlyRequest, Categories](reqUrl, request)
}

// GetAvailableCategories fetches all available medical records categories.
func (c *Client) GetAvailableCategories(contractId int) (*Categories, error) {
	request := c.tokenAndContractRequest(contractId)
	reqUrl := c.urlAppendingPath("/api/agents/records/available_categories")
	return net.MakeRequest[api.TokenAndContractRequest, Categories](reqUrl, request)
}

func (c *Client) GetRecords(contractId int) (*emptyResponse, error) {
	request := c.tokenAndContractRequest(contractId)
	reqUrl := c.urlAppendingPath("/api/agents/records/get")
	return net.MakeRequest[api.TokenAndContractRequest, emptyResponse](reqUrl, request)
}

// GetRecord fetches a record by contractId and recordId.
func (c *Client) GetRecord(contractId int, recordId int) (*MedicalRecord, error) {
	type Request struct {
		api.TokenAndContractRequest
		RecordId int `json:"record_id"`
	}
	request := Request{
		TokenAndContractRequest: c.tokenAndContractRequest(contractId),
		RecordId:                recordId,
	}
	reqUrl := c.urlAppendingPath("/api/agents/records/get")
	return net.MakeRequest[Request, MedicalRecord](reqUrl, request)
}

func (c *Client) AddHooksForCategories(contractId int) {
	// TODO: implement it
	panic("not implemented")
}

func (c *Client) RemoveHooksForCategories(contractId int) {
	// TODO: implement it
	panic("not implemented")
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

// GetAgentTokenForContractId fetches agent token for contract.
func (c *Client) GetAgentTokenForContractId(contractId int) (*AgentToken, error) {
	request := c.tokenAndContractRequest(contractId)
	reqUrl := c.urlAppendingPath("/api/agents/token")
	return net.MakeRequest[api.TokenAndContractRequest, AgentToken](reqUrl, request)
}

// AddRecord adds medical record to Medsenger medical records table for contract. Returns recordIds.
func (c *Client) AddRecord(contractId int, categoryName, value string, recordTime time.Time, params *json.Marshaler) ([]int, error) {
	type Request struct {
		api.TokenAndContractRequest
		CategoryName string          `json:"category_name"`
		Value        string          `json:"value"`
		ReturnId     bool            `json:"return_id"`
		Time         pjson.Timestamp `json:"time"`
		Params       *json.Marshaler `json:"params,omitempty"`
	}
	request := Request{
		TokenAndContractRequest: c.tokenAndContractRequest(contractId),
		CategoryName:            categoryName,
		Value:                   value,
		ReturnId:                true,
		Time:                    pjson.Timestamp{Time: recordTime},
		Params:                  params,
	}
	reqUrl := c.urlAppendingPath("/api/agents/records/add")
	ids, err := net.MakeRequest[Request, []int](reqUrl, request)
	if err != nil {
		return nil, err
	}
	return *ids, nil
}
