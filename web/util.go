package web

import (
	"net/http"
)

func notAcceptable(w http.ResponseWriter) {
	errorResponse(w, Error{
		Status:  http.StatusNotAcceptable, // 406 Not Acceptable
		Details: "Request MUST accept \"application/json\"",
	})
}

func unsupportedMediaType(w http.ResponseWriter) {
	errorResponse(w, Error{
		Status:  http.StatusUnsupportedMediaType, // 415 Unsupported Media Type
		Details: "Request body MUST be \"application/json\"",
	})
}
