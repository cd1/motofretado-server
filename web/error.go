package web

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Sirupsen/logrus"
)

// Error represents an HTTP error in JSON. The status code should be a valid
// HTTP status.
type Error struct {
	Status  int    `json:"status"`
	Details string `json:"details"`
}

func errorResponse(w http.ResponseWriter, e Error) {
	logrus.WithFields(logrus.Fields{
		"status":  fmt.Sprintf("%v %v", e.Status, http.StatusText(e.Status)),
		"details": e.Details,
	}).Warn("HTTP error")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(e.Status)
	if err := json.NewEncoder(w).Encode(e); err != nil {
		logrus.WithError(err).Error("error encoding Error to JSON")
	}
}
