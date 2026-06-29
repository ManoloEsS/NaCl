package middleware

import "testing"

func TestRequestLogger(t *testing.T) {
	// TODO: next handler returns 200 → log contains method, path, 200, duration
	// TODO: next handler sets 404 → log contains 404
	// TODO: no status set → log contains 200 (default)
}
