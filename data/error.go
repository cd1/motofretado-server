package data

import "fmt"

// InvalidParameterError represents an error due to an invalid parameter
// being specified.
type InvalidParameterError struct {
	Name  string
	Value interface{}
}

// MissingParameterError represents an error due to a missing parameter.
type MissingParameterError struct {
	Name string
}

// Error returns a string representation of the error.
func (e InvalidParameterError) Error() string {
	return fmt.Sprintf("invalid parameter \"%v\" = \"%v\"", e.Name, e.Value)
}

// Error returns a string representation of the error.
func (e MissingParameterError) Error() string {
	return fmt.Sprintf("missing parameter \"%v\"", e.Name)
}
