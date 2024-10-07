package maigo

import (
	"time"

	"github.com/TikhonP/maigo/internal/api"
	"github.com/TikhonP/maigo/internal/json"
)

type getRecordsOptions struct {
	api.TokenAndContractRequest
	CategoryName string         `json:"category_name,omitempty"`
	Limit        int            `json:"limit,omitempty"`
	Offset       int            `json:"offset,omitempty"`
	From         json.Timestamp `json:"from,omitempty"`
	To           json.Timestamp `json:"to,omitempty"`
}

func applyGetRecordsOptions(opts *getRecordsOptions, options ...GetRecordsOption) {
	for _, option := range options {
		option.apply(opts)
	}
}

type GetRecordsOption interface {
	apply(*getRecordsOptions)
}

// funcGetRecordsOption wraps a function that modifies getRecordsOptions into an
// implementation of the GetRecordsOption interface.
type funcGetRecordsOption struct {
	f func(*getRecordsOptions)
}

func (fmo *funcGetRecordsOption) apply(do *getRecordsOptions) {
	fmo.f(do)
}

func newFuncGetRecordsOption(f func(*getRecordsOptions)) *funcGetRecordsOption {
	return &funcGetRecordsOption{
		f: f,
	}
}

// WithCategoryName is an option for GetRecords that specifies the category name
// of the records to retrieve.
// You can specify multiple category names by calling this function multiple times.
func WithCategoryName(categoryName string) GetRecordsOption {
	return newFuncGetRecordsOption(func(o *getRecordsOptions) {
		if len(o.CategoryName) > 0 {
			o.CategoryName += ","
		}
		o.CategoryName += categoryName
	})
}

// Limit is an option for GetRecords that specifies the maximum number of records.
func Limit(limit int) GetRecordsOption {
	return newFuncGetRecordsOption(func(o *getRecordsOptions) {
		o.Limit = limit
	})
}

// Offset is an option for GetRecords that specifies the number of records to skip.
func Offset(offset int) GetRecordsOption {
	return newFuncGetRecordsOption(func(o *getRecordsOptions) {
		o.Offset = offset
	})
}

// FromTime is an option for GetRecords that specifies the start time of the records to retrieve.
func FromTime(from time.Time) GetRecordsOption {
	return newFuncGetRecordsOption(func(o *getRecordsOptions) {
		o.From = json.Timestamp{Time: from}
	})
}

// ToTime is an option for GetRecords that specifies the end time of the records to retrieve.
func ToTime(to time.Time) GetRecordsOption {
	return newFuncGetRecordsOption(func(o *getRecordsOptions) {
		o.To = json.Timestamp{Time: to}
	})
}

