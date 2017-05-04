package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Sirupsen/logrus"
	"motorola.com/cdeives/motofretado/web/jsonapi"
)

func errorResponse(w http.ResponseWriter, e jsonapi.ErrorData) {
	statusInt, err := strconv.Atoi(e.Status)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"status": e.Status,
		}).Error("invalid HTTP error status; using 500 Internal Server Error")
		statusInt = http.StatusInternalServerError
	}

	logrus.WithFields(logrus.Fields{
		"status": fmt.Sprintf("%v %v", e.Status, http.StatusText(statusInt)),
		"title":  e.Title,
		"detail": e.Detail,
	}).Error("HTTP error")

	doc := jsonapi.ErrorsDocument{
		JSONAPI: &jsonapi.Root{
			Version: jsonapi.CurrentVersion,
		},
		Errors: []jsonapi.ErrorData{e},
	}

	w.Header().Set("Content-Type", jsonapi.ContentType)
	w.WriteHeader(statusInt)
	if err := json.NewEncoder(w).Encode(doc); err != nil {
		logrus.WithError(err).Error("error encoding Errors to JSON")
	}
}
