package data

import (
	"database/sql"
	"time"

	"github.com/Sirupsen/logrus"
	_ "github.com/lib/pq" // database/sql driver
	"motorola.com/cdeives/motofretado/model"
)

// PostgresDB is a connection to a PostgreSQL database. To create a new instance
// of it, use NewPostgresDB.
type PostgresDB struct {
	conn *sql.DB
}

// NewPostgresDB creates a new connection to a PostgreSQL database.
// The connection URL is the same one used by the command "psql". All tables are
// created during this function (if needed). After using the connection, the
// user must call Close.
func NewPostgresDB(url string) (PostgresDB, error) {
	logrus.WithFields(logrus.Fields{
		"url": url,
	}).Debug("opening connection to Postgres")
	conn, err := sql.Open("postgres", url)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"url": url,
		}).Error("could not open a Postgres connection")
		return PostgresDB{}, err
	}

	if err = conn.Ping(); err != nil {
		logrus.WithError(err).Error("error pinging the Postgres database connection")
		return PostgresDB{}, err
	}

	_, err = conn.Exec(`CREATE TABLE IF NOT EXISTS buses (
			id TEXT PRIMARY KEY,
			latitude FLOAT8 NOT NULL,
			longitude FLOAT8 NOT NULL,
			updated_at timestamp NOT NULL,
			created_at timestamp NOT NULL
		)`)
	if err != nil {
		logrus.WithError(err).Error("error creating the table \"buses\"")
		return PostgresDB{}, err
	}

	db := PostgresDB{conn: conn}

	return db, nil
}

// Close closes the connection to the database.
func (db PostgresDB) Close() error {
	logrus.Debug("closing connection to Postgres")
	return db.conn.Close()
}

// CreateBus creates a new bus in the database. It returns the created bus with
// the auto generated values.
func (db PostgresDB) CreateBus(bus model.Bus) (model.Bus, error) {
	logrus.WithFields(logrus.Fields{
		"id":         bus.ID,
		"latitude":   bus.Latitude,
		"longitude":  bus.Longitude,
		"updated_at": bus.UpdatedAt,
		"created_at": bus.CreatedAt,
	}).Debug("creating bus")
	if bus.ID == "" {
		err := MissingParameterError{"id"}
		logrus.WithError(err).Error("missing bus ID")
		return model.Bus{}, err
	}

	if !bus.UpdatedAt.IsZero() {
		err := InvalidParameterError{
			Name:  "updated_at",
			Value: bus.UpdatedAt,
		}
		logrus.WithError(err).WithFields(logrus.Fields{
			"updated_at": bus.UpdatedAt,
		}).Error("cannot specify update time when creating bus")
		return model.Bus{}, err
	}

	if !bus.CreatedAt.IsZero() {
		err := InvalidParameterError{
			Name:  "created_at",
			Value: bus.CreatedAt,
		}
		logrus.WithError(err).WithFields(logrus.Fields{
			"created_at": bus.CreatedAt,
		}).Error("cannot specify create time when creating bus")
		return model.Bus{}, err
	}

	now := time.Now()
	bus.UpdatedAt = now
	bus.CreatedAt = now

	res, err := db.conn.Exec("INSERT INTO buses (id, latitude, longitude, updated_at, created_at) VALUES ($1, $2, $3, $4, $5)",
		bus.ID, bus.Latitude, bus.Longitude, bus.UpdatedAt, bus.CreatedAt)
	if err != nil {
		logrus.WithError(err).Error("error creating bus")
		return model.Bus{}, err
	}

	nRows, err := res.RowsAffected()
	if err != nil {
		logrus.WithError(err).Warn("could not get number of affected rows")
	}
	if expected := int64(1); nRows != expected {
		logrus.WithFields(logrus.Fields{
			"actual_rows":   nRows,
			"expected_rows": expected,
		}).Warn("unexpected number of affected rows")
	}

	return bus, nil
}

// DeleteBus deletes a bus from the database.
func (db PostgresDB) DeleteBus(id string) error {
	logrus.WithFields(logrus.Fields{
		"id": id,
	}).Debug("deleting bus")
	res, err := db.conn.Exec("DELETE FROM buses WHERE id = $1", id)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"id": id,
		}).Error("error deleting bus")
		return err
	}

	nRows, err := res.RowsAffected()
	if err != nil {
		logrus.WithError(err).Warn("could not get number of affected rows")
	}
	if expected := int64(1); nRows != expected {
		logrus.WithFields(logrus.Fields{
			"actual_rows":   nRows,
			"expected_rows": expected,
		}).Warn("unexpected number of affected rows")
	}

	return nil
}

// ExistsBus checks if a bus exists in the database.
func (db PostgresDB) ExistsBus(id string) (bool, error) {
	logrus.WithFields(logrus.Fields{
		"id": id,
	}).Debug("checking if bus exists")

	var exists bool

	if err := db.conn.QueryRow("SELECT EXISTS(SELECT 1 FROM buses WHERE id = $1)", id).Scan(&exists); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"id": id,
		}).Error("error checking if bus exists")
		return false, err
	}

	return exists, nil
}

// ReadAllBuses returns all buses from the database.
func (db PostgresDB) ReadAllBuses() ([]model.Bus, error) {
	logrus.Debug("reading all buses")
	rows, err := db.conn.Query("SELECT id, latitude, longitude, updated_at, created_at FROM buses")
	if err != nil {
		logrus.WithError(err).Error("error reading all buses")
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			logrus.WithError(err).Warn("could not close query rows correctly")
		}
	}()

	var buses []model.Bus

	for rows.Next() {
		var id string
		var latitude, longitude float64
		var updatedAt, createdAt time.Time

		if err = rows.Scan(&id, &latitude, &longitude, &updatedAt, &createdAt); err != nil {
			logrus.WithError(err).Error("error reading the values of one of the buses")
			return nil, err
		}

		bus := model.Bus{
			ID:        id,
			Latitude:  latitude,
			Longitude: longitude,
			UpdatedAt: updatedAt,
			CreatedAt: createdAt,
		}

		buses = append(buses, bus)
	}
	if err = rows.Err(); err != nil {
		logrus.WithError(err).Error("error iterating through all buses")
		return nil, err
	}

	return buses, nil
}

// ReadBus returns a specific bus from the database based on its ID. If it
// doesn't exist, an error (sql.ErrNoRows) will be returned.
func (db PostgresDB) ReadBus(id string) (model.Bus, error) {
	logrus.WithFields(logrus.Fields{
		"id": id,
	}).Debug("reading bus")

	var latitude, longitude float64
	var updatedAt, createdAt time.Time

	err := db.conn.QueryRow("SELECT latitude, longitude, updated_at, created_at FROM buses WHERE id = $1", id).
		Scan(&latitude, &longitude, &updatedAt, &createdAt)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"id": id,
		}).Error("error reading bus")
		return model.Bus{}, err
	}

	bus := model.Bus{
		ID:        id,
		Latitude:  latitude,
		Longitude: longitude,
		UpdatedAt: updatedAt,
		CreatedAt: createdAt,
	}

	return bus, nil
}

// UpdateBus updates a bus in the database based on its ID. It returns the
// updated bus with its final values.
func (db PostgresDB) UpdateBus(id string, bus model.Bus) (model.Bus, error) {
	logrus.WithFields(logrus.Fields{
		"id":         bus.ID,
		"latitude":   bus.Latitude,
		"longitude":  bus.Longitude,
		"updated_at": bus.UpdatedAt,
		"created_at": bus.CreatedAt,
	}).Debug("updating bus data")
	if id == "" {
		err := MissingParameterError{"id"}
		logrus.WithError(err).Error("missing bus ID")
		return model.Bus{}, err
	}

	if bus.ID != "" {
		err := InvalidParameterError{
			Name:  "id",
			Value: bus.ID,
		}
		logrus.WithError(err).Error("cannot specify ID when updating bus")
		return model.Bus{}, err
	}

	if !bus.CreatedAt.IsZero() {
		err := InvalidParameterError{
			Name:  "created_at",
			Value: bus.CreatedAt,
		}
		logrus.WithError(err).WithFields(logrus.Fields{
			"created_at": bus.CreatedAt,
		}).Error("cannot specify create time when updating bus")
		return model.Bus{}, err
	}

	existingBus, err := db.ReadBus(id)
	if err != nil {
		return model.Bus{}, err
	}

	bus.CreatedAt = existingBus.CreatedAt

	now := time.Now()
	if bus.UpdatedAt.IsZero() {
		bus.UpdatedAt = now
	} else if bus.UpdatedAt.Before(existingBus.UpdatedAt) {
		err := InvalidParameterError{
			Name:  "updated_at",
			Value: bus.UpdatedAt,
		}
		logrus.WithError(err).WithFields(logrus.Fields{
			"updated_at":          bus.UpdatedAt,
			"existing_updated_at": existingBus.UpdatedAt,
		}).Error("cannot specify update time before current update time")
		return model.Bus{}, err
	} else if bus.UpdatedAt.After(now) {
		err := InvalidParameterError{
			Name:  "updated_at",
			Value: bus.UpdatedAt,
		}
		logrus.WithError(err).WithFields(logrus.Fields{
			"updated_at": bus.UpdatedAt,
		}).Error("cannot specify update time in the future")
		return model.Bus{}, err
	}

	if bus.Latitude == 0 {
		bus.Latitude = existingBus.Latitude
	}

	if bus.Longitude == 0 {
		bus.Longitude = existingBus.Longitude
	}

	res, err := db.conn.Exec("UPDATE buses SET latitude = $2, longitude = $3, updated_at = $4 WHERE id = $1", id, bus.Latitude, bus.Longitude, bus.UpdatedAt)
	if err != nil {
		logrus.WithError(err).Error("error updating bus")
		return model.Bus{}, err
	}

	nRows, err := res.RowsAffected()
	if err != nil {
		logrus.WithError(err).Warn("could not get number of affected rows")
	}
	if expected := int64(1); nRows != expected {
		logrus.WithFields(logrus.Fields{
			"actual_rows":   nRows,
			"expected_rows": expected,
		}).Warn("unexpected number of affected rows")
	}

	bus.ID = id

	return bus, nil
}
