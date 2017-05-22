package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequestScheme(t *testing.T) {
	t.Run("http", func(subT *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "http://www.test.com", nil)
		scheme := requestScheme(req)
		assert.Equal(subT, "http", scheme)
	})

	t.Run("https", func(subT *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "https://www.test.com", nil)
		scheme := requestScheme(req)
		assert.Equal(subT, "https", scheme)
	})
}
