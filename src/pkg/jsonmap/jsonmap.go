package jsonmap

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// JSONMap represent a struct as map
type JSONMap map[string]interface{}

// Value impl of Valuer - to JSON marshal
func (p JSONMap) Value() (driver.Value, error) {
	j, err := json.Marshal(p)
	return j, err
}

// Scan impl of Scanner - from JSON unmarshal
func (p *JSONMap) Scan(src interface{}) error {
	source, ok := src.([]byte)
	if !ok {
		return errors.New("type assertion .([]byte) failed")
	}

	var i interface{}
	err := json.Unmarshal(source, &i)
	if err != nil {
		return err
	}

	*p, ok = i.(map[string]interface{})
	if !ok {
		return errors.New("type assertion .(map[string]interface{}) failed")
	}

	return nil
}
