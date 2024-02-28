package api

// clinic describes Id and Name information about a clinic.
type clinic struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type Clinics []clinic
