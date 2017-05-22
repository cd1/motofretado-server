package web

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/cd1/motofretado-server/data"
	"github.com/cd1/motofretado-server/web/jsonapi"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var busesHandler BusesHandler

func TestBusesHandler_get(t *testing.T) {
	subTestFunc := func(header http.Header, expectedStatus int) func(*testing.T) {
		return func(subT *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/bus", nil)
			req.Header = header

			w := httptest.NewRecorder()
			var params httprouter.Params

			busesHandler.get(w, req, params)
			require.Equal(subT, expectedStatus, w.Code, "invalid HTTP status")

			if expectedStatus == http.StatusOK {
				var doc jsonapi.BusesDocument

				err := json.NewDecoder(w.Body).Decode(&doc)
				require.NoError(subT, err, "failed to decode data from JSON")

				buses, err := jsonapi.FromBusesDocument(doc)
				require.NoError(subT, err, "failed to convert JSONAPI data")

				assert.Len(subT, buses, busesCount, "unexpected number of buses created for test")
			}
		}
	}

	h := make(http.Header)

	t.Run("not acceptable", subTestFunc(h, http.StatusNotAcceptable))

	h.Set("Accept", jsonapi.ContentType)
	t.Run("success", subTestFunc(h, http.StatusOK))
}

func TestBusesHandler_post(t *testing.T) {
	subTestFunc := func(bus data.Bus, body io.Reader, header http.Header, expectedStatus int, deleteOnExit bool) func(*testing.T) {
		return func(subT *testing.T) {
			if body == nil {
				var buf bytes.Buffer

				doc := jsonapi.ToBusDocument(bus)
				if err := json.NewEncoder(&buf).Encode(doc); err != nil {
					subT.Skipf("failed to encode data to JSON: %v", err)
				}

				body = &buf
			}

			req := httptest.NewRequest(http.MethodPost, "/bus", body)
			req.Header = header

			w := httptest.NewRecorder()
			var params httprouter.Params

			busesHandler.post(w, req, params)
			if deleteOnExit {
				defer repo.DeleteBus(bus.ID)
			}

			require.Equal(subT, expectedStatus, w.Code, "invalid HTTP status")

			if expectedStatus == http.StatusCreated {
				assert.NotEmpty(subT, w.HeaderMap.Get("Location"), "\"Location\" header should be set")

				var createdBus data.Bus
				var createdBusDoc jsonapi.BusDocument

				err := json.NewDecoder(w.Body).Decode(&createdBusDoc)
				require.NoError(subT, err, "failed to decode data from JSON")

				createdBus, err = jsonapi.FromBusDocument(createdBusDoc)
				require.NoError(subT, err, "failed to convert JSONAPI data")
				assert.Equal(subT, bus.ID, createdBus.ID, "unexpected bus ID")
				assert.Equal(subT, bus.Latitude, createdBus.Latitude, "unexpected bus latitude")
				assert.Equal(subT, bus.Longitude, createdBus.Longitude, "unexpected bus longitude")
			}
		}
	}

	var bus data.Bus

	h := make(http.Header)
	t.Run("not acceptable",
		subTestFunc(bus, nil, h, http.StatusNotAcceptable, true))

	h.Set("Accept", jsonapi.ContentType)
	t.Run("unsupported media type",
		subTestFunc(bus, nil, h, http.StatusUnsupportedMediaType, true))

	h.Set("Content-Type", jsonapi.ContentType)
	t.Run("invalid JSON format",
		subTestFunc(bus, strings.NewReader("foo bar {{{"), h, http.StatusBadRequest, true))

	doc := jsonapi.ToBusDocument(bus)
	doc.JSONAPI.Version = "foo"
	t.Run("bad JSONAPI version", func(subT *testing.T) {
		var buf bytes.Buffer

		if err := json.NewEncoder(&buf).Encode(doc); err != nil {
			t.Skipf("failed to encode data to JSON: %v", err)
		}
		subTestFunc(bus, &buf, h, http.StatusBadRequest, true)(subT)
	})

	doc = jsonapi.ToBusDocument(bus)
	doc.Data.Type = "foo"
	t.Run("bad JSONAPI data type", func(subT *testing.T) {
		var buf bytes.Buffer

		if err := json.NewEncoder(&buf).Encode(doc); err != nil {
			t.Skipf("failed to encode data to JSON: %v", err)
		}
		subTestFunc(bus, &buf, h, http.StatusConflict, true)(subT)
	})

	bus.Latitude = 1.23
	bus.Longitude = 4.56
	t.Run("missing ID",
		subTestFunc(bus, nil, h, http.StatusUnprocessableEntity, true))

	bus.ID = "foo"
	bus.CreatedAt = time.Now()
	t.Run("creation time specified",
		subTestFunc(bus, nil, h, http.StatusUnprocessableEntity, true))

	bus.CreatedAt = time.Time{}
	t.Run("success",
		subTestFunc(bus, nil, h, http.StatusCreated, false))

	t.Run("ID already exists",
		subTestFunc(bus, nil, h, http.StatusConflict, true))
}

func BenchmarkBusesHandler_get(b *testing.B) {
	subBenchFunc := func(method string) func(*testing.B) {
		return func(subB *testing.B) {
			req := httptest.NewRequest(method, "/bus", nil)
			req.Header.Set("Accept", jsonapi.ContentType)

			w := httptest.NewRecorder()
			var params httprouter.Params

			subB.ResetTimer()
			for n := 0; n < subB.N; n++ {
				busesHandler.get(w, req, params)
				if expectedStatus := http.StatusOK; w.Code != expectedStatus {
					subB.Errorf("unexpected HTTP status; got = %v, want = %v",
						w.Code, expectedStatus)
				}
			}
		}
	}

	for _, method := range []string{http.MethodGet, http.MethodHead} {
		b.Run(method, subBenchFunc(method))
	}
}

func BenchmarkBusesHandler_post(b *testing.B) {
	var body bytes.Buffer

	bus := data.Bus{ID: "bench-post"}
	doc := jsonapi.ToBusDocument(bus)

	w := httptest.NewRecorder()
	var params httprouter.Params

	b.ResetTimer()
	b.StopTimer()
	for n := 0; n < b.N; n++ {
		if err := json.NewEncoder(&body).Encode(doc); err != nil {
			b.Skipf("failed to encode data to JSON: %v", err)
		}

		req := httptest.NewRequest(http.MethodPost, "/bus", &body)
		req.Header.Set("Accept", jsonapi.ContentType)
		req.Header.Set("Content-Type", jsonapi.ContentType)

		b.StartTimer()
		busesHandler.post(w, req, params)
		if expectedStatus := http.StatusCreated; w.Code != expectedStatus {
			b.Errorf("unexpected HTTP status; got = %v, want = %v",
				w.Code, expectedStatus)
		}
		b.StopTimer()

		if err := repo.DeleteBus(bus.ID); err != nil {
			b.Error(err)
		}
	}
}
