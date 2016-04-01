package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"golang.org/x/net/webdav"
	"motorola.com/cdeives/motofretado/data"
	"motorola.com/cdeives/motofretado/model"
)

const busHandlerAllowedMethods = "GET, HEAD, OPTIONS, PATCH"

// BusHandler handles the HTTP requests on the bus resource. It is responsible
// for listing detailed information, updating and deleting individual buses.
type BusHandler struct {
	DB data.DB
}

func (h BusHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "DELETE":
		h.delete(w, req)
	case "GET", "HEAD":
		h.get(w, req)
	case "OPTIONS":
		h.options(w)
	case "PATCH":
		h.patch(w, req)
	default:
		methodNotAllowed(w, req.Method, busHandlerAllowedMethods) // 405 Method Not Allowed
	}
}

func (h BusHandler) delete(w http.ResponseWriter, req *http.Request) {
	if req.Header.Get("Accept") != "application/json" {
		notAcceptable(w) // 406 Not Acceptable

		return
	}

	id := strings.TrimPrefix(req.URL.Path, "/bus/")
	if id == "" {
		errorResponse(w, Error{
			Status:  http.StatusBadRequest, // 400 Bad Request
			Details: "Empty bus ID",
		})

		return
	}

	exists, err := h.DB.ExistsBus(id)
	if err != nil {
		errorResponse(w, Error{
			Status:  http.StatusInternalServerError, // 500 Internal Server Error
			Details: err.Error(),
		})

		return
	}

	if !exists {
		errorResponse(w, Error{
			Status:  http.StatusNotFound, // 404 Not Found
			Details: fmt.Sprintf("Bus \"%v\" doesn't exist", id),
		})

		return
	}

	if err = h.DB.DeleteBus(id); err != nil {
		errorResponse(w, Error{
			Status:  http.StatusInternalServerError, // 500 Internal Server Error
			Details: err.Error(),
		})

		return
	}

	w.WriteHeader(http.StatusNoContent) // 204 No Content
}

func (h BusHandler) get(w http.ResponseWriter, req *http.Request) {
	if req.Header.Get("Accept") != "application/json" {
		notAcceptable(w) // 406 Not Acceptable

		return
	}

	id := strings.TrimPrefix(req.URL.Path, "/bus/")
	if id == "" {
		errorResponse(w, Error{
			Status:  http.StatusBadRequest, // 400 Bad Request
			Details: "Empty bus ID",
		})

		return
	}

	exists, err := h.DB.ExistsBus(id)
	if err != nil {
		errorResponse(w, Error{
			Status:  http.StatusInternalServerError, // 500 Internal Server Error
			Details: err.Error(),
		})

		return
	}

	if !exists {
		errorResponse(w, Error{
			Status:  http.StatusNotFound, // 404 Not Found
			Details: fmt.Sprintf("Bus \"%v\" doesn't exist", id),
		})

		return
	}

	bus, err := h.DB.ReadBus(id)
	if err != nil {
		errorResponse(w, Error{
			Status:  http.StatusInternalServerError, // 500 Internal Server Error
			Details: err.Error(),
		})

		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(bus); err != nil { // 200 OK
		logrus.WithError(err).Error("could not encode bus to JSON")
	}
}

func (BusHandler) options(w http.ResponseWriter) {
	w.Header().Set("Allow", busHandlerAllowedMethods)
	w.WriteHeader(http.StatusNoContent) // 204 No Content
}

func (h BusHandler) patch(w http.ResponseWriter, req *http.Request) {
	if req.Header.Get("Accept") != "application/json" {
		notAcceptable(w) // 406 Not Acceptable

		return
	}

	if req.Header.Get("Content-Type") != "application/json" {
		unsupportedMediaType(w) // 415 Unsupported Media Type

		return
	}

	id := strings.TrimPrefix(req.URL.Path, "/bus/")
	if id == "" {
		errorResponse(w, Error{
			Status:  http.StatusBadRequest, // 400 Bad Request
			Details: "Empty bus ID",
		})

		return
	}

	exists, err := h.DB.ExistsBus(id)
	if err != nil {
		errorResponse(w, Error{
			Status:  http.StatusInternalServerError, // 500 Internal Server Error
			Details: err.Error(),
		})

		return
	}

	if !exists {
		errorResponse(w, Error{
			Status:  http.StatusNotFound, // 404 Not Found
			Details: fmt.Sprintf("Bus \"%v\" doesn't exist", id),
		})

		return
	}

	var bus model.Bus

	if err := json.NewDecoder(req.Body).Decode(&bus); err != nil {
		errorResponse(w, Error{
			Status:  http.StatusBadRequest, // 400 Bad Request
			Details: err.Error(),
		})

		return
	}

	if bus.ID != "" {
		errorResponse(w, Error{
			Status:  webdav.StatusUnprocessableEntity, // 422 Unprocessable Entity
			Details: fmt.Sprintf("Bus ID [%v] cannot be updated", bus.ID),
		})

		return
	}

	if !bus.CreatedAt.IsZero() {
		errorResponse(w, Error{
			Status:  webdav.StatusUnprocessableEntity, // 422 Unprocessable Entity
			Details: fmt.Sprintf("Bus create time [%v] cannot be updated", bus.CreatedAt),
		})

		return
	}

	now := time.Now()
	if bus.UpdatedAt.IsZero() {
		bus.UpdatedAt = now
	} else if bus.UpdatedAt.After(now) {
		errorResponse(w, Error{
			Status:  webdav.StatusUnprocessableEntity, // 422 Unprocessable Entity
			Details: fmt.Sprintf("Bus update time [%v] cannot be in the future", bus.UpdatedAt),
		})

		return
	}

	existingBus, err := h.DB.ReadBus(bus.ID)
	if err != nil {
		errorResponse(w, Error{
			Status:  http.StatusInternalServerError, // 500 Internal Server Error
			Details: err.Error(),
		})

		return
	}

	if bus.UpdatedAt.Before(existingBus.UpdatedAt) {
		errorResponse(w, Error{
			Status:  webdav.StatusUnprocessableEntity, // 422 Unprocessable Entity
			Details: fmt.Sprintf("Bus update time [%v] cannot be before last update time [%v]", bus.UpdatedAt, existingBus.UpdatedAt),
		})

		return
	}

	updatedBus, err := h.DB.UpdateBus(id, bus)
	if err != nil {
		errorResponse(w, Error{
			Status:  http.StatusInternalServerError, // 500 Internal Server Error
			Details: err.Error(),
		})

		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(updatedBus); err != nil { // 200 OK
		logrus.WithError(err).Error("could not encode bus to JSON")
	}
}
