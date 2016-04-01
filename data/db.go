package data

import (
	"errors"

	"motorola.com/cdeives/motofretado/model"
)

// Common error values for database operations.
var (
	ErrFutureTime = errors.New("timestamp cannot be in the future")
	ErrIDExists   = errors.New("ID already exists")
	ErrIDNotFound = errors.New("ID not found")
	ErrPastTime   = errors.New("timestamp cannot be in the past")
)

// DB is a common interface for persistance operations. This interface can be
// implemented by in-memory databases, SQLite, MySQL, etc.
type DB interface {
	Close() error
	ExistsBus(string) (bool, error)
	CreateBus(model.Bus) (model.Bus, error)
	ReadAllBuses() ([]model.Bus, error)
	ReadBus(string) (model.Bus, error)
	UpdateBus(string, model.Bus) (model.Bus, error)
	DeleteBus(string) error
}
