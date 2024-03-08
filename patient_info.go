package maigo

import "github.com/TikhonP/maigo/internal/json"

type Sex string

const (
	Male   Sex = "male"
	Female Sex = "female"
)

// DoctorHelper describes helper for doctor attached to the contract.
type DoctorHelper struct {
	Id     int    `json:"id"`      // DoctorHelper unique identifier.
	Name   string `json:"name"`    // User name.
	Role   string `json:"role"`    // Role assigned to helper.
	UserId int    `json:"user_id"` // Id of the user.
}

// Scenario describes scenario that connected to a contract.
type Scenario struct {
	Id       int    `json:"id"`       // Scenario unique identifier.
	Name     string `json:"name"`     // Scenario name.
	Category string `json:"category"` // Scenario category tag.

	// Template for contract conclusion. With statistic about concluded contract.
	//
	// Agent questionnaires and medicines uses it.
	ConclusionTemplate string `json:"conclusion_template"`
}

// ContractInfo describes contract including some patient and doctor info.
type ContractInfo struct {

	// Contract information
	Id             int            `json:"id"`              // Contract unique identifier.
	ContractNumber string         `json:"contract_number"` // Contract identifier assigned to contract by clinic.
	Scenario       Scenario       `json:"scenario"`        // Scenario assigned to the contract.
	StartDate      json.Timestamp `json:"start_timestamp"` // Contract start date.
	EndDate        json.Timestamp `json:"end_timestamp"`   // Contract end date.
	DaysDuration   int            `json:"days"`            // Contract duration in days.
	IsArchived     bool           `json:"archive"`         // Is contracted archived.

	// Patient information
	PatientName           string          `json:"name"`            // Patient's name.
	PatientEmail          string          `json:"email"`           // Patient's email.
	PatientSex            Sex             `json:"sex"`             // Patient's sex.
	PatientPhone          string          `json:"phone"`           // Patient's phone.
	PatientBirthday       json.StringDate `json:"birthday"`        // Patient's birthday.
	PatientTimezoneOffset int             `json:"timezone_offset"` // Patient's timezone offset from GMT.
	PatientAge            string          `json:"age"`             // Patient's age.

	// Doctor information
	DoctorName           string         `json:"doctor_name"`            // Doctor's name.
	DoctorPhone          string         `json:"doctor_phone"`           // Doctor's phone.
	DoctorId             int            `json:"doctor_id"`              // Doctor's unique identifier.
	DoctorUserId         int            `json:"doctor_user_id"`         // Doctor's user ID.
	DoctorHelpers        []DoctorHelper `json:"doctor_helpers"`         // Helpers for doctors assigned to an contract.
	DoctorTimezoneOffset int            `json:"doctor_timezone_offset"` // Doctor's timezone offset.

	// Clinic information
	ClinicId       int    `json:"clinic_id"`   // Clinic's unique identifier.
	ClinicName     string `json:"clinic_name"` // Clinic's name.
	ClinicTimezone string `json:"timezone"`    // Clinic's timezone.

}
