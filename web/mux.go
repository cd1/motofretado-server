package web

import (
	"fmt"
	"net/http"
	"os"
	"runtime/debug"
	"strconv"

	"github.com/Sirupsen/logrus"
	"github.com/cd1/motofretado-server/data"
	"github.com/cd1/motofretado-server/web/jsonapi"
	"github.com/gorilla/handlers"
	"github.com/julienschmidt/httprouter"
	"github.com/urfave/negroni"
)

var logOutput = os.Stderr

// BuildMux builds the HTTP mux for the web server. It is responsible for
// creating and chaining all available HTTP handlers.
func BuildMux(repo *data.Repository) http.Handler {
	router := httprouter.New()

	logrus.WithFields(logrus.Fields{
		"path": "/bus",
	}).Debug("registering HTTP handler")
	buses := BusesHandler{repo: repo}
	router.GET("/bus", buses.get)
	router.HEAD("/bus", buses.get)
	router.POST("/bus", buses.post)

	logrus.WithFields(logrus.Fields{
		"path": "/bus/:id",
	}).Debug("registering HTTP handler")
	bus := BusHandler{repo: repo}
	router.GET("/bus/:id", bus.get)
	router.HEAD("/bus/:id", bus.get)
	router.PATCH("/bus/:id", bus.patch)
	router.DELETE("/bus/:id", bus.doDelete)

	router.MethodNotAllowed = http.HandlerFunc(methodNotAllowed)
	router.NotFound = http.HandlerFunc(notFound)
	router.PanicHandler = panicRecovery

	n := negroni.New()
	n.UseFunc(func(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
		handlers.LoggingHandler(logOutput, next).ServeHTTP(w, req)
	})
	n.UseFunc(func(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
		handlers.HTTPMethodOverrideHandler(next).ServeHTTP(w, req)
	})
	n.UseFunc(func(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
		handlers.CompressHandler(next).ServeHTTP(w, req)
	})
	n.UseHandler(router)

	return n
}

func methodNotAllowed(w http.ResponseWriter, req *http.Request) {
	errorResponse(w, jsonapi.ErrorData{
		Status: strconv.Itoa(http.StatusMethodNotAllowed),
		Title:  "HTTP method not allowed",
		Detail: req.Method,
	})
}

func notFound(w http.ResponseWriter, req *http.Request) {
	errorResponse(w, jsonapi.ErrorData{
		Status: strconv.Itoa(http.StatusNotFound),
		Title:  "URL not found",
		Detail: req.URL.Path,
	})
}

func panicRecovery(w http.ResponseWriter, _ *http.Request, value interface{}) {
	stackTrace := debug.Stack()
	fmt.Fprintf(logOutput, "%s", stackTrace)

	errorResponse(w, jsonapi.ErrorData{
		Status: strconv.Itoa(http.StatusInternalServerError),
		Title:  "Unrecoverable error",
		Detail: fmt.Sprintf("PANIC: %s", value),
	})
}
