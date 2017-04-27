package web

import (
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/julienschmidt/httprouter"
	"github.com/urfave/negroni"
	"motorola.com/cdeives/motofretado/data"
)

// BuildMux builds the HTTP mux for the web server. It is responsible for
// creating and chaining all available HTTP handlers.
func BuildMux(db data.DB) http.Handler {
	router := httprouter.New()

	logrus.Debug("registering HTTP handler /bus")
	buses := BusesHandler{DB: db}
	router.GET("/bus", buses.get)
	router.HEAD("/bus", buses.get)
	router.POST("/bus", buses.post)

	logrus.Debug("registering HTTP handler /bus/:id")
	bus := BusHandler{DB: db}
	router.GET("/bus/:id", bus.get)
	router.HEAD("/bus/:id", bus.get)
	router.PATCH("/bus/:id", bus.patch)
	router.DELETE("/bus/:id", bus.delete)

	n := negroni.Classic()
	n.UseFunc(OverrideMethodHandler)
	n.UseHandler(router)

	return n
}
