package json

import (
	"errors"
	"time"
)

// StringDate is like time.Time, but knows how to unmarshal from JSON string like "dd.mm.YYYY" and
// marshal back into the same JSON representation.
type StringDate struct {
	time.Time
}

func (t *StringDate) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}
	// TODO(https://go.dev/issue/47353): Properly unescape a JSON string.
	if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
		return errors.New("Time.UnmarshalJSON: input is not a JSON string")
	}
	data = data[len(`"`) : len(data)-len(`"`)]

	const layout = "02.01.2006"
	var err error
	t.Time, err = time.Parse(layout, string(data))
	return err
}
