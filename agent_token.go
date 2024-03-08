package maigo

type AgentToken struct {
	Token   string `json:"agent_token"`
	Patient string `json:"patient_agent_token"`
	Doctor  string `json:"doctor_agent_token"`
}
