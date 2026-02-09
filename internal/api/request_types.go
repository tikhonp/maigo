// Package api provides request and response types for Medsenger API.
package api

type TokenOnlyRequest struct {
	APIKey string `json:"api_key"`
}

type TokenAndContractRequest struct {
	TokenOnlyRequest
	ContractID int `json:"contract_id"`
}
