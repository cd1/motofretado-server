package data

import "motorola.com/cdeives/motofretado/model"

type Source interface {
	CreateBus(model.Bus) error
	ReadAllBuses() ([]model.Bus, error)
	ReadBus(string) (model.Bus, error)
	UpdateBus(model.Bus) error
	DeleteBus(string) error

	Close() error
}
