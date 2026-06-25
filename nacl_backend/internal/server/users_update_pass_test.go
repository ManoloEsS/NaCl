package server

import (
	"testing"
)

func TestHandleUpdateUserPassword(t *testing.T) {
	testDB := newTestDB(t)
	defer testDB.Close()
	cleanupTestDB(t, testDB, "users")

	server := newTestServer(t, testDB)

	_ = server
	_ = testDB

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
