package maigo

type MedicalRecordSource struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type MedicalRecord struct {
	Id        int                 `json:"id"`
	Value     interface{}         `json:"value"`
	Additions []interface{}       `json:"additions"`
	Source    MedicalRecordSource `json:"source"`
	Category  Category            `json:"category_info"`
}

