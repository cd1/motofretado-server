package web

import (
	"fmt"
	"net/http"
)

func methodNotAllowed(w http.ResponseWriter, method string, allowedMethods string) {
	w.Header().Set("Allow", allowedMethods)
	errorResponse(w, Error{
		Status:  http.StatusMethodNotAllowed, // 405 Method Not Allowed
		Details: fmt.Sprintf("HTTP method \"%v\" isn't allowed for this resource", method),
	})
}

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
