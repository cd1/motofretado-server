package data

import (
	"database/sql"

	"github.com/Sirupsen/logrus"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pkg/errors"
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
		return nil, errors.Wrap(err, "could not open a Postgres connection")
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS buses (
			id TEXT PRIMARY KEY,
			latitude FLOAT8 NOT NULL,
			longitude FLOAT8 NOT NULL,
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL
		)`)
	if err != nil {
		return nil, errors.Wrap(err, "error creating the table \"buses\"")
	}

	src := postgresSource{db: db}
	src.insertStmt, err = db.Preparex(`INSERT INTO buses (id, latitude, longitude, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)`)
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare INSERT statement")
	}
	src.selectAllStmt, err = db.Preparex(`SELECT id, latitude, longitude, created_at, updated_at FROM buses ORDER BY id`)
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare SELECT (all) statement")
	}
	src.selectStmt, err = db.Preparex(`SELECT latitude, longitude, created_at, updated_at FROM buses WHERE id = $1`)
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare SELECT statement")
	}
	src.updateStmt, err = db.Preparex(`UPDATE buses SET latitude = $2, longitude = $3, updated_at = $4 WHERE id = $1`)
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare UPDATE statement")
	}
	src.deleteStmt, err = db.Preparex(`DELETE FROM buses WHERE id = $1`)
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare DELETE statement")
	}

	return &Repository{src: src}, nil
}

func (src postgresSource) CreateBus(bus model.Bus) error {
	res, err := src.insertStmt.Exec(bus.ID, bus.Latitude, bus.Longitude, bus.CreatedAt, bus.UpdatedAt)
	if err != nil {
		if err.(*pq.Error).Code == "23505" { // unique_violation
			return errors.WithMessage(DuplicateError{bus.ID}, "bus with the same ID already exists")
		}
		return errors.Wrap(err, "error creating bus")
	}

	nRows, err := res.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "could not get number of affected rows")
	}
	if nRows != 1 {
		return errors.Wrap(err, "unexpected number of rows were inserted")
	}

	return nil
}

func (src postgresSource) DeleteBus(id string) error {
	res, err := src.deleteStmt.Exec(id)
	if err != nil {
		return errors.Wrap(err, "error deleting bus")
	}

	nRows, err := res.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "could not get number of affected rows")
	}
	if nRows == 0 {
		return errors.WithMessage(ErrNoSuchRow, "no rows have been deleted")
	}
	if nRows > 1 {
		return errors.New("more rows than expected were deleted")
	}

	return nil
}

func (src postgresSource) ReadAllBuses() ([]model.Bus, error) {
	var buses []model.Bus

	if err := src.selectAllStmt.Select(&buses); err != nil {
		return nil, errors.Wrap(err, "failed to read all buses")
	}

	return buses, nil
}

func (src postgresSource) ReadBus(id string) (model.Bus, error) {
	bus := model.Bus{ID: id}

	if err := src.selectStmt.Get(&bus, id); err != nil {
		if err == sql.ErrNoRows {
			return model.Bus{}, errors.WithMessage(ErrNoSuchRow, "bus not found")
		}

		return model.Bus{}, errors.Wrap(err, "error reading bus")
	}

	return bus, nil
}

func (src postgresSource) UpdateBus(bus model.Bus) error {
	res, err := src.updateStmt.Exec(bus.ID, bus.Latitude, bus.Longitude, bus.UpdatedAt)
	if err != nil {
		return errors.Wrap(err, "error updating bus")
	}

	nRows, err := res.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "could not get number of affected rows")
	}
	if nRows == 0 {
		return errors.WithMessage(ErrNoSuchRow, "no rows were updated")
	}
	if nRows > 1 {
		return errors.New("more rows than expected were updated")
	}

	return nil
}

func (src postgresSource) Close() error {
	if err := src.insertStmt.Close(); err != nil {
		return errors.Wrap(err, "failed to close INSERT statement")
	}

	if err := src.selectAllStmt.Close(); err != nil {
		return errors.Wrap(err, "failed to close SELECT (all) statement")
	}

	if err := src.selectStmt.Close(); err != nil {
		return errors.Wrap(err, "failed to close SELECT statement")
	}

	if err := src.updateStmt.Close(); err != nil {
		return errors.Wrap(err, "failed to close UPDATE statement")
	}

	if err := src.deleteStmt.Close(); err != nil {
		return errors.Wrap(err, "failed to close DELETE statement")
	}

	if err := src.db.Close(); err != nil {
		return errors.Wrap(err, "failed to close connection to Postgres")
	}

	return nil
}
