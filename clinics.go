package maigo

// clinic describes Id and Name information about a clinic.
type clinic struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Clinics []clinic
