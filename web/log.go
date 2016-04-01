package web

import (
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
)

// LogWrapper logs an HTTP handler (at INFO level) using the Common Log Format.
func LogWrapper(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		customW := ResponseWriter{W: w}

		start := time.Now()
		h.ServeHTTP(&customW, req)
		end := time.Now()

		logrus.Infof("%v - - [%v] \"%v %v %v\" %v %v \"%v\" \"%v\"",
			req.RemoteAddr, end.Sub(start), req.Method, req.URL.Path, req.Proto, customW.StatusCode(), customW.ContentSize(), req.Referer(), req.UserAgent())
	})
}
