package jsonapi

import "github.com/hashicorp/go-version"

const (
	ContentType    = "application/vnd.api+json"
	CurrentVersion = "1.0"
)

var currentVersionStruct = version.Must(version.NewVersion(CurrentVersion))

type Root struct {
	Version string `json:"version"`
}

type Links struct {
	Self string `json:"self"`
}
