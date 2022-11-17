package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/maragudk/errors"
)

type Time struct {
	T time.Time
}

// rfc3339Milli is like time.RFC3339Nano, but with millisecond precision, and fractional seconds do not have trailing
// zeros removed.
const rfc3339Milli = "2006-01-02T15:04:05.000Z07:00"

// Value satisfies driver.Valuer interface.
func (t Time) Value() (driver.Value, error) {
	return t.T.UTC().Format(rfc3339Milli), nil
}

// Scan satisfies sql.Scanner interface.
func (t *Time) Scan(src any) error {
	if src == nil {
		return nil
	}

	s, ok := src.(string)
	if !ok {
		return errors.Newf("error scanning time, got %+v", src)
	}

	parsedT, err := time.Parse(rfc3339Milli, s)
	if err != nil {
		return err
	}

	t.T = parsedT.UTC()

	return nil
}

type Job struct {
	ID       int
	Name     string
	Payload  Map
	Timeout  time.Duration
	Run      Time
	Received *Time
	Created  Time
	Updated  Time
}

type Map map[string]string

// Value satisfies driver.Valuer interface.
func (m Map) Value() (driver.Value, error) {
	data, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return string(data), nil
}

// Scan satisfies sql.Scanner interface.
func (m *Map) Scan(src any) error {
	if src == nil {
		return nil
	}

	s, ok := src.(string)
	if !ok {
		return errors.Newf("error scanning string, got %+v", src)
	}

	return json.Unmarshal([]byte(s), m)
}
