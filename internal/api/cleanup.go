package api

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/mediacrunch/mediacrunch/internal/cache"
	"github.com/mediacrunch/mediacrunch/internal/db"
	"github.com/mediacrunch/mediacrunch/internal/logger"
	"github.com/mediacrunch/mediacrunch/internal/models"
	"github.com/mediacrunch/mediacrunch/internal/util"
)

// CleanupHandler handles .bak backup file cleanup endpoints.
type CleanupHandler struct {
	DB     *db.DB
	Cache  *cache.Cache
	Logger *logger.Logger
}

// NewCleanupHandler creates a new CleanupHandler.
func NewCleanupHandler(database *db.DB, c *cache.Cache, log *logger.Logger) *CleanupHandler {
	return &CleanupHandler{
		DB:     database,
		Cache:  c,
		Logger: log,
	}
}

// bakFileInfo describes a single .bak file found during a directory walk.
type bakFileInfo struct {
	Path string
	Size int64
}

// cleanupResult is the per-instance result of a cleanup operation.
type cleanupResult struct {
	InstanceID   string   `json:"instance_id"`
	InstanceName string   `json:"instance_name"`
	InstanceType string   `json:"instance_type"`
	DryRun       bool     `json:"dry_run"`
	DeletedFiles int      `json:"deleted_files"`
	FreedBytes   int64    `json:"freed_bytes"`
	Errors       []string `json:"errors"`
}

// CleanupInstance handles POST /api/instances/:id/cleanup-backups
func (h *CleanupHandler) CleanupInstance(c *gin.Context) {
	id := c.Param("id")

	var input struct {
		DryRun bool `json:"dry_run"`
	}
	_ = c.ShouldBindJSON(&input)

	inst, err := h.getInstanceByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "instance not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	result := h.cleanupInstance(inst, input.DryRun)
	c.JSON(http.StatusOK, result)
}

// CleanupAll handles POST /api/cleanup-backups
func (h *CleanupHandler) CleanupAll(c *gin.Context) {
	var input struct {
		DryRun bool `json:"dry_run"`
	}
	_ = c.ShouldBindJSON(&input)

	rows, err := h.DB.Query(`SELECT id, type, name, host, api_key, path_prefix, settings, created_at FROM instances`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var results []cleanupResult
	var totalDeleted int
	var totalFreedBytes int64

	for rows.Next() {
		var inst models.Instance
		var createdAt string
		if err := rows.Scan(&inst.ID, &inst.Type, &inst.Name, &inst.Host, &inst.APIKey, &inst.PathPrefix, &inst.SettingsJSON, &createdAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		inst.CreatedAt = util.ParseTimestamp(createdAt)
		if err := inst.UnmarshalSettings(); err != nil {
			fmt.Printf("WARN: unmarshal settings for instance %s: %v\n", inst.ID, err)
		}

		result := h.cleanupInstance(inst, input.DryRun)
		results = append(results, result)
		totalDeleted += result.DeletedFiles
		totalFreedBytes += result.FreedBytes
	}

	if results == nil {
		results = []cleanupResult{}
	}

	c.JSON(http.StatusOK, gin.H{
		"dry_run":           input.DryRun,
		"results":           results,
		"total_deleted":     totalDeleted,
		"total_freed_bytes": totalFreedBytes,
	})
}

// cleanupInstance scans one instance's metadata directories for .bak* files and
// deletes them (or counts them when dryRun is true).
func (h *CleanupHandler) cleanupInstance(inst models.Instance, dryRun bool) cleanupResult {
	result := cleanupResult{
		InstanceID:   inst.ID,
		InstanceName: inst.Name,
		InstanceType: inst.Type,
		DryRun:       dryRun,
		Errors:       []string{},
	}

	root := scanRootForInstance(inst)
	if root == "" {
		return result
	}

	// Nothing to clean if the metadata root doesn't exist
	if _, err := os.Stat(root); err != nil {
		return result
	}

	files, err := walkBakFiles(root)
	if err != nil {
		msg := fmt.Sprintf("walk %s: %v", root, err)
		result.Errors = append(result.Errors, msg)
		h.Logger.Errorf("cleanup", inst.ID, "%s", msg)
		return result
	}

	for _, f := range files {
		result.FreedBytes += f.Size
		if dryRun {
			result.DeletedFiles++
			continue
		}
		if err := os.Remove(f.Path); err != nil {
			msg := fmt.Sprintf("could not delete %s: %v", f.Path, err)
			result.Errors = append(result.Errors, msg)
			h.Logger.Errorf("cleanup", inst.ID, "%s", msg)
			continue
		}
		result.DeletedFiles++
	}

	// Invalidate cache for this instance
	h.Cache.Invalidate("media:" + inst.ID)
	h.Cache.Invalidate("stats:" + inst.ID)

	h.Logger.Infof("cleanup", inst.ID, "Cleaned up %d .bak files (%s)", result.DeletedFiles, models.FormatSize(result.FreedBytes))

	return result
}

// scanRootForInstance determines the directory to walk for a given instance type.
func scanRootForInstance(inst models.Instance) string {
	switch inst.Type {
	case "plex":
		return filepath.Join(inst.PathPrefix, "config", "Metadata")
	case "radarr", "sonarr":
		return filepath.Join(inst.PathPrefix, "MediaCover")
	default:
		return ""
	}
}

// walkBakFiles walks a directory tree and returns all files matching the
// *.bak* pattern (same filter used by the scanner) with their paths and sizes.
func walkBakFiles(root string) ([]bakFileInfo, error) {
	var files []bakFileInfo
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			// Skip entries that can't be read (e.g. permission denied) and continue
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if !strings.Contains(d.Name(), ".bak") {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		files = append(files, bakFileInfo{Path: path, Size: info.Size()})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

// getInstanceByID fetches a single instance from the database by ID.
func (h *CleanupHandler) getInstanceByID(id string) (models.Instance, error) {
	var inst models.Instance
	var createdAt string
	err := h.DB.QueryRow(`
		SELECT id, type, name, host, api_key, path_prefix, settings, created_at FROM instances WHERE id = ?
	`, id).Scan(&inst.ID, &inst.Type, &inst.Name, &inst.Host, &inst.APIKey, &inst.PathPrefix, &inst.SettingsJSON, &createdAt)
	if err != nil {
		return inst, err
	}
	inst.CreatedAt = util.ParseTimestamp(createdAt)
	if err := inst.UnmarshalSettings(); err != nil {
		fmt.Printf("WARN: unmarshal settings for instance %s: %v\n", inst.ID, err)
	}
	return inst, nil
}
