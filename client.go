// Package maigo provides Go SDK for Medsenger API:
// A high performance, open source, SDK for Medsenger AI agents.
package maigo

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/tikhonp/maigo/internal/api"
	"github.com/tikhonp/maigo/internal/assert"
	pjson "github.com/tikhonp/maigo/internal/json"
	"github.com/tikhonp/maigo/internal/net"
)

// Client encapsulates a range of functionality related to
// actions for Medsenger AI actions.
type Client struct {
	apiKey string // Secret assigned to agent.
	host   string // Medsenger service target hostname.
}

func (c *Client) DebugData() string {
	return fmt.Sprintf("apiKey: %s..., host: %s", c.apiKey[:10], c.host)
}

// urlAppendingPath generates *url.URL based on Client.host and provided path.
func (c *Client) urlAppendingPath(path string) *url.URL {
	return &url.URL{Scheme: "https", Host: c.host, Path: path}
}

func (c *Client) tokenAndContractRequest(contractID int) api.TokenAndContractRequest {
	return api.TokenAndContractRequest{
		TokenOnlyRequest: api.TokenOnlyRequest{APIKey: c.apiKey},
		ContractID:       contractID,
	}
}

// Init creates Medsenger AI Client with provided apiKey.
//
// Default host is "medsenger.ru". Host can be modified using Client.UpdateHost method.
func Init(apiKey string) *Client {
	assert.Assert(len(apiKey) > 10, "apiKey must be at least 10 characters long")
	return &Client{apiKey: apiKey, host: "medsenger.ru"}
}

// UpdateHost modifies host for all Client requests.
func (c *Client) UpdateHost(host string) *Client {
	c.host = host
	return c
}

// GetContractInfo fetches information about contract with provided contractId.
func (c *Client) GetContractInfo(contractID int) (*ContractInfo, error) {
	request := c.tokenAndContractRequest(contractID)
	reqURL := c.urlAppendingPath("/api/agents/patient/info")
	return net.MakeRequest[api.TokenAndContractRequest, ContractInfo](reqURL, request)
}

// GetClinicsInfo fetches all clinics.
func (c *Client) GetClinicsInfo() (*Clinics, error) {
	request := api.TokenOnlyRequest{APIKey: c.apiKey}
	reqURL := c.urlAppendingPath("/api/agents/clinics")
	return net.MakeRequest[api.TokenOnlyRequest, Clinics](reqURL, request)
}

// SendMessage sends message in contract chat.
func (c *Client) SendMessage(contractID int, text string, opts ...SendMessageOption) (msgID int, err error) {
	type Request struct {
		api.TokenAndContractRequest
		Message *sendMessageOptions `json:"message"`
	}
	type Response struct {
		State string `json:"state"`
		ID    int    `json:"id"`
	}
	request := Request{
		TokenAndContractRequest: c.tokenAndContractRequest(contractID),
		Message:                 newSendMessageOptions(text, opts...),
	}
	reqURL := c.urlAppendingPath("/api/agents/message")
	resp, err := net.MakeRequest[Request, Response](reqURL, request)
	return resp.ID, err
}

// OutDateMessage hides the message from a chat.
func (c *Client) OutDateMessage(contractID int, messageID int) error {
	type Request struct {
		api.TokenAndContractRequest
		MessageID int `json:"message_id"`
	}
	request := Request{TokenAndContractRequest: c.tokenAndContractRequest(contractID), MessageID: messageID}
	reqURL := c.urlAppendingPath("/api/agents/message/outdate")
	return net.MakeRequestWithEmptyResponse(reqURL, request)
}

// GetCategories fetches all medical records categories.
func (c *Client) GetCategories() (*Categories, error) {
	request := api.TokenOnlyRequest{APIKey: c.apiKey}
	reqURL := c.urlAppendingPath("/api/agents/records/categories")
	return net.MakeRequest[api.TokenOnlyRequest, Categories](reqURL, request)
}

// GetAvailableCategories fetches all available medical records categories.
func (c *Client) GetAvailableCategories(contractID int) (*Categories, error) {
	request := c.tokenAndContractRequest(contractID)
	reqURL := c.urlAppendingPath("/api/agents/records/available_categories")
	return net.MakeRequest[api.TokenAndContractRequest, Categories](reqURL, request)
}

// GetRecords fetches medical records by contractId.
// By default all recrds sorted ascending by time. So if you need to get latest record you need to set limit to 1.
func (c *Client) GetRecords(contractID int, opts ...GetRecordsOption) ([]MedicalRecord, error) {
	request := getRecordsOptions{
		TokenAndContractRequest: c.tokenAndContractRequest(contractID),
	}
	applyGetRecordsOptions(&request, opts...)
	reqURL := c.urlAppendingPath("/api/agents/records/get/all")
	records, err := net.MakeRequest[getRecordsOptions, []MedicalRecord](reqURL, request)
	if err != nil {
		return nil, err
	}
	return *records, nil
}

// GetRecord fetches a record by contractId and recordId.
func (c *Client) GetRecord(contractID int, recordID int) (*MedicalRecord, error) {
	type Request struct {
		api.TokenAndContractRequest
		RecordID int `json:"record_id"`
	}
	request := Request{
		TokenAndContractRequest: c.tokenAndContractRequest(contractID),
		RecordID:                recordID,
	}
	reqURL := c.urlAppendingPath("/api/agents/records/get")
	return net.MakeRequest[Request, MedicalRecord](reqURL, request)
}

func (c *Client) AddHooksForCategories(contractID int) {
	// TODO: implement it
	panic("not implemented")
}

func (c *Client) RemoveHooksForCategories(contractID int) {
	// TODO: implement it
	panic("not implemented")
}

// SendRecordAddition commit addition to a record.
func (c *Client) SendRecordAddition(contractID int, recordID int, note string) error {
	type Request struct {
		api.TokenAndContractRequest
		RecordID int    `json:"record_id"`
		Note     string `json:"addition"`
	}
	request := Request{
		TokenAndContractRequest: c.tokenAndContractRequest(contractID),
		RecordID:                recordID,
		Note:                    note,
	}
	reqURL := c.urlAppendingPath("/api/agents/records/addition")
	return net.MakeRequestWithEmptyResponse(reqURL, request)
}

// AddRecord adds medical record to Medsenger medical records table for contract. Returns recordId.
func (c *Client) AddRecord(contractID int, categoryName, value string, recordTime time.Time, params *json.Marshaler) (*int, error) {
	type Request struct {
		api.TokenAndContractRequest
		CategoryName string          `json:"category_name"`
		Value        string          `json:"value"`
		ReturnID     bool            `json:"return_id"`
		Time         pjson.Timestamp `json:"time"`
		Params       *json.Marshaler `json:"params,omitempty"`
	}
	request := Request{
		TokenAndContractRequest: c.tokenAndContractRequest(contractID),
		CategoryName:            categoryName,
		Value:                   value,
		ReturnID:                true,
		Time:                    pjson.Timestamp{Time: recordTime},
		Params:                  params,
	}
	reqURL := c.urlAppendingPath("/api/agents/records/add")
	ids, err := net.MakeRequest[Request, []int](reqURL, request)
	if err != nil {
		return nil, err
	}
	if len(*ids) == 0 {
		return nil, errors.New("empty id response")
	}
	return &(*ids)[0], nil
}

type Record struct {
	CategoryName string          `json:"category_name"`
	Value        string          `json:"value"`
	Time         pjson.Timestamp `json:"time"`
}

func NewRecord(categoryName, value string, time time.Time) Record {
	return Record{
		CategoryName: categoryName,
		Value:        value,
		Time:         pjson.Timestamp{Time: time},
	}
}

// AddRecords adds multiple records to Medsenger medical records table for contract. Returns recordIds.
func (c *Client) AddRecords(contractID int, records []Record) ([]int, error) {
	type Request struct {
		api.TokenAndContractRequest
		Values   []Record `json:"values"`
		ReturnID bool     `json:"return_id"`
	}
	request := Request{
		TokenAndContractRequest: c.tokenAndContractRequest(contractID),
		Values:                  records,
		ReturnID:                true,
	}
	reqURL := c.urlAppendingPath("/api/agents/records/add")
	ids, err := net.MakeRequest[Request, []int](reqURL, request)
	if err != nil {
		return nil, err
	}
	return *ids, nil
}
