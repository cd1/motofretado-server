package data

import "motorola.com/cdeives/motofretado/model"

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
