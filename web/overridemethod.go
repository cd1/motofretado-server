package web

import (
	"net/http"

	"github.com/Sirupsen/logrus"
)

// OverrideMethodWrapper changes req.Method to the value of the header
// X-HTTP-Method-Override if that header exists and the original request
// is POST.
func OverrideMethodHandler(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	if req.Method == http.MethodPost {
		if newMethod := req.Header.Get("X-HTTP-Method-Override"); len(newMethod) > 0 {
			req.Method = newMethod
			logrus.WithFields(logrus.Fields{
				"method": newMethod,
			}).Debug("HTTP method overridden")
		}
	}

	next(w, req)
}
