package data

import "time"

// Bus represents a bus ("fretado") on the system. It contains the last location
// information (i.e. latitude + longitude).
type Bus struct {
	ID        string
	Latitude  float64
	Longitude float64
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
