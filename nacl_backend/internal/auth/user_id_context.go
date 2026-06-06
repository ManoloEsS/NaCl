package auth

import (
	"context"

	"github.com/google/uuid"
)

// contextKey is an unexported type to prevent collisions with context keys from other packages
type contextKey struct {
	name string
}

// userIDKey is an unexported pointer to a contextKey, ensuring no other package can use this same context key
var userIDKey = &contextKey{"userID"}

// WithUserID stores the user ID in the context, retrieved by UserIDFromContext.
func WithUserID(ctx context.Context, id uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDKey, id)
}

// UserIDFromContext retrieves the user ID from the context. Returns uuid.Nil and false if not present
func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	v, ok := ctx.Value(userIDKey).(uuid.UUID)
	return v, ok
}
