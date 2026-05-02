package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/guncv/ticket-reservation-server/internal/db"
)

type DebugHandler struct {
	pool *db.PgPool
}

func NewDebugHandler(pool *db.PgPool) *DebugHandler {
	return &DebugHandler{
		pool: pool,
	}
}

// PoolStats returns current database connection pool statistics
// GET /api/v1/debug/pool-stats
func (h *DebugHandler) PoolStats(c *gin.Context) {
	stats := h.pool.Stats()

	c.JSON(http.StatusOK, gin.H{
		"total_conns":        stats.TotalConns,
		"acquired_conns":     stats.AcquiredConns,
		"idle_conns":         stats.IdleConns,
		"max_conns":          stats.MaxConns,
		"acquire_count":      stats.AcquireCount,
		"acquire_duration_ms": stats.AcquireDuration.Milliseconds(),
		"empty_acquire_count": stats.EmptyAcquireCount,
		"canceled_acquires":  stats.CanceledAcquires,
		"constructing_conns": stats.ConstructingConns,
		"pool_utilization":   stats.PoolUtilization,
		"status":             poolStatus(stats),
	})
}

// poolStatus returns a human-readable status based on pool utilization
func poolStatus(stats db.PoolStats) string {
	switch {
	case stats.PoolUtilization >= 0.9:
		return "CRITICAL - pool nearly exhausted"
	case stats.PoolUtilization >= 0.7:
		return "WARNING - pool under heavy load"
	case stats.PoolUtilization >= 0.5:
		return "MODERATE - pool moderately utilized"
	default:
		return "HEALTHY - pool has capacity"
	}
}
