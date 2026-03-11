package redis

import "time"

const (
	RedisPrefixGoogleAuthState = "google_auth_state:"
)

const (
	RedisTTLOAuthState = 5 * time.Minute
)
