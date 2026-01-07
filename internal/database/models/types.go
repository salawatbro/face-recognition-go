package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// Metadata is a custom type for storing JSON metadata
type Metadata map[string]interface{}

// Scan implements sql.Scanner interface
func (m *Metadata) Scan(value interface{}) error {
	if value == nil {
		*m = make(map[string]interface{})
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("invalid type for Metadata")
	}

	if len(bytes) == 0 {
		*m = make(map[string]interface{})
		return nil
	}

	return json.Unmarshal(bytes, m)
}

// Value implements driver.Valuer interface
func (m Metadata) Value() (driver.Value, error) {
	if m == nil {
		return "{}", nil
	}
	return json.Marshal(m)
}

// Embedding is a custom type for storing float32 arrays as JSON
type Embedding []float32

// Scan implements sql.Scanner interface
func (e *Embedding) Scan(value interface{}) error {
	if value == nil {
		*e = []float32{}
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("invalid type for Embedding")
	}

	if len(bytes) == 0 {
		*e = []float32{}
		return nil
	}

	return json.Unmarshal(bytes, e)
}

// Value implements driver.Valuer interface
func (e Embedding) Value() (driver.Value, error) {
	if e == nil {
		return "[]", nil
	}
	return json.Marshal(e)
}
