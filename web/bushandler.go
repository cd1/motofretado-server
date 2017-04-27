package web

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/julienschmidt/httprouter"
	"golang.org/x/net/webdav"
	"motorola.com/cdeives/motofretado/data"
	"motorola.com/cdeives/motofretado/model"
)

// BusHandler handles the HTTP requests on the bus resource. It is responsible
// for listing detailed information, updating and deleting individual buses.
type BusHandler struct {
	DB data.DB
}

func (h BusHandler) delete(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
	if req.Header.Get("Accept") != "application/json" {
		notAcceptable(w) // 406 Not Acceptable

		return
	}

	id := params.ByName("id")
	if len(id) == 0 {
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

func (h BusHandler) get(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
	if req.Header.Get("Accept") != "application/json" {
		notAcceptable(w) // 406 Not Acceptable

		return
	}

	id := params.ByName("id")
	if len(id) == 0 {
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

func (h BusHandler) patch(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
	if req.Header.Get("Accept") != "application/json" {
		notAcceptable(w) // 406 Not Acceptable

		return
	}

	if req.Header.Get("Content-Type") != "application/json" {
		unsupportedMediaType(w) // 415 Unsupported Media Type

		return
	}

	id := params.ByName("id")
	if len(id) == 0 {
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

	updatedBus, err := h.DB.UpdateBus(id, bus)
	if err != nil {
		switch err.(type) {
		case data.InvalidParameterError, data.MissingParameterError:
			errorResponse(w, Error{
				Status:  webdav.StatusUnprocessableEntity, // 422 Unprocessable Entity
				Details: err.Error(),
			})
		default:
			errorResponse(w, Error{
				Status:  http.StatusInternalServerError, // 500 Internal Server Error
				Details: err.Error(),
			})
		}

		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(updatedBus); err != nil { // 200 OK
		logrus.WithError(err).Error("could not encode bus to JSON")
	}
}
