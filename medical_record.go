package maigo

import "time"

type MedicalRecordSource struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// MedicalRecordFile describes a file attached to a medical record.
type MedicalRecordFile struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type MedicalRecord struct {
	ID            int                 `json:"id"`
	Value         any                 `json:"value"`
	Additions     []any               `json:"additions"`
	Source        MedicalRecordSource `json:"source"`
	Category      Category            `json:"category_info"`
	Group         string              `json:"group,omitempty"`
	Params        any                 `json:"params,omitempty"`
	AttachedFiles []MedicalRecordFile `json:"attached_files,omitempty"`

	// Time and Uploaded are populated only by the gRPC transport (record
	// creation and last-update time). The REST transport leaves them as the
	// zero time.
	Time     time.Time `json:"-"`
	Uploaded time.Time `json:"-"`
}
