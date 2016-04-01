package web

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Sirupsen/logrus"
	"golang.org/x/net/webdav"
)

// Error represents an HTTP error in JSON. The status code should be a valid
// HTTP status.
type Error struct {
	Status  int    `json:"status"`
	Details string `json:"details"`
}

func errorResponse(w http.ResponseWriter, e Error) {
	var statusText string

	if t := http.StatusText(e.Status); t != "" {
		statusText = t
	} else if t := webdav.StatusText(e.Status); t != "" {
		statusText = t
	}

	logrus.WithFields(logrus.Fields{
		"status":  fmt.Sprintf("%v %v", e.Status, statusText),
		"details": e.Details,
	}).Warn("HTTP error")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(e.Status)
	if err := json.NewEncoder(w).Encode(e); err != nil {
		logrus.WithError(err).Error("error encoding Error to JSON")
	}
}
