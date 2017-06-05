package jsonapi

import (
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/hashicorp/go-version"
	"github.com/pkg/errors"
	"motorola.com/cdeives/motofretado/data"
)

const BusType = "bus"

type BusDocument struct {
	JSONAPI *Root   `json:"jsonapi,omitempty"`
	Data    BusData `json:"data"`
	Links   *Links  `json:"links,omitempty"`
}

type BusesDocument struct {
	JSONAPI *Root     `json:"jsonapi,omitempty"`
	Data    []BusData `json:"data"`
	Links   *Links    `json:"links,omitempty"`
}

type BusData struct {
	Type       string         `json:"type"`
	ID         string         `json:"id"`
	Attributes *BusAttributes `json:"attributes"`
	Links      *Links         `json:"links,omitempty"`
}

type BusAttributes struct {
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func ToBusDocument(bus data.Bus) BusDocument {
	doc := BusDocument{
		JSONAPI: &Root{
			Version: CurrentVersion,
		},
		Data: toBusData(bus),
	}

	return doc
}

func ToBusesDocument(buses []data.Bus) BusesDocument {
	doc := BusesDocument{
		JSONAPI: &Root{
			Version: CurrentVersion,
		},
		Data: make([]BusData, len(buses)),
	}

	for i, b := range buses {
		doc.Data[i] = toBusData(b)
	}

	return doc
}

func FromBusDocument(doc BusDocument) (data.Bus, error) {
	if err := validateVersion(doc.JSONAPI); err != nil {
		return data.Bus{}, err
	}

	return fromBusData(doc.Data)
}

func FromBusesDocument(doc BusesDocument) ([]data.Bus, error) {
	if err := validateVersion(doc.JSONAPI); err != nil {
		return nil, err
	}

	buses := make([]data.Bus, len(doc.Data))

	for i, d := range doc.Data {
		bus, err := fromBusData(d)
		if err != nil {
			return nil, err
		}

		buses[i] = bus
	}

	return buses, nil
}

func validateVersion(api *Root) error {
	if api != nil && len(api.Version) > 0 {
		v, err := version.NewVersion(api.Version)
		if err != nil {
			logrus.WithError(err).Error("failed to parse JSONAPI version")
			versionErr := UnsupportedVersionError{
				Version:        api.Version,
				CurrentVersion: CurrentVersion,
			}
			return errors.Wrap(versionErr, "failed to parse JSONAPI version")
		}

		if v.GreaterThan(currentVersionStruct) {
			versionErr := UnsupportedVersionError{
				Version:        api.Version,
				CurrentVersion: CurrentVersion,
			}
			return errors.WithMessage(versionErr, "unsupported JSONAPI version")
		}
	}

	return nil
}

func toBusData(bus data.Bus) BusData {
	busData := BusData{
		Type: BusType,
		ID:   bus.ID,
		Attributes: &BusAttributes{
			Latitude:  bus.Latitude,
			Longitude: bus.Longitude,
			CreatedAt: bus.CreatedAt,
			UpdatedAt: bus.UpdatedAt,
		},
	}

	return busData
}

func fromBusData(busData BusData) (data.Bus, error) {
	if busData.Type != BusType {
		err := InvalidTypeError{
			Type:         busData.Type,
			ExpectedType: BusType,
		}
		return data.Bus{}, errors.WithMessage(err, "invalid JSONAPI busData type")
	}

	bus := data.Bus{
		ID: busData.ID,
	}

	if busData.Attributes != nil {
		bus.Latitude = busData.Attributes.Latitude
		bus.Longitude = busData.Attributes.Longitude
		bus.CreatedAt = busData.Attributes.CreatedAt
		bus.UpdatedAt = busData.Attributes.UpdatedAt
	}

	return bus, nil
}
