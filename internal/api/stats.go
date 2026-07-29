package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mediacrunch/mediacrunch/internal/cache"
	"github.com/mediacrunch/mediacrunch/internal/db"
	"github.com/mediacrunch/mediacrunch/internal/version"
)

// StatsHandler handles statistics-related API endpoints.
type StatsHandler struct {
	DB    *db.DB
	Cache *cache.Cache
}

// NewStatsHandler creates a new StatsHandler.
func NewStatsHandler(database *db.DB, c *cache.Cache) *StatsHandler {
	return &StatsHandler{DB: database, Cache: c}
}

// OverallStats handles GET /api/stats
func (h *StatsHandler) OverallStats(c *gin.Context) {
	var totalInstances, totalItems, totalImages int
	var totalSize, totalSavings sql.NullInt64

	_ = h.DB.QueryRow(`SELECT COUNT(*) FROM instances`).Scan(&totalInstances)
	_ = h.DB.QueryRow(`SELECT COUNT(*), COALESCE(SUM(total_size), 0) FROM media_items`).Scan(&totalItems, &totalSize)
	_ = h.DB.QueryRow(`SELECT COALESCE(SUM(total_images), 0) FROM media_items`).Scan(&totalImages)
	_ = h.DB.QueryRow(`SELECT COALESCE(SUM(saved_bytes), 0) FROM compression_results WHERE status = 'compressed'`).Scan(&totalSavings)

	c.JSON(http.StatusOK, gin.H{
		"total_instances": totalInstances,
		"total_items":     totalItems,
		"total_images":    totalImages,
		"total_size":      totalSize.Int64,
		"total_savings":   totalSavings.Int64,
	})
}

// InstanceStats handles GET /api/instances/:id/stats
func (h *StatsHandler) InstanceStats(c *gin.Context) {
	id := c.Param("id")

	cacheKey := "stats:" + id
	if cached, ok := h.Cache.Get(cacheKey); ok {
		c.JSON(http.StatusOK, cached)
		return
	}

	// Verify instance exists
	var count int
	err := h.DB.QueryRow(`SELECT COUNT(*) FROM instances WHERE id = ?`, id).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "instance not found"})
		return
	}

	var totalItems, totalImages, compressedItems int
	var totalSize, totalSavings sql.NullInt64

	_ = h.DB.QueryRow(`SELECT COUNT(*), COALESCE(SUM(total_size), 0) FROM media_items WHERE instance_id = ?`, id).Scan(&totalItems, &totalSize)
	_ = h.DB.QueryRow(`SELECT COALESCE(SUM(total_images), 0) FROM media_items WHERE instance_id = ?`, id).Scan(&totalImages)
	_ = h.DB.QueryRow(`SELECT COUNT(*) FROM media_items WHERE instance_id = ? AND compressed = 1`, id).Scan(&compressedItems)
	_ = h.DB.QueryRow(`
		SELECT COALESCE(SUM(cr.saved_bytes), 0) FROM compression_results cr
		JOIN compression_jobs cj ON cr.job_id = cj.id
		WHERE cj.instance_id = ? AND cr.status = 'compressed'
	`, id).Scan(&totalSavings)

	response := gin.H{
		"instance_id":      id,
		"total_items":      totalItems,
		"total_images":     totalImages,
		"total_size":       totalSize.Int64,
		"compressed_items": compressedItems,
		"total_savings":    totalSavings.Int64,
	}

	h.Cache.Set(cacheKey, response)
	c.JSON(http.StatusOK, response)
}

// HealthCheck handles GET /api/health
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"version": version.Version,
		"commit":  version.Commit,
		"built":   version.BuildDate,
	})
}
