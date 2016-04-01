package web

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/Sirupsen/logrus"
)

// PanicWrapper catches all panic events and sends a 500 Internal Server Error
// response with the panic message and the current stack trace.
func PanicWrapper(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				logrus.WithFields(logrus.Fields{
					"message": fmt.Sprintf("%s", r),
					"stack":   string(debug.Stack()),
				}).Error("!!! PANIC !!!")
				errorResponse(w, Error{
					Status:  http.StatusInternalServerError,
					Details: fmt.Sprintf("Unexpected error [%v]", r),
				})
			}
		}()

		h.ServeHTTP(w, req)
	})
}
