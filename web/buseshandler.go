package web

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/julienschmidt/httprouter"
	"motorola.com/cdeives/motofretado/data"
	"motorola.com/cdeives/motofretado/model"
)

// BusesHandler handles the HTTP requests on the bus collection. It is
// responsible for listing all the buses and creating new ones.
type BusesHandler struct {
	DB data.DB
}

func (h BusesHandler) get(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
	if req.Header.Get("Accept") != "application/json" {
		notAcceptable(w)

		return
	}

	buses, err := h.DB.ReadAllBuses()
	if err != nil {
		errorResponse(w, Error{
			Status:  http.StatusInternalServerError,
			Details: err.Error(),
		})

		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(buses); err != nil {
		logrus.WithError(err).Error("could not encode buses to JSON")
	}
}

func (h BusesHandler) post(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
	if req.Header.Get("Accept") != "application/json" {
		notAcceptable(w)

		return
	}

	if req.Header.Get("Content-Type") != "application/json" {
		unsupportedMediaType(w)

		return
	}

	var bus model.Bus

	if err := json.NewDecoder(req.Body).Decode(&bus); err != nil {
		errorResponse(w, Error{
			Status:  http.StatusBadRequest,
			Details: err.Error(),
		})

		return
	}

	if bus.ID == "" {
		errorResponse(w, Error{
			Status:  http.StatusUnprocessableEntity,
			Details: "Missing bus ID",
		})

		return
	}

	exists, err := h.DB.ExistsBus(bus.ID)
	if err != nil {
		errorResponse(w, Error{
			Status:  http.StatusInternalServerError,
			Details: err.Error(),
		})

		return
	}

	if exists {
		errorResponse(w, Error{
			Status:  http.StatusConflict,
			Details: fmt.Sprintf("Bus \"%v\" already exists", bus.ID),
		})

		return
	}

	createdBus, err := h.DB.CreateBus(bus)
	if err != nil {
		switch err.(type) {
		case data.InvalidParameterError, data.MissingParameterError:
			errorResponse(w, Error{
				Status:  http.StatusUnprocessableEntity, // 422 Unprocessable Entity
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
	w.Header().Set("Location", "/bus/"+createdBus.ID)
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(createdBus); err != nil {
		logrus.WithError(err).Error("could not encode bus to JSON")
	}
}
