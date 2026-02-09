package maigo

type MedicalRecordSource struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type MedicalRecord struct {
	ID        int                 `json:"id"`
	Value     any                 `json:"value"`
	Additions []any               `json:"additions"`
	Source    MedicalRecordSource `json:"source"`
	Category  Category            `json:"category_info"`
}
