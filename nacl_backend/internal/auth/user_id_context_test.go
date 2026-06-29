package auth

import (
	"testing"
)

func TestUserIDContext(t *testing.T) {
	// TODO: store userID, retrieve → returns same ID, ok=true
	// TODO: retrieve from empty context → returns Nil, ok=false
	// TODO: retrieve with wrong type in context → returns Nil, ok=false
}
