package maigo

import (
	"github.com/tikhonp/maigo/internal/api"
	"github.com/tikhonp/maigo/internal/net"
)

// Message is a single chat message. The Medsenger messages response is not
// strongly typed by this SDK; access fields by key.
type Message = map[string]any

type messagesResponse struct {
	Messages []Message `json:"messages"`
}

// GetMessages fetches chat messages for a contract starting after fromID. Pass
// fromID 0 to fetch from the beginning.
func (c *Client) GetMessages(contractID int, fromID int) ([]Message, error) {
	type Request struct {
		api.TokenAndContractRequest
		FromID int `json:"from_id"`
	}
	request := Request{
		TokenAndContractRequest: c.tokenAndContractRequest(contractID),
		FromID:                  fromID,
	}
	reqURL := c.urlAppendingPath("/api/agents/messages")
	resp, err := net.MakeRequest[Request, messagesResponse](reqURL, request)
	if err != nil {
		return nil, err
	}
	return resp.Messages, nil
}
