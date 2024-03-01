package api

type TokenOnlyRequest struct {
	ApiKey string `json:"api_key"`
}

type TokenAndContractRequest struct {
	TokenOnlyRequest
	ContractId int `json:"contract_id"`
}
