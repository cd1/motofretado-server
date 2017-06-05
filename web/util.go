package web

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/cd1/motofretado-server/web/jsonapi"
)

func notAcceptable(w http.ResponseWriter) {
	errorResponse(w, jsonapi.ErrorData{
		Status: strconv.Itoa(http.StatusNotAcceptable), // 406 Not Acceptable
		Title:  "HTTP method not acceptable",
		Detail: fmt.Sprintf("Request MUST accept \"%v\"", jsonapi.ContentType),
	})
}

func unsupportedMediaType(w http.ResponseWriter) {
	errorResponse(w, jsonapi.ErrorData{
		Status: strconv.Itoa(http.StatusUnsupportedMediaType), // 415 Unsupported Media Type
		Title:  "HTTP request content type not supported",
		Detail: fmt.Sprintf("Request body MUST be \"%v\"", jsonapi.ContentType),
	})
}

func requestScheme(req *http.Request) string {
	var scheme string

	if req.TLS == nil {
		scheme = "http"
	} else {
		scheme = "https"
	}

	return scheme
}
