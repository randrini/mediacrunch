package api

import (
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"github.com/mediacrunch/mediacrunch/internal/cache"
	"github.com/mediacrunch/mediacrunch/internal/compressor"
	"github.com/mediacrunch/mediacrunch/internal/db"
	"github.com/mediacrunch/mediacrunch/internal/logger"
	"github.com/mediacrunch/mediacrunch/internal/scanner"
)

var visitors = make(map[string]*rate.Limiter)
var mu sync.Mutex

func getLimiter(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()
	if v, exists := visitors[ip]; exists {
		return v
	}
	limiter := rate.NewLimiter(30, 10) // 30 requests/sec, burst 10
	visitors[ip] = limiter
	return limiter
}

func rateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		limiter := getLimiter(c.ClientIP())
		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// NewRouter creates and configures the Gin router with all API routes.
func NewRouter(database *db.DB) *gin.Engine {
	router := gin.Default()

	// CORS middleware for development
	router.Use(corsMiddleware())

	// Rate limiting middleware on API routes
	api := router.Group("/api")
	api.Use(rateLimitMiddleware())

	// Initialize cache, logger, and handlers
	cacheStore := cache.New(5 * time.Minute)
	log := logger.NewLogger(database.DB)
	scannerDispatcher := scanner.NewDispatcher(database, log)
	comp := compressor.NewCompressor(database, log, cacheStore)

	instanceHandler := NewInstanceHandler(database, scannerDispatcher, cacheStore)
	mediaHandler := NewMediaHandler(database, cacheStore)
	compressHandler := NewCompressHandler(database, comp, cacheStore, log)
	statsHandler := NewStatsHandler(database, cacheStore)
	cleanupHandler := NewCleanupHandler(database, cacheStore, log)

	// API routes
	{
		// Health
		api.GET("/health", HealthCheck)

		// Stats
		api.GET("/stats", statsHandler.OverallStats)

		// Instances
		api.POST("/instances", instanceHandler.CreateInstance)
		api.POST("/instances/test", instanceHandler.TestConnection)

		// Plex authentication (PIN-based OAuth)
		api.POST("/plex/pin", instanceHandler.CreatePlexPIN)
		api.GET("/plex/pin/:id", instanceHandler.CheckPlexPIN)
		api.GET("/instances", instanceHandler.ListInstances)
		api.GET("/instances/:id", instanceHandler.GetInstance)
		api.PUT("/instances/:id", instanceHandler.UpdateInstance)
		api.DELETE("/instances/:id", instanceHandler.DeleteInstance)
		api.POST("/instances/:id/scan", instanceHandler.ScanInstance)
		api.POST("/instances/:id/lock", instanceHandler.LockInstance)
		api.GET("/instances/:id/settings", instanceHandler.GetSettings)
		api.PUT("/instances/:id/settings", instanceHandler.UpdateSettings)

		// Per-instance stats
		api.GET("/instances/:id/stats", statsHandler.InstanceStats)

		// Media
		api.GET("/instances/:id/media", mediaHandler.ListMedia)

		// Compression
		api.GET("/compress", compressHandler.ListJobs)  // list recent jobs
		api.POST("/compress", compressHandler.StartCompression)
		api.GET("/compress/:id", compressHandler.GetJobStatus)
		api.POST("/compress/:id/cancel", compressHandler.CancelJob)
		api.GET("/compress/:id/results", compressHandler.GetJobResults)

		// Logs
		logHandler := NewLogHandler(database)
		api.GET("/logs", logHandler.GetLogs)
		api.DELETE("/logs", logHandler.ClearLogs)

		// Cleanup
		api.POST("/instances/:id/cleanup-backups", cleanupHandler.CleanupInstance)
		api.POST("/cleanup-backups", cleanupHandler.CleanupAll)
	}

	// Serve Vue SPA static files
	spaDir := os.Getenv("MC_SPA_DIR")
	if spaDir == "" {
		// Try common locations
		candidates := []string{
			"web/dist",
			"/app/web/dist",
		}
		for _, dir := range candidates {
			if _, err := os.Stat(filepath.Join(dir, "index.html")); err == nil {
				spaDir = dir
				break
			}
		}
	}

	if spaDir != "" {
		router.Use(SPARecovery(spaDir))
		// Serve static assets
		router.Static("/assets", filepath.Join(spaDir, "assets"))
		// Serve index.html for all non-API, non-asset routes (SPA fallback)
		router.NoRoute(func(c *gin.Context) {
			path := c.Request.URL.Path
			// Serve actual files if they exist (JS, CSS, images, etc.)
			fullPath := filepath.Join(spaDir, path)
			if _, err := os.Stat(fullPath); err == nil && path != "/" {
				c.File(fullPath)
				return
			}
			// Otherwise serve index.html for SPA routing
			c.File(filepath.Join(spaDir, "index.html"))
		})
	}

	return router
}

// SPARecovery returns middleware that serves the SPA index.html for HTML accept headers.
func SPARecovery(spaDir string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}

// corsMiddleware allows all origins for development.
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		c.Header("Access-Control-Max-Age", "86400")
		// Prevent caching of API responses so sort/filter changes are always fresh
		c.Header("Cache-Control", "no-store, no-cache, must-revalidate")
		c.Header("Pragma", "no-cache")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
