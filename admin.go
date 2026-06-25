package maigo

import (
	"github.com/tikhonp/maigo/internal/api"
	"github.com/tikhonp/maigo/internal/net"
)

// NotifyAdmin sends a message to the Medsenger administrators. channel selects
// the notification channel; an empty channel defaults to "it".
func (c *Client) NotifyAdmin(message string, channel string) error {
	if channel == "" {
		channel = "it"
	}
	type Request struct {
		api.TokenOnlyRequest
		Message string `json:"message"`
		Channel string `json:"channel"`
	}
	request := Request{
		TokenOnlyRequest: api.TokenOnlyRequest{APIKey: c.apiKey},
		Message:          message,
		Channel:          channel,
	}
	reqURL := c.urlAppendingPath("/api/agents/notify_admin")
	return net.MakeRequestWithEmptyResponse(reqURL, request)
}

// GetAdminClinicInfo fetches clinic information for an admin, authorized by the
// provided admin agent token. The response shape is not strongly typed by this
// SDK.
func (c *Client) GetAdminClinicInfo(adminToken string) (map[string]any, error) {
	type Request struct {
		api.TokenOnlyRequest
		AgentToken string `json:"agent_token"`
	}
	request := Request{
		TokenOnlyRequest: api.TokenOnlyRequest{APIKey: c.apiKey},
		AgentToken:       adminToken,
	}
	reqURL := c.urlAppendingPath("/api/agents/admin/clinic_info")
	resp, err := net.MakeRequest[Request, map[string]any](reqURL, request)
	if err != nil || resp == nil {
		return nil, err
	}
	return *resp, nil
}
