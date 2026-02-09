package maigo

import "github.com/golang-jwt/jwt/v5"

// RequestRole defines the role of the backand making the request to the agent.
type RequestRole string

const (
	// System requests (messages, orders, hooks, initialization)
	RequestRoleSystem RequestRole = "system"
	// Requests from patients
	RequestRolePatient RequestRole = "patient"
	// Requests from doctors
	RequestRoleDoctor RequestRole = "doctor"
	// Requests from clinic registrars
	RequestRoleRegistrator RequestRole = "registrator"
	// Requests from medical supervisors
	RequestRoleSupervisor RequestRole = "supervisor"
)

type JWTClaims struct {
	jwt.RegisteredClaims

	// ContractID may be null for system requests, so we use a pointer to string to allow nil values.
	ContractID *string `json:"contract_id,omitempty"`
	// AgentID is the ID of your agent.
	AgentID string `json:"agent_id"`
	// Roles can be system, patient, doctor, registrator, supervisor.
	Roles []RequestRole `json:"roles"`
	// Type is the type of token, e.g., "agent_access".
	Type string `json:"type"`
}

func decodeAgentJWT(tokenString, apiKey string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&JWTClaims{},
		func(token *jwt.Token) (any, error) {
			return []byte(apiKey), nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, jwt.ErrTokenInvalidClaims
	}
}
