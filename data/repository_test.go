package data

import (
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var repo *Repository

func newRepository() (*Repository, error) {
	env := os.Getenv("POSTGRES_URL")
	return NewPostgresRepository(env)
}

func TestMain(m *testing.M) {
	// set up
	var err error

	repo, err = newRepository()
	if err != nil {
		panic(err)
	}

	status := m.Run()

	// tear down
	repo.Close()

	os.Exit(status)
}

func TestNewPostgresRepository(t *testing.T) {
	// don't use the global connection because we need to close it here
	_, err := NewPostgresRepository("")
	assert.Error(t, err, "invalid dialect and/or URL")

	testRepo, err := newRepository()
	require.NoError(t, err, "failed to open connection")

	if err = testRepo.Close(); err != nil {
		t.Logf("failed to close the connection: %v", err)
	}
}

func TestRepository_Close(t *testing.T) {
	// don't use the global connection because we need to close it here
	testRepo, err := newRepository()
	if err != nil {
		t.Skipf("failed to open test connection: %v", err)
	}

	err = testRepo.Close()
	assert.NoError(t, err)
}

func TestRepository_CreateBus(t *testing.T) {
	t.Run("missing ID", func(subT *testing.T) {
		var bus Bus

		_, err := repo.CreateBus(bus)
		defer repo.DeleteBus(bus.ID)

		switch causeErr := errors.Cause(err); causeErr.(type) {
		case MissingParameterError:
			assert.Equal(subT, "id", causeErr.(MissingParameterError).Name, "wrong missing parameter name")
		default:
			assert.Fail(subT, "unexpected error", "%T: %[1]v", causeErr)
		}
	})

	t.Run("creation time non-null", func(subT *testing.T) {
		bus := Bus{
			ID:        "test-create",
			CreatedAt: time.Now(),
		}

		_, err := repo.CreateBus(bus)
		defer repo.DeleteBus(bus.ID)

		switch causeErr := errors.Cause(err); causeErr.(type) {
		case InvalidParameterError:
			assert.Equal(subT, "created_at", causeErr.(InvalidParameterError).Name, "wrong invalid parameter name")
		default:
			assert.Fail(subT, "unexpected error", "%T: %[1]v", causeErr)
		}
	})

	t.Run("update time non-null", func(subT *testing.T) {
		bus := Bus{
			ID:        "test-create",
			UpdatedAt: time.Now(),
		}

		_, err := repo.CreateBus(bus)
		defer repo.DeleteBus(bus.ID)

		switch causeErr := errors.Cause(err); causeErr.(type) {
		case InvalidParameterError:
			assert.Equal(subT, "updated_at", causeErr.(InvalidParameterError).Name, "wrong invalid parameter name")
		default:
			assert.Fail(subT, "unexpected error", "%T: %[1]v", causeErr)
		}
	})

	t.Run("existing ID", func(subT *testing.T) {
		bus := Bus{ID: "test-create"}

		_, err := repo.CreateBus(bus)
		require.NoError(subT, err, "failed to create bus")
		defer repo.DeleteBus(bus.ID)

		_, err = repo.CreateBus(bus)
		switch causeErr := errors.Cause(err); causeErr.(type) {
		case DuplicateError:
			assert.Equal(subT, bus.ID, causeErr.(DuplicateError).ID, "wrong existing row ID")
		default:
			assert.Fail(subT, "unexpected error", "%T: %[1]v", causeErr)
		}
	})

	t.Run("success", func(subT *testing.T) {
		bus := Bus{
			ID:        "test-create",
			Latitude:  123,
			Longitude: 456,
		}

		createdBus, err := repo.CreateBus(bus)
		require.NoError(subT, err, "failed to create bus")
		defer repo.DeleteBus(bus.ID)

		assert.Equal(subT, bus.ID, createdBus.ID, "bus ID")
		assert.Equal(subT, bus.Latitude, createdBus.Latitude, "bus latitude")
		assert.Equal(subT, bus.Longitude, createdBus.Longitude, "bus longitude")
	})
}

func TestRepository_ReadBus(t *testing.T) {
	bus := Bus{ID: "test-read"}

	createdBus, err := repo.CreateBus(bus)
	if err != nil {
		t.Skipf("failed to create bus which would be read: %v", err)
	}
	defer repo.DeleteBus(bus.ID)

	t.Run("non-existing", func(subT *testing.T) {
		_, err = repo.ReadBus("non-existing")
		assert.EqualError(subT, errors.Cause(err), ErrNoSuchRow.Error())
	})

	t.Run("existing", func(subT *testing.T) {
		readBus, err := repo.ReadBus(bus.ID)
		require.NoError(subT, err, "failed to read bus")
		assert.Equal(subT, readBus.ID, createdBus.ID, "bus ID")
		assert.Equal(subT, readBus.Latitude, createdBus.Latitude, "bus latitude")
		assert.Equal(subT, readBus.Longitude, createdBus.Longitude, "bus longitude")
	})
}

func TestRepository_ReadAllBuses(t *testing.T) {
	t.Run("empty", func(subT *testing.T) {
		buses, err := repo.ReadAllBuses()
		require.NoError(subT, err, "failed to read all buses")
		assert.Empty(subT, buses)
	})

	t.Run("some buses", func(subT *testing.T) {
		nBuses := 2

		for i := 0; i < nBuses; i++ {
			bus := Bus{ID: fmt.Sprintf("test-readall-%v", i)}
			if _, err := repo.CreateBus(bus); err != nil {
				subT.Skipf("failed to create bus which would be read: %v", err)
			}
			defer repo.DeleteBus(bus.ID)
		}

		buses, err := repo.ReadAllBuses()
		require.NoError(subT, err, "failed to read all buses")
		assert.Len(subT, buses, nBuses)
	})
}

func TestRepository_UpdateBus(t *testing.T) {
	t.Run("missing ID", func(subT *testing.T) {
		bus := Bus{ID: "test-update"}

		if _, err := repo.CreateBus(bus); err != nil {
			subT.Skipf("failed to create bus which would be updated: %v", err)
		}
		defer repo.DeleteBus(bus.ID)

		bus.ID = ""

		_, err := repo.UpdateBus(bus)
		switch causeErr := errors.Cause(err); causeErr.(type) {
		case MissingParameterError:
			assert.Equal(subT, "id", causeErr.(MissingParameterError).Name, "bad missing parameter name")
		default:
			assert.Fail(subT, "unexpected error", "%T: %[1]v", causeErr)
		}
	})

	t.Run("non-existing ID", func(subT *testing.T) {
		bus := Bus{ID: "test-update"}

		if _, err := repo.CreateBus(bus); err != nil {
			subT.Skipf("failed to create bus which would be updated: %v", err)
		}
		defer repo.DeleteBus(bus.ID)

		bus.ID = "bar"

		_, err := repo.UpdateBus(bus)
		assert.EqualError(subT, errors.Cause(err), ErrNoSuchRow.Error())
	})

	t.Run("creation time non-null", func(subT *testing.T) {
		bus := Bus{ID: "test-update"}

		if _, err := repo.CreateBus(bus); err != nil {
			subT.Skipf("failed to create bus which would be updated: %v", err)
		}
		defer repo.DeleteBus(bus.ID)

		bus.CreatedAt = time.Now()

		_, err := repo.UpdateBus(bus)
		switch causeErr := errors.Cause(err); causeErr.(type) {
		case InvalidParameterError:
			assert.Equal(subT, "created_at", causeErr.(InvalidParameterError).Name, "wrong invalid parameter name")
		default:
			assert.Fail(subT, "unexpected error", "%T: %[1]v", causeErr)
		}
	})

	t.Run("update time non-null", func(subT *testing.T) {
		bus := Bus{ID: "test-update"}

		if _, err := repo.CreateBus(bus); err != nil {
			subT.Skipf("failed to create bus which would be updated: %v", err)
		}
		defer repo.DeleteBus(bus.ID)

		bus.UpdatedAt = time.Now()

		_, err := repo.UpdateBus(bus)
		switch causeErr := errors.Cause(err); causeErr.(type) {
		case InvalidParameterError:
			assert.Equal(subT, "updated_at", causeErr.(InvalidParameterError).Name, "wrong invalid parameter name")
		default:
			assert.Fail(subT, "unexpected error", "%T: %[1]v", causeErr)
		}
	})

	t.Run("success", func(subT *testing.T) {
		bus := Bus{ID: "test-update"}

		if _, err := repo.CreateBus(bus); err != nil {
			subT.Skipf("failed to create bus which would be updated: %v", err)
		}
		defer repo.DeleteBus(bus.ID)

		bus.Latitude = 1.23
		bus.Longitude = 4.56

		updatedBus, err := repo.UpdateBus(bus)
		require.NoError(subT, err, "failed to create bus")
		assert.Equal(subT, bus.ID, updatedBus.ID, "bus ID")
		assert.Equal(subT, bus.Latitude, updatedBus.Latitude, "bus latitude")
		assert.Equal(subT, bus.Longitude, updatedBus.Longitude, "bus longitude")
	})
}

func TestRepository_DeleteBus(t *testing.T) {
	bus := Bus{ID: "test-delete"}

	t.Run("existing", func(subT *testing.T) {
		if _, err := repo.CreateBus(bus); err != nil {
			t.Skipf("failed to create bus which would be deleted: %v", err)
		}

		err := repo.DeleteBus(bus.ID)
		assert.NoError(subT, err, "failed to delete bus")
	})

	t.Run("non-existing", func(subT *testing.T) {
		err := repo.DeleteBus("non-existing")
		assert.EqualError(subT, errors.Cause(err), ErrNoSuchRow.Error())
	})
}

func BenchmarkNewPostgresRepository(b *testing.B) {
	for n := 0; n < b.N; n++ {
		// don't use the global DB connection because we need to close it here
		benchRepo, err := newRepository()
		if err != nil {
			b.Fatal(err)
		}

		b.StopTimer()
		if err := benchRepo.Close(); err != nil {
			b.Error(err)
		}
		b.StartTimer()
	}
}

func BenchmarkRepository_Close(b *testing.B) {
	for n := 0; n < b.N; n++ {
		// don't use the global DB connection because we need to close it here
		b.StopTimer()
		benchRepo, err := newRepository()
		if err != nil {
			b.Fatal(err)
		}
		b.StartTimer()

		if err := benchRepo.Close(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRepository_CreateBus(b *testing.B) {
	bus := Bus{ID: "bench-create"}

	for n := 0; n < b.N; n++ {
		if _, err := repo.CreateBus(bus); err != nil {
			b.Fatal(err)
		}

		b.StopTimer()
		if err := repo.DeleteBus(bus.ID); err != nil {
			b.Error(err)
		}
		b.StartTimer()
	}
}

func BenchmarkRepository_ReadBus(b *testing.B) {
	bus := Bus{ID: "bench-read"}

	if _, err := repo.CreateBus(bus); err != nil {
		b.Skipf("failed to create bus which would be read: %v", err)
	}
	defer repo.DeleteBus(bus.ID)

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		if _, err := repo.ReadBus(bus.ID); err != nil {
			b.Error(err)
		}
	}

	// don't benchmark the deferred function
	b.StopTimer()
}

func BenchmarkRepository_ReadAllBuses(b *testing.B) {
	nBuses := []int{1, 2, 4, 8}

	for _, nb := range nBuses {
		b.Run(strconv.Itoa(nb), func(subB *testing.B) {
			for i := 0; i < nb; i++ {
				bus := Bus{ID: fmt.Sprintf("bench-readall-%v", i)}
				if _, err := repo.CreateBus(bus); err != nil {
					subB.Skipf("failed to create bus which would be read: %v", err)
				}
				defer repo.DeleteBus(bus.ID)
			}

			subB.ResetTimer()

			for n := 0; n < subB.N; n++ {
				if _, err := repo.ReadAllBuses(); err != nil {
					subB.Error(err)
				}
			}

			// don't benchmark the deferred functions
			subB.StopTimer()
		})
	}
}

func BenchmarkRepository_UpdateBus(b *testing.B) {
	bus := Bus{ID: "bench-update"}

	if _, err := repo.CreateBus(bus); err != nil {
		b.Skipf("failed to create bus which would be updated: %v", err)
	}
	defer repo.DeleteBus(bus.ID)

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		bus.Latitude = float64(n)
		bus.Longitude = float64(n)

		if _, err := repo.UpdateBus(bus); err != nil {
			b.Error(err)
		}
	}

	// don't benchmark the deferred function
	b.StopTimer()
}

func BenchmarkRepository_DeleteBus(b *testing.B) {
	bus := Bus{ID: "bench-delete"}

	for n := 0; n < b.N; n++ {
		b.StopTimer()
		if _, err := repo.CreateBus(bus); err != nil {
			b.Fatal(err)
		}
		b.StartTimer()

		if err := repo.DeleteBus(bus.ID); err != nil {
			b.Error(err)
		}
	}
}
