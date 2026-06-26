package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandleIndex(t *testing.T) {
	pool, queries := newTestDB(t)
	defer pool.Close()

	server := newTestServer(t, queries)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	server.HTTPServer.Handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "this is root", rr.Body.String())
}
