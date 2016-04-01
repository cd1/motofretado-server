package main

import (
	"net/http"
	"os"

	"github.com/Sirupsen/logrus"
	"motorola.com/cdeives/motofretado/data"
	"motorola.com/cdeives/motofretado/web"
)

func main() {
	// reading environment variables
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	debug := (os.Getenv("DEBUG") == "TRUE")

	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	db := data.NewMemDB()
	mux := web.BuildMux(db)

	logrus.WithFields(logrus.Fields{
		"port": port,
	}).Info("starting web server")
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"port": port,
		}).Fatal("error running web server")
	}
}
