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
	debug bool
	port  int
)

func init() {
	flag.BoolVar(&debug, "debug", false, "enable debug logs")
	flag.IntVar(&port, "port", 8080, "the port to listen on")
}

func main() {
	flag.Parse()

	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	db := data.NewMemDB()
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
