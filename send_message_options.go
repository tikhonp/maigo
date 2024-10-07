package maigo

import (
	"time"

	"github.com/TikhonP/maigo/internal/json"
)

type sendMessageOptions struct {
	Text            string              `json:"text"`
	ForwardToDoctor bool                `json:"forward_to_doctor"`
	ActionLink      string              `json:"action_link,omitempty"`
	SendFrom        UserRole            `json:"send_from,omitempty"`
	ActionName      string              `json:"action_name,omitempty"`
	ActionOneTime   bool                `json:"action_onetime"`
	ActionBig       bool                `json:"action_big"`
	ActionType      MessageActionType   `json:"action_type"`
	OnlyDoctor      bool                `json:"only_doctor"`
	NeedAnswer      bool                `json:"need_answer"`
	OnlyPatient     bool                `json:"only_patient"`
	ActionDeadline  *json.Timestamp     `json:"action_deadline,omitempty"`
	IsUrgent        bool                `json:"is_urgent"`
	Attachments     []MessageAttachment `json:"attachments,omitempty"`
}

func newSendMessageOptions(text string, opts ...SendMessageOption) *sendMessageOptions {
	smo := &sendMessageOptions{
		Text:            text,
		ForwardToDoctor: true,
		ActionOneTime:   true,
		OnlyDoctor:      false,
		OnlyPatient:     false,
		IsUrgent:        false,
		NeedAnswer:      false,
		ActionBig:       true,
		ActionType:      Action,
	}
	for _, opt := range opts {
		opt.apply(smo)
	}
	return smo
}

type UserRole string

const (
	Patient UserRole = "patient"
	Doctor  UserRole = "doctor"
)

// MessageActionType represents types of showing action behaviour.
type MessageActionType string

const (
	Action    MessageActionType = "action"  // Open action in iFrame ot WebView.
	UrlAction MessageActionType = "url"     // Open action as outside url.
	AppUrl    MessageActionType = "app_url" // Open action as outside url that shows only in mobile app.
)

type MessageAttachment struct {
}

type SendMessageOption interface {
	apply(*sendMessageOptions)
}

// funcSendMessageOption wraps a function that modifies sendMessageOptions into an
// implementation of the SendMessageOption interface.
type funcSendMessageOption struct {
	f func(*sendMessageOptions)
}

func (fmo *funcSendMessageOption) apply(do *sendMessageOptions) {
	fmo.f(do)
}

func newFuncSendMessageOption(f func(*sendMessageOptions)) *funcSendMessageOption {
	return &funcSendMessageOption{
		f: f,
	}
}

// WithAction returns a SendMessageOption which sets action button to a message
// with name and link title.
func WithAction(name string, link string, actionType MessageActionType) SendMessageOption {
	return newFuncSendMessageOption(func(o *sendMessageOptions) {
		o.ActionName = name
		o.ActionLink = link
		o.ActionType = actionType
	})
}

// WithReusableAction returns a SendMessageOption which sets message deletion behaviour
// to not remove message after first success use of an action.
func WithReusableAction() SendMessageOption {
	return newFuncSendMessageOption(func(o *sendMessageOptions) {
		o.ActionOneTime = false
	})
}

// WithSmallAction return a SendMessageOption which sets action window to small.
// Actually no one using it now.
func WithSmallAction() SendMessageOption {
	return newFuncSendMessageOption(func(o *sendMessageOptions) {
		o.ActionBig = false
	})
}

// WithActionDeadline returns a SendMessageOption which sets date when message becomes inactive.
func WithActionDeadline(t time.Time) SendMessageOption {
	return newFuncSendMessageOption(func(o *sendMessageOptions) {
		o.ActionDeadline = &json.Timestamp{Time: t}
	})
}

// OnlyDoctor returns a SendMessageOption which sets message to be shown only to doctor.
func OnlyDoctor() SendMessageOption {
	return newFuncSendMessageOption(func(o *sendMessageOptions) {
		o.OnlyDoctor = true
	})
}

// OnlyPatient returns a SendMessageOption which sets message to be shown only to patient.
func OnlyPatient() SendMessageOption {
	return newFuncSendMessageOption(func(o *sendMessageOptions) {
		o.OnlyPatient = true
	})
}

// MarkMessagesAnsweredForDoctor returns a SendMessageOption which sets all messages from patient as answered.
func MarkMessagesAnsweredForDoctor() SendMessageOption {
	return newFuncSendMessageOption(func(o *sendMessageOptions) {
		o.ForwardToDoctor = false
	})
}

// NeedAnswer returns a SendMessageOption which forces doctor to answer this message in a chat.
func NeedAnswer() SendMessageOption {
	return newFuncSendMessageOption(func(o *sendMessageOptions) {
		o.NeedAnswer = true
	})
}

// Urgent returns a SendMessageOption which set message to urgent state (colors it to red usually).
func Urgent() SendMessageOption {
	return newFuncSendMessageOption(func(o *sendMessageOptions) {
		o.IsUrgent = true
	})
}

// WithPatientSenderRole returns a SendMessageOption which set sender role to patient (default id doctor).
func WithPatientSenderRole() SendMessageOption {
	return newFuncSendMessageOption(func(o *sendMessageOptions) {
		o.SendFrom = Patient
	})
}

// WithAttachments returns a SendMessageOption which sets attachments to a message.
func WithAttachments(a []MessageAttachment) SendMessageOption {
	return newFuncSendMessageOption(func(o *sendMessageOptions) {
		o.Attachments = a
	})
}

