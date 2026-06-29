package middleware

import "testing"

func TestExtractToken(t *testing.T) {
	// TODO: "Bearer mytoken" → "mytoken", nil
	// TODO: no Authorization header → "", ErrNoAuthHeader
	// TODO: "Basic mytoken" → "", ErrNoBearerToken
	// TODO: "Bearer a b" → "", ErrInvalidBearerToken
}

func TestTokenValidator(t *testing.T) {
	// TODO: valid JWT → next handler called, 200
	// TODO: no token → 401
	// TODO: invalid token → 401
	// TODO: expired token → 401
}
