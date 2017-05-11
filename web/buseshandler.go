package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Sirupsen/logrus"
	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"
	"motorola.com/cdeives/motofretado/data"
	"motorola.com/cdeives/motofretado/web/jsonapi"
)

// BusesHandler handles the HTTP requests on the bus collection. It is
// responsible for listing all the buses and creating new ones.
type BusesHandler struct {
	repo *data.Repository
}

func (h BusesHandler) get(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
	if req.Header.Get("Accept") != jsonapi.ContentType {
		notAcceptable(w)

		return
	}

	buses, err := h.repo.ReadAllBuses()
	if err != nil {
		errorResponse(w, jsonapi.ErrorData{
			Status: strconv.Itoa(http.StatusInternalServerError),
			Title:  "Unexpected error",
			Detail: err.Error(),
		})

		return
	}

	busesDoc := jsonapi.ToBusesDocument(buses)
	scheme := requestScheme(req)
	for i, b := range busesDoc.Data {
		busesDoc.Data[i].Links = &jsonapi.Links{
			Self: fmt.Sprintf("%v://%v/bus/%v", scheme, req.Host, b.ID),
		}
	}

	w.Header().Set("Content-Type", jsonapi.ContentType)
	if err := json.NewEncoder(w).Encode(busesDoc); err != nil {
		logrus.WithError(err).Error("could not encode buses to JSON")
	}
}

func (h BusesHandler) post(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
	if req.Header.Get("Accept") != jsonapi.ContentType {
		notAcceptable(w)

		return
	}

	if req.Header.Get("Content-Type") != jsonapi.ContentType {
		unsupportedMediaType(w)

		return
	}

	var busDoc jsonapi.BusDocument

	if err := json.NewDecoder(req.Body).Decode(&busDoc); err != nil {
		errorResponse(w, jsonapi.ErrorData{
			Status: strconv.Itoa(http.StatusBadRequest),
			Title:  "Invalid JSON format",
			Detail: err.Error(),
		})

		return
	}

	bus, err := jsonapi.FromBusDocument(busDoc)
	if err != nil {
		switch errors.Cause(err).(type) {
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

	createdBus, err := h.repo.CreateBus(bus)
	if err != nil {
		switch causeErr := errors.Cause(err); causeErr.(type) {
		case data.DuplicateError:
			errorResponse(w, jsonapi.ErrorData{
				Status: strconv.Itoa(http.StatusConflict), // 409 Conflict
				Title:  "Existing bus ID",
				Detail: fmt.Sprintf("Bus \"%v\" already exists", bus.ID),
				Source: &jsonapi.ErrorSource{
					Pointer: "/data/id",
				},
			})
		case data.InvalidParameterError:
			errorResponse(w, jsonapi.ErrorData{
				Status: strconv.Itoa(http.StatusUnprocessableEntity), // 422 Unprocessable Entity
				Title:  "Invalid bus field",
				Detail: err.Error(),
				Source: &jsonapi.ErrorSource{
					Pointer: "/data/attributes/" + causeErr.(data.InvalidParameterError).Name,
				},
			})
		case data.MissingParameterError:
			jsonapiErr := jsonapi.ErrorData{
				Status: strconv.Itoa(http.StatusUnprocessableEntity), // 422 Unprocessable Entity
				Title:  "Missing bus field",
				Detail: err.Error(),
				Source: &jsonapi.ErrorSource{},
			}

			missingParameterName := causeErr.(data.MissingParameterError).Name
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

		return
	}

	createdBusDoc := jsonapi.ToBusDocument(createdBus)
	selfURL := fmt.Sprintf("%v://%v/bus/%v", requestScheme(req), req.Host, createdBus.ID)
	createdBusDoc.Data.Links = &jsonapi.Links{
		Self: selfURL,
	}

	w.Header().Set("Content-Type", jsonapi.ContentType)
	w.Header().Set("Location", selfURL)
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(createdBusDoc); err != nil {
		logrus.WithError(err).Error("could not encode bus to JSON")
	}
}
