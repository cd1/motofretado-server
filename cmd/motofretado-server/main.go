package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/cd1/motofretado-server/data"
	"github.com/cd1/motofretado-server/web"
	"github.com/urfave/cli"
)

func main() {
	var dbURL string
	var debug bool
	var port int

	app := cli.NewApp()
	app.Name = "Moto Fretado server"
	app.Usage = "The web server behind the Moto Fretado app"
	app.Version = "0.6.0-dev"
	app.Authors = []cli.Author{
		{
			Name:  "Crístian Deives",
			Email: "cristiandeives@gmail.com",
		},
	}
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "database-url, u",
			Usage:       "database `URL`",
			Destination: &dbURL,
			EnvVar:      "DATABASE_URL",
		},
		cli.BoolFlag{
			Name:        "debug, d",
			Usage:       "enable debug logs",
			Destination: &debug,
		},
		cli.IntFlag{
			Name:        "port, p",
			Value:       8080,
			Usage:       "listen to the specified `port`",
			Destination: &port,
			EnvVar:      "PORT",
		},
	}

	app.Action = func(c *cli.Context) error {
		if debug {
			logrus.SetLevel(logrus.DebugLevel)
		}

		repo, err := data.NewPostgresRepository(dbURL)
		if err != nil {
			logrus.Error("error opening a database connection")
			return cli.NewExitError(err.Error(), 1)
		}
		defer func() {
			if err := repo.Close(); err != nil {
				logrus.WithError(err).Warn("could not close the database connection")
			}
		}()

		mux := web.BuildMux(repo)

		logrus.WithFields(logrus.Fields{
			"port": port,
		}).Info("starting web server")
		if err := http.ListenAndServe(fmt.Sprintf(":%v", port), mux); err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"port": port,
			}).Error("error running web server")
			return cli.NewExitError(err.Error(), 1)
		}

		return nil
	}

	if err := app.Run(os.Args); err != nil {
		logrus.WithError(err).Fatal("error running the program")
	}
}
