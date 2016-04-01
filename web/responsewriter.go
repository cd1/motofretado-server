package web

import (
	"net/http"
)

// ResponseWriter is a custom http.ResponseWriter implementation which keeps
// track of the status code and the number of bytes written to the body. It is
// needed by the function "LogWrapper" so it can have its information correctly.
type ResponseWriter struct {
	W           http.ResponseWriter
	contentSize int64
	statusCode  int64
	wroteHeader bool
}

// Header returns this response's HTTP headers .
func (w ResponseWriter) Header() http.Header {
	return w.W.Header()
}

// Write writes to this response and counts the number of bytes written. If the
// status code hasn't been written yet, 200 OK is set automatically.
func (w *ResponseWriter) Write(p []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}

	s, err := w.W.Write(p)
	if err != nil {
		return 0, err
	}

	w.contentSize += int64(s)

	return s, nil
}

// WriteHeader sets the response's status code.
func (w *ResponseWriter) WriteHeader(status int) {
	w.W.WriteHeader(status)

	w.wroteHeader = true
	w.statusCode = int64(status)
}

// StatusCode returns the status code set by WriteHeader.
func (w ResponseWriter) StatusCode() int64 {
	return w.statusCode
}

// ContentSize returns the body size written to the response.
func (w ResponseWriter) ContentSize() int64 {
	return w.contentSize
}
