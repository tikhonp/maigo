// Package maigo provides Go SDK for Medsenger API:
// A high performance, open source, SDK for Medsenger AI agents.
package maigo

import (
	"errors"
	"fmt"
	"net/url"
	"sync"
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

	grpcHost string      // gRPC endpoint; empty disables gRPC.
	grpc     *grpcClient // Lazily-connected gRPC client (nil when disabled).

	userMu    sync.Mutex  // Guards userCache.
	userCache map[int]int // contractID -> gRPC user_id.
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
// Pass WithGRPC to enable the gRPC transport (with REST fallback) for record and
// category reads.
func Init(apiKey string, opts ...InitOption) *Client {
	assert.Assert(len(apiKey) > 10, "apiKey must be at least 10 characters long")
	c := &Client{apiKey: apiKey, host: "medsenger.ru"}
	for _, opt := range opts {
		opt.apply(c)
	}
	if c.grpcHost != "" {
		c.grpc = newGRPCClient(apiKey, c.grpcHost)
	}
	return c
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
	if c.grpc != nil {
		if cats, err := c.grpc.getCategories(); err == nil {
			result := Categories(cats)
			return &result, nil
		}
	}
	request := api.TokenOnlyRequest{APIKey: c.apiKey}
	reqURL := c.urlAppendingPath("/api/agents/records/categories")
	return net.MakeRequest[api.TokenOnlyRequest, Categories](reqURL, request)
}

// GetAvailableCategories fetches all available medical records categories.
func (c *Client) GetAvailableCategories(contractID int) (*Categories, error) {
	if c.grpc != nil {
		if uid, err := c.resolveUserID(contractID); err == nil {
			if cats, err := c.grpc.getCategoriesForUser(uid); err == nil {
				result := Categories(cats)
				return &result, nil
			}
		}
	}
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
	if c.grpc != nil {
		if records, err := c.grpcGetRecords(contractID, &request); err == nil {
			return records, nil
		}
	}
	reqURL := c.urlAppendingPath("/api/agents/records/get/all")
	records, err := net.MakeRequest[getRecordsOptions, []MedicalRecord](reqURL, request)
	if err != nil {
		return nil, err
	}
	return *records, nil
}

// GetRecord fetches a record by contractId and recordId.
func (c *Client) GetRecord(contractID int, recordID int) (*MedicalRecord, error) {
	if c.grpc != nil {
		rec, err := c.grpc.getRecordByID(recordID)
		if err == nil {
			return rec, nil
		}
		if errors.Is(err, errGRPCNotFound) {
			return nil, nil
		}
	}
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

// DeleteRecord deletes a medical record from a contract.
func (c *Client) DeleteRecord(contractID int, recordID int) error {
	type Request struct {
		api.TokenAndContractRequest
		RecordID int `json:"record_id"`
	}
	request := Request{
		TokenAndContractRequest: c.tokenAndContractRequest(contractID),
		RecordID:                recordID,
	}
	reqURL := c.urlAppendingPath("/api/agents/records/delete")
	return net.MakeRequestWithEmptyResponse(reqURL, request)
}

// SetClassifier sets the record classifier code for a contract.
func (c *Client) SetClassifier(contractID int, code string) error {
	type Request struct {
		api.TokenAndContractRequest
		Code string `json:"code"`
	}
	request := Request{
		TokenAndContractRequest: c.tokenAndContractRequest(contractID),
		Code:                    code,
	}
	reqURL := c.urlAppendingPath("/api/agents/classifier")
	return net.MakeRequestWithEmptyResponse(reqURL, request)
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

// AddRecord adds a medical record to the Medsenger records table for a contract
// and returns the new record id.
func (c *Client) AddRecord(contractID int, categoryName string, value any, opts ...AddRecordOption) (int, error) {
	type Request struct {
		api.TokenAndContractRequest
		CategoryName string `json:"category_name"`
		Value        any    `json:"value"`
		ReturnID     bool   `json:"return_id"`
		addRecordOptions
	}
	request := Request{
		TokenAndContractRequest: c.tokenAndContractRequest(contractID),
		CategoryName:            categoryName,
		Value:                   value,
		ReturnID:                true,
		addRecordOptions:        newAddRecordOptions(opts...),
	}
	reqURL := c.urlAppendingPath("/api/agents/records/add")
	ids, err := net.MakeRequest[Request, []int](reqURL, request)
	if err != nil {
		return 0, err
	}
	if len(*ids) == 0 {
		return 0, errors.New("empty id response")
	}
	return (*ids)[0], nil
}

type Record struct {
	CategoryName string           `json:"category_name"`
	Value        any              `json:"value"`
	Time         *pjson.Timestamp `json:"time,omitempty"`
	Params       map[string]any   `json:"params,omitempty"`
	Files        []Attachment     `json:"files,omitempty"`
	Replace      bool             `json:"replace"`
}

func NewRecord(categoryName string, value any, recordTime time.Time) Record {
	return Record{
		CategoryName: categoryName,
		Value:        value,
		Time:         &pjson.Timestamp{Time: recordTime},
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

// DecodeAgentJWT decodes JWT token issued for agent and returns its claims.
func (c *Client) DecodeAgentJWT(tokenString string) (*JWTClaims, error) {
	return decodeAgentJWT(tokenString, c.apiKey)
}

// ValidateAgentJWT decodes the token and verifies it is an "agent_access" token
// with at least one role, returning its claims. It returns ErrWrongTokenType or
// ErrNoRoles when those checks fail.
func (c *Client) ValidateAgentJWT(tokenString string) (*JWTClaims, error) {
	return validateAgentJWT(tokenString, c.apiKey)
}
