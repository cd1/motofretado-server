package data

import (
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
)

type Repository struct {
	src Source
}

func (r Repository) CreateBus(bus Bus) (Bus, error) {
	logrus.WithFields(logrus.Fields{
		"id":         bus.ID,
		"latitude":   bus.Latitude,
		"longitude":  bus.Longitude,
		"created_at": bus.CreatedAt,
		"updated_at": bus.UpdatedAt,
	}).Debug("creating bus")
	if len(bus.ID) == 0 {
		return Bus{}, errors.WithMessage(MissingParameterError{"id"}, "missing bus ID")
	}

	if !bus.CreatedAt.IsZero() {
		err := InvalidParameterError{
			Name:  "created_at",
			Value: bus.CreatedAt,
		}
		return Bus{}, errors.WithMessage(err, "bus creation time cannot be specified")
	}

	if !bus.UpdatedAt.IsZero() {
		err := InvalidParameterError{
			Name:  "updated_at",
			Value: bus.UpdatedAt,
		}
		return Bus{}, errors.WithMessage(err, "bus update time cannot be specified")
	}

	now := time.Now()
	bus.CreatedAt = now
	bus.UpdatedAt = now

	if err := r.src.CreateBus(bus); err != nil {
		return Bus{}, err
	}

	return bus, nil
}

func (r Repository) ReadAllBuses() ([]Bus, error) {
	logrus.Debug("reading all buses")
	return r.src.ReadAllBuses()
}

func (r Repository) ReadBus(id string) (Bus, error) {
	logrus.WithFields(logrus.Fields{
		"id": id,
	}).Debug("reading bus")
	if len(id) == 0 {
		return Bus{}, errors.WithMessage(MissingParameterError{"id"}, "missing bus ID")
	}

	return r.src.ReadBus(id)
}

func (r Repository) UpdateBus(bus Bus) (Bus, error) {
	logrus.WithFields(logrus.Fields{
		"id":         bus.ID,
		"latitude":   bus.Latitude,
		"longitude":  bus.Longitude,
		"created_at": bus.CreatedAt,
		"updated_at": bus.UpdatedAt,
	}).Debug("updating bus")
	if len(bus.ID) == 0 {
		return Bus{}, errors.WithMessage(MissingParameterError{"id"}, "missing bus ID")
	}

	existingBus, err := r.src.ReadBus(bus.ID)
	if err != nil {
		return Bus{}, errors.Wrap(err, "failed to check existing bus")
	}

	if !bus.CreatedAt.IsZero() {
		if !bus.CreatedAt.Equal(existingBus.CreatedAt) {
			err := InvalidParameterError{
				Name:  "created_at",
				Value: bus.CreatedAt,
			}
			return Bus{}, errors.WithMessage(err, "bus creation time cannot be specified")
		}
	} else {
		bus.CreatedAt = existingBus.CreatedAt
	}

	if !bus.UpdatedAt.IsZero() && !bus.UpdatedAt.Equal(existingBus.UpdatedAt) {
		err := InvalidParameterError{
			Name:  "updated_at",
			Value: bus.UpdatedAt,
		}
		return Bus{}, errors.WithMessage(err, "bus update time cannot be specified")
	}
	bus.UpdatedAt = time.Now()

	if err := r.src.UpdateBus(bus); err != nil {
		return Bus{}, err
	}

	return bus, nil
}

func (r Repository) DeleteBus(id string) error {
	logrus.WithFields(logrus.Fields{
		"id": id,
	}).Debug("deleting bus")
	if len(id) == 0 {
		return MissingParameterError{"id"}
	}

	return r.src.DeleteBus(id)
}

func (r Repository) Close() error {
	logrus.Debug("closing connection to database")
	return r.src.Close()
}
