package web

import (
	"net/http"

	"github.com/Sirupsen/logrus"
)

// OverrideMethodWrapper changes req.Method to the value of the header
// X-HTTP-Method-Override if that header exists and the original request
// is POST.
func OverrideMethodWrapper(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method == "POST" {
			if newMethod := req.Header.Get("X-HTTP-Method-Override"); newMethod != "" {
				req.Method = newMethod
				logrus.WithFields(logrus.Fields{
					"method": newMethod,
				}).Debug("HTTP method overriden")
			}
		}

		h.ServeHTTP(w, req)
	})
}
