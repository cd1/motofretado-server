package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cd1/motofretado-server/web/jsonapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildMux(t *testing.T) {
	mux := BuildMux(repo)

	t.Run("URL not found", func(subT *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/foo", nil)

		mux.ServeHTTP(w, req)

		assert.Equal(subT, http.StatusNotFound, w.Code, "unexpected status code")
	})

	t.Run("method not allowed", func(subT *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("FOO", "/bus", nil)

		mux.ServeHTTP(w, req)

		assert.Equal(subT, http.StatusMethodNotAllowed, w.Code, "method FOO shouldn't be allowed")
	})

	t.Run("OPTIONS", func(subT *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodOptions, "/bus", nil)

		mux.ServeHTTP(w, req)

		// http.StatusNoContent (issue julienschmidt/httprouter #156)
		assert.Equal(subT, http.StatusOK, w.Code, "unexpected status code")
		assert.NotEmpty(subT, w.Header().Get("Allow"), "\"Allow\" header should be present in OPTIONS response")
	})

	t.Run("gzip", func(subT *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/bus", nil)
		req.Header.Set("Accept", jsonapi.ContentType)
		req.Header.Set("Accept-Encoding", "gzip")

		mux.ServeHTTP(w, req)

		require.Equal(subT, http.StatusOK, w.Code, "unexpected status code")
		assert.Equal(subT, w.Header().Get("Content-Encoding"), "gzip", "\"Content-Encoding\" header should contain the encoding \"gzip\"")
	})
}

func BenchmarkBuildMux(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_ = BuildMux(repo)
	}
}
