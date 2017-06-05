package data

type Source interface {
	CreateBus(Bus) error
	ReadAllBuses() ([]Bus, error)
	ReadBus(string) (Bus, error)
	UpdateBus(Bus) error
	DeleteBus(string) error

	Close() error
}
