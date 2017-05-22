package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/cd1/motofretado-server/data"
	"github.com/cd1/motofretado-server/web/jsonapi"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/require"
)

var busHandler BusHandler

func TestBusHandler_doDelete(t *testing.T) {
	subTestFunc := func(id string, expectedStatus int) func(*testing.T) {
		return func(subT *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/bus/%v", id), nil)

			w := httptest.NewRecorder()
			params := httprouter.Params{
				{
					Key:   "id",
					Value: id,
				},
			}

			busHandler.doDelete(w, req, params)
			require.Equal(subT, expectedStatus, w.Code, "unexpected HTTP status code")
		}
	}

	t.Run("empty ID", subTestFunc("", http.StatusBadRequest))

	t.Run("not found", subTestFunc("not-found", http.StatusNotFound))

	t.Run("success", func(subT *testing.T) {
		bus := data.Bus{ID: "test-delete"}
		if _, err := repo.CreateBus(bus); err != nil {
			subT.Skipf("failed to create bus which would be deleted: %v", err)
		}

		subTestFunc(bus.ID, http.StatusNoContent)(subT)
	})
}

func TestBusHandler_get(t *testing.T) {
	subTestFunc := func(id string, header http.Header, expectedStatus int) func(*testing.T) {
		return func(subT *testing.T) {
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/bus/%v", id), nil)
			req.Header = header

			w := httptest.NewRecorder()
			params := httprouter.Params{
				{
					Key:   "id",
					Value: id,
				},
			}

			busHandler.get(w, req, params)
			require.Equal(subT, expectedStatus, w.Code, "unexpected HTTP status code")
		}
	}

	id := "initial-bus-0"
	h := make(http.Header)

	t.Run("empty ID", subTestFunc("", h, http.StatusBadRequest))

	t.Run("not acceptable", subTestFunc(id, h, http.StatusNotAcceptable))

	h.Set("Accept", jsonapi.ContentType)
	t.Run("not found", subTestFunc("not-found", h, http.StatusNotFound))

	t.Run("success", subTestFunc(id, h, http.StatusOK))
}

func TestBusHandler_patch(t *testing.T) {
	subTestFunc := func(bus data.Bus, body io.Reader, header http.Header, expectedStatus int, create bool) func(*testing.T) {
		return func(subT *testing.T) {
			if create {
				busToCreate := data.Bus{ID: bus.ID}

				_, err := repo.CreateBus(busToCreate)
				if err != nil {
					subT.Skipf("failed to create bus which would be updated: %v", err)
				}
				defer repo.DeleteBus(bus.ID)
			}

			if body == nil {
				var buf bytes.Buffer

				doc := jsonapi.ToBusDocument(bus)
				if err := json.NewEncoder(&buf).Encode(doc); err != nil {
					subT.Skipf("failed to encode bus to JSON: %v", err)
				}

				body = &buf
			}

			req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/bus/%v", bus.ID), body)
			req.Header = header

			w := httptest.NewRecorder()
			params := httprouter.Params{
				{
					Key:   "id",
					Value: bus.ID,
				},
			}

			busHandler.patch(w, req, params)

			require.Equal(subT, expectedStatus, w.Code, "unexpected HTTP status code")
			if expectedStatus == http.StatusOK {
				var doc jsonapi.BusDocument

				err := json.NewDecoder(w.Body).Decode(&doc)
				require.NoError(subT, err, "failed to decode data from JSON")

				_, err = jsonapi.FromBusDocument(doc)
				require.NoError(subT, err, "failed to convert data from JSONAPI")
			}
		}
	}

	var bus data.Bus

	h := make(http.Header)

	t.Run("empty ID",
		subTestFunc(bus, nil, h, http.StatusBadRequest, false))

	bus.ID = "test-patch"
	t.Run("not acceptable",
		subTestFunc(bus, nil, h, http.StatusNotAcceptable, false))

	h.Set("Accept", jsonapi.ContentType)
	t.Run("unsupported media type",
		subTestFunc(bus, nil, h, http.StatusUnsupportedMediaType, false))

	h.Set("Content-Type", jsonapi.ContentType)
	t.Run("invalid JSON format",
		subTestFunc(bus, strings.NewReader("foo bar {{{"), h, http.StatusBadRequest, false))

	t.Run("invalid JSONAPI version", func(subT *testing.T) {
		doc := jsonapi.ToBusDocument(bus)
		doc.JSONAPI.Version = "foo"

		var buf bytes.Buffer

		if err := json.NewEncoder(&buf).Encode(doc); err != nil {
			subT.Skipf("failed to encode data from JSON: %v", err)
		}

		subTestFunc(bus, &buf, h, http.StatusBadRequest, false)(subT)
	})

	t.Run("invalid JSONAPI data type", func(subT *testing.T) {
		doc := jsonapi.ToBusDocument(bus)
		doc.Data.Type = "foo"

		var buf bytes.Buffer

		if err := json.NewEncoder(&buf).Encode(doc); err != nil {
			subT.Skipf("failed to encode data from JSON: %v", err)
		}

		subTestFunc(bus, &buf, h, http.StatusConflict, false)(subT)
	})

	t.Run("not found",
		subTestFunc(bus, nil, h, http.StatusNotFound, false))

	t.Run("different IDs", func(subT *testing.T) {
		doc := jsonapi.ToBusDocument(bus)
		doc.Data.ID = "foo"

		var buf bytes.Buffer

		if err := json.NewEncoder(&buf).Encode(doc); err != nil {
			subT.Skipf("failed to encode data from JSON: %v", err)
		}

		subTestFunc(bus, &buf, h, http.StatusBadRequest, false)(subT)
	})

	bus.ID = "test-patch"
	bus.CreatedAt = time.Now()
	t.Run("creation time specified",
		subTestFunc(bus, nil, h, http.StatusUnprocessableEntity, true))

	bus.CreatedAt = time.Time{}
	t.Run("success",
		subTestFunc(bus, nil, h, http.StatusOK, true))
}

func BenchmarkBusHandler_doDelete(b *testing.B) {
	bus := data.Bus{ID: "bench-delete"}

	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/bus/%v", bus.ID), nil)

	w := httptest.NewRecorder()
	params := httprouter.Params{
		{
			Key:   "id",
			Value: bus.ID,
		},
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		b.StopTimer()
		if _, err := repo.CreateBus(bus); err != nil {
			b.Fatal(err)
		}
		b.StartTimer()

		busHandler.doDelete(w, req, params)
		if expectedStatus := http.StatusNoContent; w.Code != expectedStatus {
			b.Errorf("unexpected HTTP status; got = %v, want = %v", w.Code, expectedStatus)
		}
	}
}

func BenchmarkBusHandler_get(b *testing.B) {
	subBenchFunc := func(method string) func(*testing.B) {
		return func(subB *testing.B) {
			busID := "initial-bus-0"
			req := httptest.NewRequest(method, fmt.Sprintf("/bus/%v", busID), nil)
			req.Header.Set("Accept", jsonapi.ContentType)

			w := httptest.NewRecorder()
			params := httprouter.Params{
				{
					Key:   "id",
					Value: busID,
				},
			}

			subB.ResetTimer()
			for n := 0; n < subB.N; n++ {
				busHandler.get(w, req, params)
				if expectedStatus := http.StatusOK; w.Code != expectedStatus {
					subB.Errorf("unexpected HTTP status; got = %v, want = %v", w.Code, expectedStatus)
				}
			}
		}
	}

	for _, method := range []string{http.MethodGet, http.MethodHead} {
		b.Run(method, subBenchFunc(method))
	}
}

func BenchmarkBusHandler_patch(b *testing.B) {
	bus := data.Bus{ID: "bench-patch"}
	if _, err := repo.CreateBus(bus); err != nil {
		b.Fatal(err)
	}
	defer repo.DeleteBus(bus.ID)

	w := httptest.NewRecorder()
	params := httprouter.Params{
		{
			Key:   "id",
			Value: bus.ID,
		},
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		b.StopTimer()
		updatedBus := bus
		updatedBus.Latitude = float64(n)

		doc := jsonapi.ToBusDocument(updatedBus)

		var buf bytes.Buffer

		if err := json.NewEncoder(&buf).Encode(doc); err != nil {
			b.Fatal(err)
		}

		req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/bus/%v", bus.ID), &buf)
		req.Header.Set("Accept", jsonapi.ContentType)
		req.Header.Set("Content-Type", jsonapi.ContentType)
		b.StartTimer()

		busHandler.patch(w, req, params)
		if expectedStatus := http.StatusOK; w.Code != expectedStatus {
			b.Errorf("unexpected HTTP status; got = %v, want = %v", w.Code, expectedStatus)
		}
	}
}
