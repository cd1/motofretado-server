package web

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Sirupsen/logrus"
	"github.com/julienschmidt/httprouter"
	"motorola.com/cdeives/motofretado/data"
	"motorola.com/cdeives/motofretado/web/jsonapi"
)

// BusHandler handles the HTTP requests on the bus resource. It is responsible
// for listing detailed information, updating and deleting individual buses.
type BusHandler struct {
	DB data.DB
}

func (h BusHandler) delete(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
	id := params.ByName("id")
	if len(id) == 0 {
		errorResponse(w, jsonapi.ErrorData{
			Status: strconv.Itoa(http.StatusBadRequest), // 400 Bad Request
			Title:  "Empty bus ID",
			Source: &jsonapi.ErrorSource{
				Pointer: "/data/id",
			},
		})

		return
	}

	exists, err := h.DB.ExistsBus(id)
	if err != nil {
		errorResponse(w, jsonapi.ErrorData{
			Status: strconv.Itoa(http.StatusInternalServerError), // 500 Internal Server Error
			Title:  "Unexpected error",
			Detail: err.Error(),
		})

		return
	}

	if !exists {
		errorResponse(w, jsonapi.ErrorData{
			Status: strconv.Itoa(http.StatusNotFound), // 404 Not Found
			Title:  "Bus ID not found",
			Detail: fmt.Sprintf("Bus \"%v\" doesn't exist", id),
			Source: &jsonapi.ErrorSource{
				Pointer: "/data/id",
			},
		})

		return
	}

	if err = h.DB.DeleteBus(id); err != nil {
		errorResponse(w, jsonapi.ErrorData{
			Status: strconv.Itoa(http.StatusInternalServerError), // 500 Internal Server Error
			Title:  "Unexpected error",
			Detail: err.Error(),
		})

		return
	}

	w.WriteHeader(http.StatusNoContent) // 204 No Content
}

func (h BusHandler) get(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
	if req.Header.Get("Accept") != jsonapi.ContentType {
		notAcceptable(w) // 406 Not Acceptable

		return
	}

	id := params.ByName("id")
	if len(id) == 0 {
		errorResponse(w, jsonapi.ErrorData{
			Status: strconv.Itoa(http.StatusBadRequest), // 400 Bad Request
			Title:  "Empty bus ID",
			Source: &jsonapi.ErrorSource{
				Pointer: "/data/id",
			},
		})

		return
	}

	exists, err := h.DB.ExistsBus(id)
	if err != nil {
		errorResponse(w, jsonapi.ErrorData{
			Status: strconv.Itoa(http.StatusInternalServerError), // 500 Internal Server Error
			Title:  "Unexpected error",
			Detail: err.Error(),
		})

		return
	}

	if !exists {
		errorResponse(w, jsonapi.ErrorData{
			Status: strconv.Itoa(http.StatusNotFound), // 404 Not Found
			Title:  "Bus ID not found",
			Detail: fmt.Sprintf("Bus \"%v\" doesn't exist", id),
			Source: &jsonapi.ErrorSource{
				Pointer: "/data/id",
			},
		})

		return
	}

	bus, err := h.DB.ReadBus(id)
	if err != nil {
		errorResponse(w, jsonapi.ErrorData{
			Status: strconv.Itoa(http.StatusInternalServerError), // 500 Internal Server Error
			Title:  "Unexpected error",
			Detail: err.Error(),
		})

		return
	}

	busDoc := jsonapi.ToBusDocument(bus)
	busDoc.Data.Links = &jsonapi.Links{
		Self: fmt.Sprintf("%v://%v/bus/%v", requestScheme(req), req.Host, id),
	}

	w.Header().Set("Content-Type", jsonapi.ContentType)
	if err := json.NewEncoder(w).Encode(busDoc); err != nil { // 200 OK
		logrus.WithError(err).Error("could not encode bus to JSON")
	}
}

func (h BusHandler) patch(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
	if req.Header.Get("Accept") != jsonapi.ContentType {
		notAcceptable(w) // 406 Not Acceptable

		return
	}

	if req.Header.Get("Content-Type") != jsonapi.ContentType {
		unsupportedMediaType(w) // 415 Unsupported Media Type

		return
	}

	id := params.ByName("id")
	if len(id) == 0 {
		errorResponse(w, jsonapi.ErrorData{
			Status: strconv.Itoa(http.StatusBadRequest), // 400 Bad Request
			Title:  "Empty bus ID",
			Source: &jsonapi.ErrorSource{
				Pointer: "/data/id",
			},
		})

		return
	}

	exists, err := h.DB.ExistsBus(id)
	if err != nil {
		errorResponse(w, jsonapi.ErrorData{
			Status: strconv.Itoa(http.StatusInternalServerError), // 500 Internal Server Error
			Title:  "Unexpected error",
			Detail: err.Error(),
		})

		return
	}

	if !exists {
		errorResponse(w, jsonapi.ErrorData{
			Status: strconv.Itoa(http.StatusNotFound), // 404 Not Found
			Title:  "Bus ID not found",
			Detail: fmt.Sprintf("Bus \"%v\" doesn't exist", id),
			Source: &jsonapi.ErrorSource{
				Pointer: "/data/id",
			},
		})

		return
	}

	var busDoc jsonapi.BusDocument

	if err := json.NewDecoder(req.Body).Decode(&busDoc); err != nil {
		errorResponse(w, jsonapi.ErrorData{
			Status: strconv.Itoa(http.StatusBadRequest), // 400 Bad Request
			Title:  "Invalid JSON format",
			Detail: err.Error(),
		})

		return
	}

	bus, err := jsonapi.FromBusDocument(busDoc)
	if err != nil {
		switch err.(type) {
		case jsonapi.UnsupportedVersionError:
			errorResponse(w, jsonapi.ErrorData{
				Status: strconv.Itoa(http.StatusBadRequest),
				Title:  "Unsupported JSONAPI version",
				Detail: err.Error(),
				Source: &jsonapi.ErrorSource{
					Pointer: "/jsonapi/version",
				},
			})
		case jsonapi.InvalidTypeError:
			errorResponse(w, jsonapi.ErrorData{
				Status: strconv.Itoa(http.StatusConflict),
				Title:  "Invalid JSONAPI data type",
				Detail: err.Error(),
				Source: &jsonapi.ErrorSource{
					Pointer: "/data/type",
				},
			})
		default:
			errorResponse(w, jsonapi.ErrorData{
				Status: strconv.Itoa(http.StatusBadRequest),
				Title:  "Invalid JSONAPI data",
				Detail: err.Error(),
				Source: &jsonapi.ErrorSource{
					Pointer: "/data",
				},
			})
		}

		return
	}

	updatedBus, err := h.DB.UpdateBus(id, bus)
	if err != nil {
		if err == sql.ErrNoRows {
			errorResponse(w, jsonapi.ErrorData{
				Status: strconv.Itoa(http.StatusNotFound), // 404 Not Found
				Title:  "Bus ID not found",
				Detail: fmt.Sprintf("Bus \"%v\" doesn't exist", id),
				Source: &jsonapi.ErrorSource{
					Pointer: "/data/id",
				},
			})
		} else {
			switch err.(type) {
			case data.InvalidParameterError:
				jsonapiErr := jsonapi.ErrorData{
					Status: strconv.Itoa(http.StatusUnprocessableEntity), // 422 Unprocessable Entity
					Title:  "Invalid bus field",
					Detail: err.Error(),
					Source: &jsonapi.ErrorSource{},
				}

				invalidParameterName := err.(data.InvalidParameterError).Name
				if invalidParameterName == "id" {
					jsonapiErr.Source.Pointer = "/data/id"
				} else {
					jsonapiErr.Source.Pointer = "/data/attributes/" + invalidParameterName
				}

				errorResponse(w, jsonapiErr)
			case data.MissingParameterError:
				jsonapiErr := jsonapi.ErrorData{
					Status: strconv.Itoa(http.StatusUnprocessableEntity), // 422 Unprocessable Entity
					Title:  "Missing bus field",
					Detail: err.Error(),
					Source: &jsonapi.ErrorSource{},
				}

				missingParameterName := err.(data.MissingParameterError).Name
				if missingParameterName == "id" {
					jsonapiErr.Source.Pointer = "/data/id"
				} else {
					jsonapiErr.Source.Pointer = "/data/attributes/" + missingParameterName
				}

				errorResponse(w, jsonapiErr)
			default:
				errorResponse(w, jsonapi.ErrorData{
					Status: strconv.Itoa(http.StatusInternalServerError), // 500 Internal Server Error
					Title:  "Unexpected error",
					Detail: err.Error(),
				})
			}
		}

		return
	}

	updatedBusDoc := jsonapi.ToBusDocument(updatedBus)
	updatedBusDoc.Data.Links = &jsonapi.Links{
		Self: fmt.Sprintf("%v://%v/bus/%v", requestScheme(req), req.Host, id),
	}

	w.Header().Set("Content-Type", jsonapi.ContentType)
	if err := json.NewEncoder(w).Encode(updatedBusDoc); err != nil { // 200 OK
		logrus.WithError(err).Error("could not encode bus to JSON")
	}
}
