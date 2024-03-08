package maigo

type Category struct {
	Id                    int    `json:"id"`          // Category unique identifier.
	Name                  string `json:"name"`        // Category name.
	Description           string `json:"description"` // Category name.
	Unit                  string `json:"unit"`
	Type                  string `json:"type"`
	DefaultRepresentation string `json:"default_representation"`
	IsLegacy              bool   `json:"is_legacy"`
	Subcategory           string `json:"subcategory"`
	DoctorCanAdd          bool   `json:"doctor_can_add"`
	DoctorCanReplace      bool   `json:"doctor_can_replace"`
}

type Categories []Category
