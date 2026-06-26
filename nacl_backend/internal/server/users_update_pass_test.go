package server

import (
	"testing"
)

func TestHandleUpdateUserPassword(t *testing.T) {
	pool, queries := newTestDB(t)
	defer pool.Close()
	cleanupTestDB(t, pool, "users")

	server := newTestServer(t, queries)

	_ = server

	tests := []struct {
		name     string
		token    string
		body     string
		wantCode int
	}{
		// successful password update
		// unauthorized - no token
		// unauthorized - invalid token
		// wrong old password
		// empty old password
		// empty new password
		// invalid JSON
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TODO: implement
		})
	}
}
