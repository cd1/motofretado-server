package model

import "time"

// Bus represents a bus ("fretado") on the system. It contains the last location
// information (i.e. latitude + longitude).
type Bus struct {
	ID        string    `json:"id"`
	Latitude  float64   `json:"lat"`
	Longitude float64   `json:"long"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedAt time.Time `json:"-"`
}
