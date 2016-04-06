package main

import (
	"flag"
	"net/http"
	"strconv"

	"github.com/Sirupsen/logrus"
	"motorola.com/cdeives/motofretado/data"
	"motorola.com/cdeives/motofretado/web"
)

var (
	dbURL string
	debug bool
	port  int
)

func init() {
	flag.StringVar(&dbURL, "db", "", "the database URL")
	flag.BoolVar(&debug, "debug", false, "enable debug logs")
	flag.IntVar(&port, "port", 8080, "the port to listen on")
}

func main() {
	flag.Parse()

	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	db, err := data.NewPostgresDB(dbURL)
	if err != nil {
		logrus.Fatal("error opening a database connection")
	}
	defer func() {
		if err := db.Close(); err != nil {
			logrus.WithError(err).Warn("could not close the database connection")
		}
	}()

	mux := web.BuildMux(db)

	logrus.WithFields(logrus.Fields{
		"port": port,
	}).Info("starting web server")
	if err := http.ListenAndServe(":"+strconv.Itoa(port), mux); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"port": port,
		}).Fatal("error running web server")
	}
}
