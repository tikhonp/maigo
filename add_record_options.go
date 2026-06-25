package maigo

import (
	"time"

	"github.com/tikhonp/maigo/internal/json"
)

type addRecordOptions struct {
	Time      *json.Timestamp `json:"time,omitempty"`
	Params    map[string]any  `json:"params,omitempty"`
	Files     []Attachment    `json:"files,omitempty"`
	Replace   bool            `json:"replace"`
	SyncFiles bool            `json:"sync_files,omitempty"`
}

func newAddRecordOptions(opts ...AddRecordOption) addRecordOptions {
	o := addRecordOptions{}
	for _, opt := range opts {
		opt.apply(&o)
	}
	return o
}

type AddRecordOption interface {
	apply(*addRecordOptions)
}

// funcAddRecordOption wraps a function that modifies addRecordOptions into an
// implementation of the AddRecordOption interface.
type funcAddRecordOption struct {
	f func(*addRecordOptions)
}

func (fo *funcAddRecordOption) apply(o *addRecordOptions) {
	fo.f(o)
}

func newFuncAddRecordOption(f func(*addRecordOptions)) *funcAddRecordOption {
	return &funcAddRecordOption{f: f}
}

// WithRecordTime returns an AddRecordOption which sets the record timestamp.
// By default the server records the current time.
func WithRecordTime(t time.Time) AddRecordOption {
	return newFuncAddRecordOption(func(o *addRecordOptions) {
		o.Time = &json.Timestamp{Time: t}
	})
}

// WithRecordParams returns an AddRecordOption which attaches arbitrary
// parameters to the record (for example {"record_classifier": "A"}).
func WithRecordParams(params map[string]any) AddRecordOption {
	return newFuncAddRecordOption(func(o *addRecordOptions) {
		o.Params = params
	})
}

// WithRecordFiles returns an AddRecordOption which attaches files to the record.
func WithRecordFiles(files ...Attachment) AddRecordOption {
	return newFuncAddRecordOption(func(o *addRecordOptions) {
		o.Files = files
	})
}

// WithReplace returns an AddRecordOption which replaces the previous record in
// the same category (and matching classifier params) instead of appending.
func WithReplace() AddRecordOption {
	return newFuncAddRecordOption(func(o *addRecordOptions) {
		o.Replace = true
	})
}

// WithSyncFiles returns an AddRecordOption which makes the server process
// attached files synchronously before responding.
func WithSyncFiles() AddRecordOption {
	return newFuncAddRecordOption(func(o *addRecordOptions) {
		o.SyncFiles = true
	})
}
