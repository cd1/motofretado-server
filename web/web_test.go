package web

import (
	"fmt"
	"os"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/cd1/motofretado-server/data"
)

const busesCount = 3

var repo *data.Repository

func setUp() {
	var err error

	env := os.Getenv("POSTGRES_URL")
	repo, err = data.NewPostgresRepository(env)
	if err != nil {
		panic(err)
	}

	for n := 0; n < busesCount; n++ {
		bus := data.Bus{
			ID: fmt.Sprintf("initial-bus-%v", n),
		}

		if _, err = repo.CreateBus(bus); err != nil {
			panic(err)
		}
	}

	busesHandler.repo = repo
	busHandler.repo = repo
}

func tearDown() {
	for n := 0; n < busesCount; n++ {
		if err := repo.DeleteBus(fmt.Sprintf("initial-bus-%v", n)); err != nil {
			logrus.WithError(err).Error("failed to delete bus")
		}
	}

	if err := repo.Close(); err != nil {
		logrus.WithError(err).Error("failed to close connection")
	}
}

func TestMain(m *testing.M) {
	setUp()
	status := m.Run()
	tearDown()

	os.Exit(status)
}
