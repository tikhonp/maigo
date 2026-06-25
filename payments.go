package maigo

import (
	"github.com/tikhonp/maigo/internal/api"
	"github.com/tikhonp/maigo/internal/net"
)

// Payment is a single payment entry. The Medsenger payments response is not
// strongly typed by this SDK; access fields by key.
type Payment = map[string]any

type paymentsResponse struct {
	Payments []Payment `json:"payments"`
}

// RequestPayment requests a payment from the patient on a contract. invID is the
// agent-side invoice identifier, amount is the payment sum and title is shown to
// the patient.
func (c *Client) RequestPayment(contractID int, invID string, amount int, title string) error {
	type Request struct {
		api.TokenAndContractRequest
		Title string `json:"title"`
		InvID string `json:"inv_id"`
		Sum   int    `json:"sum"`
	}
	request := Request{
		TokenAndContractRequest: c.tokenAndContractRequest(contractID),
		Title:                   title,
		InvID:                   invID,
		Sum:                     amount,
	}
	reqURL := c.urlAppendingPath("/api/agents/payments/request")
	return net.MakeRequestWithEmptyResponse(reqURL, request)
}

// GetPayments fetches all payments for a contract.
func (c *Client) GetPayments(contractID int) ([]Payment, error) {
	request := c.tokenAndContractRequest(contractID)
	reqURL := c.urlAppendingPath("/api/agents/payments")
	resp, err := net.MakeRequest[api.TokenAndContractRequest, paymentsResponse](reqURL, request)
	if err != nil {
		return nil, err
	}
	return resp.Payments, nil
}
