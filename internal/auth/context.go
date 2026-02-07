package auth

// ContextKey is a custom type for context keys to avoid collisions
type ContextKey string

const (
	// ClaimsContextKey is the key used to store token claims in the request context
	ClaimsContextKey ContextKey = "claims"
)
