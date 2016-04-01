package data

import (
	"time"

	"github.com/Sirupsen/logrus"
	"motorola.com/cdeives/motofretado/model"
)

// MemDB is an in-memory database. Every instance of this type will store an
// independent set of data. There's no way of saving/loading its content to/from
// somewhere else.
// This type implements the interface "DB".
type MemDB struct {
	buses []model.Bus
}

// NewMemDB creates a new instance of "MemDB".
func NewMemDB() *MemDB {
	logrus.Debug("initializing in-memory database")
	return &MemDB{
		buses: make([]model.Bus, 0),
	}
}

// Close closes the database connection. This implementation is actually no-op
// and it will never return an error.
func (db MemDB) Close() error {
	logrus.Debug("closing in-memory database")
	return nil
}

// ExistsBus checks if a bus exists in the database. This implementation will
// never return an error.
func (db MemDB) ExistsBus(id string) (bool, error) {
	logrus.WithFields(logrus.Fields{
		"id": id,
	}).Debug("checking if bus exists")
	for _, b := range db.buses {
		if b.ID == id {
			return true, nil
		}
	}

	return false, nil
}

// CreateBus inserts a bus in the database. If there's already another bus with
// the same ID, "ErrIDExists" is returned.
func (db *MemDB) CreateBus(bus model.Bus) (model.Bus, error) {
	logrus.WithFields(logrus.Fields{
		"id":         bus.ID,
		"latitude":   bus.Latitude,
		"longitude":  bus.Longitude,
		"updated_at": bus.UpdatedAt,
	}).Debug("creating bus")
	for _, b := range db.buses {
		if b.ID == bus.ID {
			logrus.WithFields(logrus.Fields{
				"id": bus.ID,
			}).Error("bus already exists")
			return model.Bus{}, ErrIDExists
		}
	}

	now := time.Now()
	bus.UpdatedAt = now
	bus.CreatedAt = now

	db.buses = append(db.buses, bus)

	return bus, nil
}

// ReadAllBuses returns all buses from the database. This implementation will
// never return an error.
func (db MemDB) ReadAllBuses() ([]model.Bus, error) {
	logrus.Debug("reading all buses")
	ret := make([]model.Bus, len(db.buses))
	copy(ret, db.buses)

	return ret, nil
}

// ReadBus returns the bus which has the specified ID. If it cannot be
// found, the function returns "ErrIDNotFound".
func (db MemDB) ReadBus(id string) (model.Bus, error) {
	logrus.WithFields(logrus.Fields{
		"id": id,
	}).Debug("reading bus")
	for _, b := range db.buses {
		if b.ID == id {
			return b, nil
		}
	}

	return model.Bus{}, ErrIDNotFound
}

// UpdateBus updates ...
func (db MemDB) UpdateBus(id string, bus model.Bus) (model.Bus, error) {
	logrus.WithFields(logrus.Fields{
		"id":         id,
		"latitude":   bus.Latitude,
		"longitude":  bus.Longitude,
		"updated_at": bus.UpdatedAt,
	}).Debug("updating bus data")

	var existingBus *model.Bus

	for i, b := range db.buses {
		if b.ID == id {
			existingBus = &db.buses[i]
		}
	}

	if existingBus == nil {
		logrus.WithFields(logrus.Fields{
			"id": id,
		}).Error("bus doesn't exist")
		return model.Bus{}, ErrIDNotFound
	}

	now := time.Now()
	if bus.UpdatedAt.IsZero() {
		existingBus.UpdatedAt = now
	} else if bus.UpdatedAt.After(now) {
		return model.Bus{}, ErrFutureTime
	} else if bus.UpdatedAt.Before(existingBus.UpdatedAt) {
		return model.Bus{}, ErrPastTime
	}

	if bus.Latitude != 0 || existingBus.Latitude == 0 {
		existingBus.Latitude = bus.Latitude
	}

	if bus.Longitude != 0 || existingBus.Longitude == 0 {
		existingBus.Longitude = bus.Longitude
	}

	return *existingBus, nil
}

// DeleteBus deletes a bus from the database. If a bus with the specified ID
// doesn't exist, the function returns "ErrIDNotFound".
func (db *MemDB) DeleteBus(id string) error {
	logrus.WithFields(logrus.Fields{
		"id": id,
	}).Debug("deleting bus")
	for i, b := range db.buses {
		if b.ID == id {
			db.buses = append(db.buses[0:i], db.buses[i+1:len(db.buses)]...)
			return nil
		}
	}

	return ErrIDNotFound
}
