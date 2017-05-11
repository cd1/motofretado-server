package data

import (
	"fmt"

	"github.com/pkg/errors"
)

// ErrNoSuchRow represents an error when some data entry could not be found.
var ErrNoSuchRow = errors.New("no such row")

// DuplicateError represents an error when an operation could not be performed
// because that row already exists.
type DuplicateError struct {
	ID string
}

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

// Error returns a string representation of the error.
func (e DuplicateError) Error() string {
	return fmt.Sprintf("row with ID=\"%v\" already exists", e.ID)
}
