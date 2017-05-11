package jsonapi

import (
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/hashicorp/go-version"
	"github.com/pkg/errors"
	"motorola.com/cdeives/motofretado/model"
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

func ToBusDocument(bus model.Bus) BusDocument {
	doc := BusDocument{
		JSONAPI: &Root{
			Version: CurrentVersion,
		},
		Data: toBusData(bus),
	}

	return doc
}

func ToBusesDocument(buses []model.Bus) BusesDocument {
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

func FromBusDocument(doc BusDocument) (model.Bus, error) {
	if err := validateVersion(doc.JSONAPI); err != nil {
		return model.Bus{}, err
	}

	return fromBusData(doc.Data)
}

func FromBusesDocument(doc BusesDocument) ([]model.Bus, error) {
	if err := validateVersion(doc.JSONAPI); err != nil {
		return nil, err
	}

	buses := make([]model.Bus, len(doc.Data))

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

func toBusData(bus model.Bus) BusData {
	data := BusData{
		Type: BusType,
		ID:   bus.ID,
		Attributes: &BusAttributes{
			Latitude:  bus.Latitude,
			Longitude: bus.Longitude,
			CreatedAt: bus.CreatedAt,
			UpdatedAt: bus.UpdatedAt,
		},
	}

	return data
}

func fromBusData(data BusData) (model.Bus, error) {
	if data.Type != BusType {
		err := InvalidTypeError{
			Type:         data.Type,
			ExpectedType: BusType,
		}
		return model.Bus{}, errors.WithMessage(err, "invalid JSONAPI data type")
	}

	bus := model.Bus{
		ID: data.ID,
	}

	if data.Attributes != nil {
		bus.Latitude = data.Attributes.Latitude
		bus.Longitude = data.Attributes.Longitude
		bus.CreatedAt = data.Attributes.CreatedAt
		bus.UpdatedAt = data.Attributes.UpdatedAt
	}

	return bus, nil
}
