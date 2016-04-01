package web

import (
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/justinas/alice"
	"motorola.com/cdeives/motofretado/data"
)

// BuildMux builds the HTTP mux for the web server. It is responsible for
// creating and chaining all available HTTP handlers.
func BuildMux(db data.DB) *http.ServeMux {
	chain := alice.New(OverrideMethodWrapper, LogWrapper, PanicWrapper)

	mux := http.NewServeMux()

	logrus.Debug("registering HTTP handler /bus")
	mux.Handle("/bus", chain.Then(BusesHandler{DB: db}))
	logrus.Debug("registering HTTP handler /bus/:id")
	mux.Handle("/bus/", chain.Then(BusHandler{DB: db}))

	return mux
}
