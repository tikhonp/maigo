package maigo

import (
	"github.com/tikhonp/maigo/internal/api"
	"github.com/tikhonp/maigo/internal/net"
)

// File is a file attached to a medical record, including its base64-encoded
// contents.
type File struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	Base64 string `json:"base64"`
}

// GetFile downloads a file attached to a record on a contract.
func (c *Client) GetFile(contractID int, fileID int) (*File, error) {
	type Request struct {
		api.TokenAndContractRequest
		FileID int `json:"file_id"`
	}
	request := Request{
		TokenAndContractRequest: c.tokenAndContractRequest(contractID),
		FileID:                  fileID,
	}
	reqURL := c.urlAppendingPath("/api/agents/records/file")
	return net.MakeRequest[Request, File](reqURL, request)
}

// GetFileLink requests a temporary download link for a file. hours sets how long
// the link is valid (the Medsenger default is 2); isOneTime makes the link
// usable only once. The response shape is not strongly typed by this SDK.
func (c *Client) GetFileLink(fileID int, hours int, isOneTime bool) (map[string]any, error) {
	if hours <= 0 {
		hours = 2
	}
	type Request struct {
		api.TokenOnlyRequest
		FileID    int  `json:"file_id"`
		Hours     int  `json:"hours"`
		IsOneTime bool `json:"is_one_time"`
	}
	request := Request{
		TokenOnlyRequest: api.TokenOnlyRequest{APIKey: c.apiKey},
		FileID:           fileID,
		Hours:            hours,
		IsOneTime:        isOneTime,
	}
	reqURL := c.urlAppendingPath("/api/agents/records/file/link")
	resp, err := net.MakeRequest[Request, map[string]any](reqURL, request)
	if err != nil || resp == nil {
		return nil, err
	}
	return *resp, nil
}

// GetAttachment downloads a message attachment by id.
func (c *Client) GetAttachment(attachmentID int) (*Attachment, error) {
	type Request struct {
		api.TokenOnlyRequest
		AttachmentID int `json:"attachment_id"`
	}
	request := Request{
		TokenOnlyRequest: api.TokenOnlyRequest{APIKey: c.apiKey},
		AttachmentID:     attachmentID,
	}
	reqURL := c.urlAppendingPath("/api/agents/attachment")
	return net.MakeRequest[Request, Attachment](reqURL, request)
}

// GetImage downloads an image attachment by id at the requested size.
func (c *Client) GetImage(imageID int, size string) (*Attachment, error) {
	type Request struct {
		api.TokenOnlyRequest
		AttachmentID int    `json:"attachment_id"`
		Size         string `json:"size"`
	}
	request := Request{
		TokenOnlyRequest: api.TokenOnlyRequest{APIKey: c.apiKey},
		AttachmentID:     imageID,
		Size:             size,
	}
	reqURL := c.urlAppendingPath("/api/agents/image")
	return net.MakeRequest[Request, Attachment](reqURL, request)
}
