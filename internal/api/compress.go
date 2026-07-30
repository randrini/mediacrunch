package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/mediacrunch/mediacrunch/internal/cache"
	"github.com/mediacrunch/mediacrunch/internal/compressor"
	"github.com/mediacrunch/mediacrunch/internal/db"
	"github.com/mediacrunch/mediacrunch/internal/logger"
	"github.com/mediacrunch/mediacrunch/internal/models"
	"github.com/mediacrunch/mediacrunch/internal/util"
)

// CompressHandler handles compression-related API endpoints.
type CompressHandler struct {
	DB         *db.DB
	Compressor *compressor.Compressor
	Cache      *cache.Cache
	Logger     *logger.Logger
	jobs       map[string]context.CancelFunc // track running jobs for cancellation
	mu         sync.Mutex
}

// NewCompressHandler creates a new CompressHandler.
func NewCompressHandler(database *db.DB, comp *compressor.Compressor, c *cache.Cache, log *logger.Logger) *CompressHandler {
	return &CompressHandler{
		DB:         database,
		Compressor: comp,
		Cache:      c,
		Logger:     log,
		jobs:       make(map[string]context.CancelFunc),
	}
}

// StartCompression handles POST /api/compress
func (h *CompressHandler) StartCompression(c *gin.Context) {
	var input struct {
		InstanceID   string         `json:"instance_id" binding:"required"`
		MediaItemIDs []string       `json:"media_item_ids"`
		Quality      map[string]int `json:"quality"`
		MaxWidth     map[string]int `json:"max_width"`
		MinSizeKB    map[string]int64 `json:"min_size_kb"`
		Backup       bool           `json:"backup"`
		MinSavingKB  int64          `json:"min_saving_kb"`
		LockPlex     bool           `json:"lock_plex"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify instance exists
	var count int
	err := h.DB.QueryRow(`SELECT COUNT(*) FROM instances WHERE id = ?`, input.InstanceID).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "instance not found"})
		return
	}

	// Fetch instance settings for merging
	var settingsJSON string
	if err := h.DB.QueryRow(`SELECT settings FROM instances WHERE id = ?`, input.InstanceID).Scan(&settingsJSON); err != nil {
		fmt.Printf("WARN: could not load instance settings: %v\n", err)
		settingsJSON = "{}"
	}
	var instanceSettings models.InstanceSettings
	if err := json.Unmarshal([]byte(settingsJSON), &instanceSettings); err != nil {
		fmt.Printf("WARN: could not parse instance settings: %v\n", err)
	}

	// Build job config with instance settings as defaults
	config := models.JobConfig{
		Quality:     input.Quality,
		MaxWidth:    input.MaxWidth,
		MinSizeKB:   input.MinSizeKB,
		Backup:      input.Backup,
		MinSavingKB: input.MinSavingKB,
		LockPlex:    input.LockPlex,
	}
	config = mergeSettings(config, instanceSettings)

	configJSON, err := json.Marshal(config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("marshal config: %v", err)})
		return
	}

	// Determine total items
	var totalItems int
	if len(input.MediaItemIDs) > 0 {
		totalItems = len(input.MediaItemIDs)
	} else {
		err = h.DB.QueryRow(`SELECT COUNT(*) FROM media_items WHERE instance_id = ?`, input.InstanceID).Scan(&totalItems)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	job := models.CompressionJob{
		ID:           uuid.NewString(),
		InstanceID:   input.InstanceID,
		MediaItemIDs: input.MediaItemIDs,
		Status:       "pending",
		ConfigJSON:   string(configJSON),
		Config:       config,
		TotalItems:   totalItems,
		CreatedAt:    time.Now(),
	}

	_, err = h.DB.Exec(`
		INSERT INTO compression_jobs (id, instance_id, status, config, total_items, processed_items, total_images, processed_images, saved_bytes, error_count, skip_count, created_at)
		VALUES (?, ?, ?, ?, ?, 0, 0, 0, 0, 0, 0, ?)
	`, job.ID, job.InstanceID, job.Status, job.ConfigJSON, job.TotalItems, job.CreatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("create job: %v", err)})
		return
	}

	// Start compression in background
	ctx, cancel := context.WithCancel(context.Background())
	h.mu.Lock()
	h.jobs[job.ID] = cancel
	h.mu.Unlock()

	h.Logger.Infof("api", input.InstanceID, "Compression job %s created for instance %s (%d items)", job.ID, input.InstanceID, totalItems)

	go func() {
		defer func() {
			h.mu.Lock()
			delete(h.jobs, job.ID)
			h.mu.Unlock()
		}()
		h.Compressor.RunCompressionJob(ctx, &job)
	}()

	c.JSON(http.StatusCreated, job)

	// Invalidate cache for this instance
	h.Cache.Invalidate("media:" + input.InstanceID)
	h.Cache.Invalidate("stats:" + input.InstanceID)
}

// ListJobs handles GET /api/compress — list recent compression jobs.
func (h *CompressHandler) ListJobs(c *gin.Context) {
	limit := 20
	if l := c.Query("limit"); l != "" {
		if v, err := fmt.Sscanf(l, "%d", &limit); err != nil || v != 1 || limit < 1 {
			limit = 20
		}
	}
	instanceID := c.Query("instance_id")

	var rows *sql.Rows
	var err error
	if instanceID != "" {
		rows, err = h.DB.Query(`
			SELECT id, instance_id, status, config, total_items, processed_items, total_images, processed_images, saved_bytes, error_count, skip_count, created_at, started_at, completed_at
			FROM compression_jobs WHERE instance_id = ? ORDER BY created_at DESC LIMIT ?
		`, instanceID, limit)
	} else {
		rows, err = h.DB.Query(`
			SELECT id, instance_id, status, config, total_items, processed_items, total_images, processed_images, saved_bytes, error_count, skip_count, created_at, started_at, completed_at
			FROM compression_jobs ORDER BY created_at DESC LIMIT ?
		`, limit)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	jobs := make([]models.CompressionJob, 0)
	for rows.Next() {
		var job models.CompressionJob
		var createdAt, startedAt, completedAt sql.NullString

		if err := rows.Scan(&job.ID, &job.InstanceID, &job.Status, &job.ConfigJSON,
			&job.TotalItems, &job.ProcessedItems, &job.TotalImages, &job.ProcessedImages,
			&job.SavedBytes, &job.ErrorCount, &job.SkipCount,
			&createdAt, &startedAt, &completedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		job.UnmarshalConfig()

		job.CreatedAt = util.ParseTimestamp(createdAt.String)
		job.StartedAt = util.ParseTimestampPtr(startedAt.String)
		job.CompletedAt = util.ParseTimestampPtr(completedAt.String)

		jobs = append(jobs, job)
	}

	c.JSON(http.StatusOK, jobs)
}

// GetJobStatus handles GET /api/compress/:id
func (h *CompressHandler) GetJobStatus(c *gin.Context) {
	id := c.Param("id")
	job, err := h.getJobByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, job)
}

// CancelJob handles POST /api/compress/:id/cancel
func (h *CompressHandler) CancelJob(c *gin.Context) {
	id := c.Param("id")

	// Try to cancel via in-memory cancel func (for running jobs)
	h.mu.Lock()
	cancel, exists := h.jobs[id]
	h.mu.Unlock()

	if exists {
		cancel()
	}

	// Always update DB status — handles both running and orphaned jobs
	// (orphaned = job was running when container restarted, cancel func is gone)
	result, err := h.DB.Exec(`
		UPDATE compression_jobs
		SET status = 'cancelled', completed_at = ?
		WHERE id = ? AND status IN ('pending', 'running')
	`, time.Now(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		// Job already completed/failed/cancelled, or doesn't exist
		c.JSON(http.StatusNotFound, gin.H{"error": "job not found or already completed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "job cancelled"})

	// Invalidate cache for the instance associated with this job
	job, err := h.getJobByID(id)
	if err == nil {
		h.Cache.Invalidate("media:" + job.InstanceID)
		h.Cache.Invalidate("stats:" + job.InstanceID)
	}
}

// GetJobResults handles GET /api/compress/:id/results
func (h *CompressHandler) GetJobResults(c *gin.Context) {
	id := c.Param("id")

	// Verify job exists
	var count int
	err := h.DB.QueryRow(`SELECT COUNT(*) FROM compression_jobs WHERE id = ?`, id).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
		return
	}

	rows, err := h.DB.Query(`
		SELECT id, job_id, media_item_id, image_path, role, original_bytes, new_bytes, saved_bytes, status, skip_reason, error, new_width, new_height, new_format, created_at
		FROM compression_results WHERE job_id = ? ORDER BY created_at ASC
	`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var results []models.CompressionResult
	for rows.Next() {
		var r models.CompressionResult
		var createdAt string
		var skipReason, errStr, newFormat sql.NullString
		var newWidth, newHeight sql.NullInt64

		if err := rows.Scan(&r.ID, &r.JobID, &r.MediaItemID, &r.ImagePath, &r.Role,
			&r.OriginalBytes, &r.NewBytes, &r.SavedBytes, &r.Status, &skipReason, &errStr,
			&newWidth, &newHeight, &newFormat, &createdAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if skipReason.Valid {
			r.SkipReason = skipReason.String
		}
		if errStr.Valid {
			r.Error = errStr.String
		}
		if newWidth.Valid {
			r.NewWidth = int(newWidth.Int64)
		}
		if newHeight.Valid {
			r.NewHeight = int(newHeight.Int64)
		}
		if newFormat.Valid {
			r.NewFormat = newFormat.String
		}

		results = append(results, r)
	}

	if results == nil {
		results = []models.CompressionResult{}
	}

	c.JSON(http.StatusOK, results)
}

func (h *CompressHandler) getJobByID(id string) (models.CompressionJob, error) {
	var job models.CompressionJob
	var createdAt, startedAt, completedAt sql.NullString

	err := h.DB.QueryRow(`
		SELECT id, instance_id, status, config, total_items, processed_items, total_images, processed_images, saved_bytes, error_count, skip_count, created_at, started_at, completed_at
		FROM compression_jobs WHERE id = ?
	`, id).Scan(&job.ID, &job.InstanceID, &job.Status, &job.ConfigJSON,
		&job.TotalItems, &job.ProcessedItems, &job.TotalImages, &job.ProcessedImages,
		&job.SavedBytes, &job.ErrorCount, &job.SkipCount,
		&createdAt, &startedAt, &completedAt)
	if err != nil {
		return job, err
	}

	job.UnmarshalConfig()

	job.CreatedAt = util.ParseTimestamp(createdAt.String)
	job.StartedAt = util.ParseTimestampPtr(startedAt.String)
	job.CompletedAt = util.ParseTimestampPtr(completedAt.String)

	return job, nil
}

// mergeSettings merges per-instance settings into the job config as defaults.
// Request values always take precedence over instance settings.
// Instance settings take precedence over global defaults.
func mergeSettings(req models.JobConfig, instSettings models.InstanceSettings) models.JobConfig {
	out := models.JobConfig{
		Quality:  make(map[string]int),
		MaxWidth: make(map[string]int),
		MinSizeKB: make(map[string]int64),
	}

	// Determine all role keys from both request and instance settings
	allQualityKeys := make(map[string]bool)
	for k := range req.Quality {
		allQualityKeys[k] = true
	}
	for k := range instSettings.Quality {
		allQualityKeys[k] = true
	}
	// Ensure "default" is always present
	allQualityKeys["default"] = true

	// Better per-role quality defaults
	defaultQuality := map[string]int{
		"default":       82,
		"poster":        82,
		"fanart":        82,
		"season_poster": 82,
		"banner":        85,
		"clearLogo":     90, // PNG preserved, but if converted, use higher quality
	}

	for k := range allQualityKeys {
		switch {
		case req.Quality != nil && req.Quality[k] != 0:
			out.Quality[k] = req.Quality[k]
		case instSettings.Quality != nil && instSettings.Quality[k] != 0:
			out.Quality[k] = instSettings.Quality[k]
		case defaultQuality[k] != 0:
			out.Quality[k] = defaultQuality[k]
		default:
			out.Quality[k] = 82
		}
	}

	allWidthKeys := make(map[string]bool)
	for k := range req.MaxWidth {
		allWidthKeys[k] = true
	}
	for k := range instSettings.MaxWidth {
		allWidthKeys[k] = true
	}
	allWidthKeys["default"] = true

	// Better per-role max_width defaults
	defaultMaxWidth := map[string]int{
		"default":       1920,
		"poster":        1000,
		"season_poster": 1000,
	}

	for k := range allWidthKeys {
		switch {
		case req.MaxWidth != nil && req.MaxWidth[k] != 0:
			out.MaxWidth[k] = req.MaxWidth[k]
		case instSettings.MaxWidth != nil && instSettings.MaxWidth[k] != 0:
			out.MaxWidth[k] = instSettings.MaxWidth[k]
		case defaultMaxWidth[k] != 0:
			out.MaxWidth[k] = defaultMaxWidth[k]
		default:
			out.MaxWidth[k] = 1920
		}
	}

	// MinSizeKB: request > instance > default
	allMinSizeKeys := make(map[string]bool)
	for k := range req.MinSizeKB {
		allMinSizeKeys[k] = true
	}
	for k := range instSettings.MinSizeKB {
		allMinSizeKeys[k] = true
	}
	allMinSizeKeys["default"] = true

	defaultMinSizeKB := map[string]int64{
		"default":       30,
		"poster":        50,
		"fanart":        75,
		"season_poster": 50,
		"banner":        15,
		"clearLogo":     10,
	}

	for k := range allMinSizeKeys {
		switch {
		case req.MinSizeKB != nil && req.MinSizeKB[k] != 0:
			out.MinSizeKB[k] = req.MinSizeKB[k]
		case instSettings.MinSizeKB != nil && instSettings.MinSizeKB[k] != 0:
			out.MinSizeKB[k] = instSettings.MinSizeKB[k]
		case defaultMinSizeKB[k] != 0:
			out.MinSizeKB[k] = defaultMinSizeKB[k]
		default:
			out.MinSizeKB[k] = 30
		}
	}

	// Backup: request > instance > false
	switch {
	case req.Backup:
		out.Backup = true
	case instSettings.Backup != nil && *instSettings.Backup:
		out.Backup = true
	default:
		out.Backup = false
	}

	// LockPlex: request > instance > false
	switch {
	case req.LockPlex:
		out.LockPlex = true
	case instSettings.LockPlex != nil && *instSettings.LockPlex:
		out.LockPlex = true
	default:
		out.LockPlex = false
	}

	// MinSavingKB: request > instance > 50
	switch {
	case req.MinSavingKB > 0:
		out.MinSavingKB = req.MinSavingKB
	case instSettings.MinSavingKB != nil && *instSettings.MinSavingKB > 0:
		out.MinSavingKB = *instSettings.MinSavingKB
	default:
		out.MinSavingKB = 50
	}

	return out
}
