package maigo

import "github.com/TikhonP/maigo/internal/api"

type getRecordsOptions struct {
	api.TokenAndContractRequest
	CategoryName string `json:"category_name,omitempty"`
}
