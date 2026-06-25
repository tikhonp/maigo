package maigo

import (
	"time"

	"github.com/tikhonp/maigo/internal/api"
	"github.com/tikhonp/maigo/internal/json"
	"github.com/tikhonp/maigo/internal/net"
)

type addTaskOptions struct {
	Number     int             `json:"number"`
	Important  bool            `json:"important"`
	Date       *json.Timestamp `json:"date,omitempty"`
	ActionLink string          `json:"action_link,omitempty"`
}

func newAddTaskOptions(opts ...AddTaskOption) addTaskOptions {
	o := addTaskOptions{Number: 1}
	for _, opt := range opts {
		opt.apply(&o)
	}
	return o
}

type AddTaskOption interface {
	apply(*addTaskOptions)
}

// funcAddTaskOption wraps a function that modifies addTaskOptions into an
// implementation of the AddTaskOption interface.
type funcAddTaskOption struct {
	f func(*addTaskOptions)
}

func (fo *funcAddTaskOption) apply(o *addTaskOptions) {
	fo.f(o)
}

func newFuncAddTaskOption(f func(*addTaskOptions)) *funcAddTaskOption {
	return &funcAddTaskOption{f: f}
}

// WithTaskTargetNumber returns an AddTaskOption which sets how many times the
// task must be completed (default 1).
func WithTaskTargetNumber(number int) AddTaskOption {
	return newFuncAddTaskOption(func(o *addTaskOptions) {
		o.Number = number
	})
}

// WithTaskImportant returns an AddTaskOption which marks the task as important.
func WithTaskImportant() AddTaskOption {
	return newFuncAddTaskOption(func(o *addTaskOptions) {
		o.Important = true
	})
}

// WithTaskDate returns an AddTaskOption which sets the task deadline.
func WithTaskDate(date time.Time) AddTaskOption {
	return newFuncAddTaskOption(func(o *addTaskOptions) {
		o.Date = &json.Timestamp{Time: date}
	})
}

// WithTaskActionLink returns an AddTaskOption which attaches an action link to
// the task.
func WithTaskActionLink(link string) AddTaskOption {
	return newFuncAddTaskOption(func(o *addTaskOptions) {
		o.ActionLink = link
	})
}

// AddTask creates a task for the contract.
func (c *Client) AddTask(contractID int, text string, opts ...AddTaskOption) error {
	type Request struct {
		api.TokenAndContractRequest
		Text string `json:"text"`
		addTaskOptions
	}
	request := Request{
		TokenAndContractRequest: c.tokenAndContractRequest(contractID),
		Text:                    text,
		addTaskOptions:          newAddTaskOptions(opts...),
	}
	reqURL := c.urlAppendingPath("/api/agents/tasks/add")
	return net.MakeRequestWithEmptyResponse(reqURL, request)
}

// FinishTask marks a task as done.
func (c *Client) FinishTask(contractID int, taskID int) error {
	type Request struct {
		api.TokenAndContractRequest
		TaskID int `json:"task_id"`
	}
	request := Request{
		TokenAndContractRequest: c.tokenAndContractRequest(contractID),
		TaskID:                  taskID,
	}
	reqURL := c.urlAppendingPath("/api/agents/tasks/done")
	return net.MakeRequestWithEmptyResponse(reqURL, request)
}

// DeleteTask removes a task.
func (c *Client) DeleteTask(contractID int, taskID int) error {
	type Request struct {
		api.TokenAndContractRequest
		TaskID int `json:"task_id"`
	}
	request := Request{
		TokenAndContractRequest: c.tokenAndContractRequest(contractID),
		TaskID:                  taskID,
	}
	reqURL := c.urlAppendingPath("/api/agents/tasks/delete")
	return net.MakeRequestWithEmptyResponse(reqURL, request)
}
