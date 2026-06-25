package maigo

import (
	"github.com/tikhonp/maigo/internal/api"
	"github.com/tikhonp/maigo/internal/net"
)

type sendOrderOptions struct {
	ReceiverID int            `json:"receiver_id,omitempty"`
	Params     map[string]any `json:"params,omitempty"`
}

func newSendOrderOptions(opts ...SendOrderOption) sendOrderOptions {
	o := sendOrderOptions{}
	for _, opt := range opts {
		opt.apply(&o)
	}
	return o
}

type SendOrderOption interface {
	apply(*sendOrderOptions)
}

// funcSendOrderOption wraps a function that modifies sendOrderOptions into an
// implementation of the SendOrderOption interface.
type funcSendOrderOption struct {
	f func(*sendOrderOptions)
}

func (fo *funcSendOrderOption) apply(o *sendOrderOptions) {
	fo.f(o)
}

func newFuncSendOrderOption(f func(*sendOrderOptions)) *funcSendOrderOption {
	return &funcSendOrderOption{f: f}
}

// WithOrderReceiverID returns a SendOrderOption which routes the order to a
// specific receiver.
func WithOrderReceiverID(receiverID int) SendOrderOption {
	return newFuncSendOrderOption(func(o *sendOrderOptions) {
		o.ReceiverID = receiverID
	})
}

// WithOrderParams returns a SendOrderOption which attaches arbitrary parameters
// to the order.
func WithOrderParams(params map[string]any) SendOrderOption {
	return newFuncSendOrderOption(func(o *sendOrderOptions) {
		o.Params = params
	})
}

// SendOrder dispatches an order on a contract. order is the order payload (an
// order name or a structured object).
func (c *Client) SendOrder(contractID int, order any, opts ...SendOrderOption) error {
	type Request struct {
		api.TokenAndContractRequest
		Order any `json:"order"`
		sendOrderOptions
	}
	request := Request{
		TokenAndContractRequest: c.tokenAndContractRequest(contractID),
		Order:                   order,
		sendOrderOptions:        newSendOrderOptions(opts...),
	}
	reqURL := c.urlAppendingPath("/api/agents/order")
	return net.MakeRequestWithEmptyResponse(reqURL, request)
}
