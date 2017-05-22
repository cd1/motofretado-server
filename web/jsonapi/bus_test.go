package jsonapi

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/cd1/motofretado-server/data"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToBusDocument(t *testing.T) {
	now := time.Now()

	bus := data.Bus{
		ID:        "test-jsonapi",
		Latitude:  1.23,
		Longitude: 4.56,
		CreatedAt: now,
		UpdatedAt: now,
	}

	doc := ToBusDocument(bus)

	assert.Equal(t, BusType, doc.Data.Type, "bad data type")
	assert.Equal(t, bus.ID, doc.Data.ID, "bad ID")
	assert.Equal(t, bus.Latitude, doc.Data.Attributes.Latitude, "bad latitude")
	assert.Equal(t, bus.Longitude, doc.Data.Attributes.Longitude, "bad longitude")
	assert.Equal(t, bus.CreatedAt, doc.Data.Attributes.CreatedAt, "bad creation time")
	assert.Equal(t, bus.UpdatedAt, doc.Data.Attributes.UpdatedAt, "bad update time")
}

func TestToBusesDocument(t *testing.T) {
	nBuses := 2
	buses := make([]data.Bus, nBuses)

	for i := range buses {
		now := time.Now()

		bus := data.Bus{
			ID:        fmt.Sprintf("test-jsonapi-%v", i),
			Latitude:  1.23,
			Longitude: 4.56,
			CreatedAt: now,
			UpdatedAt: now,
		}

		buses[i] = bus
	}

	doc := ToBusesDocument(buses)

	assert.Len(t, doc.Data, nBuses, "bad buses size")
	for i, d := range doc.Data {
		originalBus := buses[i]

		assert.Equal(t, BusType, d.Type, "bad data type")
		assert.Equal(t, originalBus.ID, d.ID, "bad ID")
		assert.Equal(t, originalBus.Latitude, d.Attributes.Latitude, "bad latitude")
		assert.Equal(t, originalBus.Longitude, d.Attributes.Longitude, "bad longitude")
		assert.Equal(t, originalBus.CreatedAt, d.Attributes.CreatedAt, "bad creation time")
		assert.Equal(t, originalBus.UpdatedAt, d.Attributes.UpdatedAt, "bad update time")
	}
}
func TestFromBusDocument(t *testing.T) {
	t.Run("invalid version", func(subT *testing.T) {
		doc := BusDocument{
			JSONAPI: &Root{
				Version: "foo",
			},
		}

		_, err := FromBusDocument(doc)
		assert.Error(subT, err)
	})

	t.Run("unsupported version", func(subT *testing.T) {
		doc := BusDocument{
			JSONAPI: &Root{
				Version: "100.0",
			},
		}

		_, err := FromBusDocument(doc)
		if assert.Error(subT, err) {
			assert.IsType(subT, UnsupportedVersionError{}, errors.Cause(err))
		}
	})

	t.Run("invalid type", func(subT *testing.T) {
		doc := BusDocument{
			JSONAPI: &Root{
				Version: CurrentVersion,
			},
			Data: BusData{
				Type: "foo",
				ID:   "bar",
			},
		}

		_, err := FromBusDocument(doc)
		if assert.Error(subT, err) {
			assert.IsType(subT, InvalidTypeError{}, errors.Cause(err))
		}
	})

	t.Run("success", func(subT *testing.T) {
		now := time.Now()

		doc := BusDocument{
			JSONAPI: &Root{
				Version: CurrentVersion,
			},
			Data: BusData{
				Type: BusType,
				ID:   "bar",
				Attributes: &BusAttributes{
					Latitude:  1.23,
					Longitude: 4.56,
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
		}

		bus, err := FromBusDocument(doc)
		require.NoError(subT, err, "failed to convert bus document")
		assert.Equal(subT, doc.Data.ID, bus.ID, "bad ID")
		assert.Equal(subT, doc.Data.Attributes.Latitude, bus.Latitude, "bad latitude")
		assert.Equal(subT, doc.Data.Attributes.Longitude, bus.Longitude, "bad longitude")
		assert.Equal(subT, doc.Data.Attributes.CreatedAt, bus.CreatedAt, "bad creation time")
		assert.Equal(subT, doc.Data.Attributes.UpdatedAt, bus.UpdatedAt, "bad update time")
	})
}

func TestFromBusesDocument(t *testing.T) {
	t.Run("invalid version", func(subT *testing.T) {
		doc := BusesDocument{
			JSONAPI: &Root{
				Version: "foo",
			},
		}

		_, err := FromBusesDocument(doc)
		assert.Error(subT, err)
	})

	t.Run("unsupported version", func(subT *testing.T) {
		doc := BusesDocument{
			JSONAPI: &Root{
				Version: "100.0",
			},
		}

		_, err := FromBusesDocument(doc)
		if assert.Error(subT, err) {
			assert.IsType(subT, UnsupportedVersionError{}, errors.Cause(err))
		}
	})

	t.Run("invalid type", func(subT *testing.T) {
		doc := BusesDocument{
			JSONAPI: &Root{
				Version: CurrentVersion,
			},
			Data: []BusData{
				{
					Type: "foo",
					ID:   "bar",
				},
			},
		}

		_, err := FromBusesDocument(doc)
		if assert.Error(subT, err) {
			assert.IsType(subT, InvalidTypeError{}, errors.Cause(err))
		}
	})

	t.Run("success", func(subT *testing.T) {
		now := time.Now()

		doc := BusesDocument{
			JSONAPI: &Root{
				Version: CurrentVersion,
			},
			Data: []BusData{
				{
					Type: BusType,
					ID:   "bar",
					Attributes: &BusAttributes{
						Latitude:  1.23,
						Longitude: 4.56,
						CreatedAt: now,
						UpdatedAt: now,
					},
				},
				{
					Type: BusType,
					ID:   "baz",
					Attributes: &BusAttributes{
						Latitude:  7.89,
						Longitude: 0.12,
						CreatedAt: now,
						UpdatedAt: now,
					},
				},
			},
		}

		buses, err := FromBusesDocument(doc)
		require.NoError(subT, err, "failed to convert buses document")
		for i, b := range buses {
			assert.Equal(subT, doc.Data[i].ID, b.ID, "bad ID")
			assert.Equal(subT, doc.Data[i].Attributes.Latitude, b.Latitude, "bad latitude")
			assert.Equal(subT, doc.Data[i].Attributes.Longitude, b.Longitude, "bad longitude")
			assert.Equal(subT, doc.Data[i].Attributes.CreatedAt, b.CreatedAt, "bad creation time")
			assert.Equal(subT, doc.Data[i].Attributes.UpdatedAt, b.UpdatedAt, "bad update time")
		}
	})
}

func BenchmarkFromBusDocument(b *testing.B) {
	doc := BusDocument{
		JSONAPI: &Root{
			Version: CurrentVersion,
		},
		Data: BusData{
			Type: BusType,
			ID:   "bench-jsonapi",
			Attributes: &BusAttributes{
				Latitude:  1.23,
				Longitude: 4.56,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
	}

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		if _, err := FromBusDocument(doc); err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkFromBusesDocument(b *testing.B) {
	nBuses := []int{1, 2, 4, 8}

	for _, nb := range nBuses {
		b.Run(strconv.Itoa(nb), func(subB *testing.B) {
			busesData := make([]BusData, nb)

			for i := range busesData {
				busesData[i] = BusData{
					Type: BusType,
					ID:   fmt.Sprintf("bench-jsonapi-%v", i),
					Attributes: &BusAttributes{
						Latitude:  1.23,
						Longitude: 4.56,
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
				}
			}

			doc := BusesDocument{
				JSONAPI: &Root{
					Version: CurrentVersion,
				},
				Data: busesData,
			}

			subB.ResetTimer()

			for n := 0; n < subB.N; n++ {
				if _, err := FromBusesDocument(doc); err != nil {
					subB.Error(err)
				}
			}
		})
	}
}

func BenchmarkToBusDocument(b *testing.B) {
	bus := data.Bus{
		ID:        "id",
		Latitude:  1.23,
		Longitude: 4.56,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		_ = ToBusDocument(bus)
	}
}

func BenchmarkToBusesDocument(b *testing.B) {
	nBuses := []int{1, 2, 4, 8}

	for _, nb := range nBuses {
		b.Run(strconv.Itoa(nb), func(subB *testing.B) {
			buses := make([]data.Bus, nb)

			for i := range buses {
				buses[i] = data.Bus{
					ID:        fmt.Sprintf("bench-jsonapi-%v", i),
					Latitude:  1.23,
					Longitude: 4.56,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
			}

			subB.ResetTimer()

			for n := 0; n < subB.N; n++ {
				_ = ToBusesDocument(buses)
			}
		})
	}
}
