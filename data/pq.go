package data

import (
	"database/sql"
	"errors"

	"github.com/Sirupsen/logrus"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"motorola.com/cdeives/motofretado/model"
)

type postgresSource struct {
	db            *sqlx.DB
	insertStmt    *sqlx.Stmt
	selectAllStmt *sqlx.Stmt
	selectStmt    *sqlx.Stmt
	updateStmt    *sqlx.Stmt
	deleteStmt    *sqlx.Stmt
}

// NewPostgresRepository creates a new connection to a PostgreSQL database.
// The connection URL is the same one used by the command "psql". All tables are
// created during this function (if needed). After using the connection, the
// user must call Close.
func NewPostgresRepository(url string) (*Repository, error) {
	logrus.WithFields(logrus.Fields{
		"url": url,
	}).Debug("opening connection to Postgres")
	db, err := sqlx.Connect("postgres", url)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"url": url,
		}).Error("could not open a Postgres connection")
		return nil, err
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS buses (
			id TEXT PRIMARY KEY,
			latitude FLOAT8 NOT NULL,
			longitude FLOAT8 NOT NULL,
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL
		)`)
	if err != nil {
		logrus.WithError(err).Error("error creating the table \"buses\"")
		return nil, err
	}

	src := postgresSource{db: db}
	src.insertStmt, err = db.Preparex(`INSERT INTO buses (id, latitude, longitude, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)`)
	if err != nil {
		return nil, err
	}
	src.selectAllStmt, err = db.Preparex(`SELECT id, latitude, longitude, created_at, updated_at FROM buses ORDER BY id`)
	if err != nil {
		return nil, err
	}
	src.selectStmt, err = db.Preparex(`SELECT latitude, longitude, created_at, updated_at FROM buses WHERE id = $1`)
	if err != nil {
		return nil, err
	}
	src.updateStmt, err = db.Preparex(`UPDATE buses SET latitude = $2, longitude = $3, updated_at = $4 WHERE id = $1`)
	if err != nil {
		return nil, err
	}
	src.deleteStmt, err = db.Preparex(`DELETE FROM buses WHERE id = $1`)
	if err != nil {
		return nil, err
	}

	return &Repository{src: src}, nil
}

func (src postgresSource) CreateBus(bus model.Bus) error {
	res, err := src.insertStmt.Exec(bus.ID, bus.Latitude, bus.Longitude, bus.CreatedAt, bus.UpdatedAt)
	if err != nil {
		if err.(*pq.Error).Code == "23505" { // unique_violation
			return DuplicateError{bus.ID}
		}

		logrus.WithError(err).Error("error creating bus")
		return err
	}

	nRows, err := res.RowsAffected()
	if err != nil {
		logrus.WithError(err).Error("could not get number of affected rows")
		return err
	}
	if nRows != 1 {
		logrus.WithFields(logrus.Fields{
			"affected_rows": nRows,
		}).Error("unexpected number of rows were inserted")
		return errors.New("xxx")
	}

	return nil
}

func (src postgresSource) DeleteBus(id string) error {
	res, err := src.deleteStmt.Exec(id)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"id": id,
		}).Error("error deleting bus")
		return err
	}

	nRows, err := res.RowsAffected()
	if err != nil {
		logrus.WithError(err).Error("could not get number of affected rows")
		return err
	}
	if nRows == 0 {
		logrus.Info("no rows have been deleted")
		return ErrNoSuchRow
	}
	if nRows > 1 {
		logrus.WithFields(logrus.Fields{
			"affected_rows": nRows,
		}).Error("more rows than expected were deleted")
		return errors.New("xxx")
	}

	return nil
}

func (src postgresSource) ReadAllBuses() ([]model.Bus, error) {
	var buses []Bus

	if err := src.selectAllStmt.Select(&buses); err != nil {
		return nil, err
	}

	return buses, nil
}

func (src postgresSource) ReadBus(id string) (model.Bus, error) {
	bus := model.Bus{ID: id}

	if err := src.selectStmt.Get(&bus, id); err != nil {
		if err == sql.ErrNoRows {
			return model.Bus{}, ErrNoSuchRow
		}

		logrus.WithError(err).WithFields(logrus.Fields{
			"id": id,
		}).Error("error reading bus")
		return model.Bus{}, err
	}

	return bus, nil
}

func (src postgresSource) UpdateBus(bus model.Bus) error {
	res, err := src.updateStmt.Exec(bus.ID, bus.Latitude, bus.Longitude, bus.UpdatedAt)
	if err != nil {
		logrus.WithError(err).Error("error updating bus")
		return err
	}

	nRows, err := res.RowsAffected()
	if err != nil {
		logrus.WithError(err).Error("could not get number of affected rows")
		return err
	}
	if nRows == 0 {
		logrus.WithFields(logrus.Fields{
			"affected_rows": nRows,
		}).Error("no rows were updated")
		return ErrNoSuchRow
	}
	if nRows > 1 {
		logrus.WithFields(logrus.Fields{
			"affected_rows": nRows,
		}).Error("more rows than expected were updated")
		return errors.New("xxx")
	}

	return nil
}

func (src postgresSource) Close() error {
	if err := src.insertStmt.Close(); err != nil {
		return err
	}

	if err := src.selectAllStmt.Close(); err != nil {
		return err
	}

	if err := src.selectStmt.Close(); err != nil {
		return err
	}

	if err := src.updateStmt.Close(); err != nil {
		return err
	}

	if err := src.deleteStmt.Close(); err != nil {
		return err
	}

	if err := src.db.Close(); err != nil {
		return err
	}

	return nil
}
