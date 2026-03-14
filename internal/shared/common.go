package shared

import "time"

const (
	LocationBangkok = "Asia/Bangkok"
)

const (
	AppEnvDev  = "DEV"
	AppEnvTest = "TEST"
	AppEnvProd = "PROD"
)

const (
	AuthorizationHeaderKey  = "authorization"
	AuthorizationTypeBearer = "bearer"
	XAccessTokenHeaderKey   = "X-Access-Token"
	RefreshTokenCookieKey   = "refresh_token"
	UserAgentKey            = "user-agent"
	ClientIPKey             = "client-ip"
	UserIDKey               = "user-id"
	RefreshTokenKey         = "refresh-token"
)

const (
	TestTokenSecretKey = "test-token-secret-key"

	// TestDatabaseConfig contains constants for test database setup
	TestPostgresImage = "postgres:18-alpine"
	TestDatabaseName  = "testdb"
	TestDatabaseUser  = "testuser"
	TestDatabasePass  = "testpass"
	TestAppName       = "trip-planner-test"
	TestSSLMode       = "disable"

	// TestRedisConfig contains constants for test Redis setup
	TestRedisImage          = "redis:7-alpine"
	TestRedisWaitLogMessage = "Ready to accept connections"

	// TestDatabaseConnectionPool contains connection pool settings for tests
	TestMaxOpenConns          = 10
	TestMaxIdleConns          = 5
	TestConnMaxLifetime       = 5 * time.Minute
	TestConnMaxLifetimeJitter = 30 * time.Second
	TestConnMaxIdleTime       = 5 * time.Minute
	TestHealthCheckPeriod     = 1 * time.Minute
	TestConnectTimeout        = 10 * time.Second
	TestEventTimeout          = 30 * time.Second

	// TestContainerConfig contains testcontainer-specific settings
	TestContainerWaitOccurrence = 2
	TestContainerStartupTimeout = 60 * time.Second
	TestContainerWaitLogMessage = "database system is ready to accept connections"

	// Test Timeout
	TestTimeout = 20 * time.Second
)
