package web

import (
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/urfave/negroni"
	"motorola.com/cdeives/motofretado/data"
)

// BuildMux builds the HTTP mux for the web server. It is responsible for
// creating and chaining all available HTTP handlers.
func BuildMux(db data.DB) http.Handler {
	mux := http.NewServeMux()

	logrus.Debug("registering HTTP handler /bus")
	mux.Handle("/bus", BusesHandler{DB: db})
	logrus.Debug("registering HTTP handler /bus/:id")
	mux.Handle("/bus/", BusHandler{DB: db})

	n := negroni.Classic()
	n.UseFunc(OverrideMethodHandler)
	n.UseHandler(mux)

	return n
}
