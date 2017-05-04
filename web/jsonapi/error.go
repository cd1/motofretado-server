package jsonapi

import "fmt"

type ErrorsDocument struct {
	JSONAPI *Root       `json:"jsonapi,omitempty"`
	Errors  []ErrorData `json:"errors"`
	Links   *Links      `json:"links,omitempty"`
}

type ErrorData struct {
	ID     string       `json:"id,omitempty"`
	Links  *ErrorLinks  `json:"links,omitempty"`
	Status string       `json:"status"`
	Code   string       `json:"code,omitempty"`
	Title  string       `json:"title,omitempty"`
	Detail string       `json:"detail,omitempty"`
	Source *ErrorSource `json:"source,omitempty"`
}

type ErrorLinks struct {
	About string `json:"about"`
}

type ErrorSource struct {
	Pointer   string `json:"pointer,omitempty"`
	Parameter string `json:"parameter,omitempty"`
}

type UnsupportedVersionError struct {
	Version        string
	CurrentVersion string
}

func (err UnsupportedVersionError) Error() string {
	return fmt.Sprintf("JSONAPI version %v cannot be greather than %v", err.Version, err.CurrentVersion)
}

type InvalidTypeError struct {
	Type         string
	ExpectedType string
}

func (err InvalidTypeError) Error() string {
	return fmt.Sprintf("expected JSONAPI data type \"%v\" but got \"%v\"", err.ExpectedType, err.Type)
}
