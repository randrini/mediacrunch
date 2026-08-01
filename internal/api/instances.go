package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/mediacrunch/mediacrunch/internal/cache"
	"github.com/mediacrunch/mediacrunch/internal/clients"
	"github.com/mediacrunch/mediacrunch/internal/db"
	"github.com/mediacrunch/mediacrunch/internal/models"
	"github.com/mediacrunch/mediacrunch/internal/scanner"
	"github.com/mediacrunch/mediacrunch/internal/util"
)

// InstanceHandler handles instance-related API endpoints.
type InstanceHandler struct {
	DB       *db.DB
	Scanner  *scanner.Dispatcher
	Cache    *cache.Cache
}

// NewInstanceHandler creates a new InstanceHandler.
func NewInstanceHandler(database *db.DB, scannerDispatcher *scanner.Dispatcher, c *cache.Cache) *InstanceHandler {
	return &InstanceHandler{
		DB:      database,
		Scanner: scannerDispatcher,
		Cache:   c,
	}
}

// CreateInstance handles POST /api/instances
func (h *InstanceHandler) CreateInstance(c *gin.Context) {
	var input struct {
		Type       string `json:"type" binding:"required"`
		Name       string `json:"name" binding:"required"`
		Host       string `json:"host" binding:"required"`
		APIKey     string `json:"api_key" binding:"required"`
		PathPrefix string `json:"path_prefix" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	validTypes := map[string]bool{"radarr": true, "sonarr": true, "plex": true}
	if !validTypes[input.Type] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "type must be radarr, sonarr, or plex"})
		return
	}

	instance := models.Instance{
		ID:         uuid.NewString(),
		Type:       input.Type,
		Name:       input.Name,
		Host:       strings.TrimRight(input.Host, "/"),
		APIKey:     input.APIKey,
		PathPrefix: input.PathPrefix,
		CreatedAt:  time.Now(),
	}

	_, err := h.DB.Exec(`
		INSERT INTO instances (id, type, name, host, api_key, path_prefix, settings, created_at)
		VALUES (?, ?, ?, ?, ?, ?, '{}', ?)
	`, instance.ID, instance.Type, instance.Name, instance.Host, instance.APIKey, instance.PathPrefix, instance.CreatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("create instance: %v", err)})
		return
	}

	c.JSON(http.StatusCreated, instance)
}

// ListInstances handles GET /api/instances
func (h *InstanceHandler) ListInstances(c *gin.Context) {
	rows, err := h.DB.Query(`SELECT id, type, name, host, api_key, path_prefix, settings, created_at FROM instances ORDER BY created_at DESC`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var instances []models.Instance
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
		instances = append(instances, inst)
	}

	if instances == nil {
		instances = []models.Instance{}
	}

	c.JSON(http.StatusOK, instances)
}

// GetInstance handles GET /api/instances/:id
func (h *InstanceHandler) GetInstance(c *gin.Context) {
	id := c.Param("id")
	inst, err := h.getInstanceByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "instance not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, inst)
}

// UpdateInstance handles PUT /api/instances/:id
func (h *InstanceHandler) UpdateInstance(c *gin.Context) {
	id := c.Param("id")

	// Verify instance exists
	_, err := h.getInstanceByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "instance not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var input struct {
		Name       string `json:"name"`
		Host       string `json:"host"`
		APIKey     string `json:"api_key"`
		PathPrefix string `json:"path_prefix"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Build dynamic update query
	var setClauses []string
	var args []interface{}

	if input.Name != "" {
		setClauses = append(setClauses, "name = ?")
		args = append(args, input.Name)
	}
	if input.Host != "" {
		setClauses = append(setClauses, "host = ?")
		args = append(args, strings.TrimRight(input.Host, "/"))
	}
	if input.APIKey != "" {
		setClauses = append(setClauses, "api_key = ?")
		args = append(args, input.APIKey)
	}
	if input.PathPrefix != "" {
		setClauses = append(setClauses, "path_prefix = ?")
		args = append(args, input.PathPrefix)
	}

	if len(setClauses) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
		return
	}

	args = append(args, id)
	query := fmt.Sprintf("UPDATE instances SET %s WHERE id = ?", strings.Join(setClauses, ", "))
	_, err = h.DB.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("update instance: %v", err)})
		return
	}

	inst, err := h.getInstanceByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to re-fetch instance"})
		return
	}
	c.JSON(http.StatusOK, inst)
}

// DeleteInstance handles DELETE /api/instances/:id
func (h *InstanceHandler) DeleteInstance(c *gin.Context) {
	id := c.Param("id")

	// Delete compression results for media items of this instance
	_, err := h.DB.Exec(`
		DELETE FROM compression_results WHERE media_item_id IN (SELECT id FROM media_items WHERE instance_id = ?)
	`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete compression results"})
		return
	}

	// Delete media items
	_, err = h.DB.Exec(`DELETE FROM media_items WHERE instance_id = ?`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete media items"})
		return
	}

	// Delete compression jobs
	_, err = h.DB.Exec(`DELETE FROM compression_jobs WHERE instance_id = ?`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete compression jobs"})
		return
	}

	// Delete instance
	result, err := h.DB.Exec(`DELETE FROM instances WHERE id = ?`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("delete instance: %v", err)})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "instance not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "instance deleted"})

	// Invalidate cache
	h.Cache.Invalidate("media:" + id)
	h.Cache.Invalidate("stats:" + id)
}

// ScanInstance handles POST /api/instances/:id/scan
func (h *InstanceHandler) ScanInstance(c *gin.Context) {
	id := c.Param("id")
	inst, err := h.getInstanceByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "instance not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Run scan in background with timeout context
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()
		_, err := h.Scanner.Scan(ctx, inst)
		if err != nil {
			fmt.Printf("ERROR scanning instance %s: %v\n", inst.Name, err)
		}
		// Invalidate cache after scan completes
		h.Cache.Invalidate("media:" + id)
		h.Cache.Invalidate("stats:" + id)
	}()

	c.JSON(http.StatusAccepted, gin.H{"message": "scan started", "instance_id": id})
}

// TestConnectionRequest is the input for the test connection endpoint.
type TestConnectionRequest struct {
	Type   string `json:"type" binding:"required"`
	Host   string `json:"host" binding:"required"`
	APIKey string `json:"api_key" binding:"required"`
}

// TestConnectionResponse is the output for the test connection endpoint.
type TestConnectionResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Details *struct {
		Version string `json:"version"`
		Name    string `json:"name"`
	} `json:"details,omitempty"`
}

// TestConnection handles POST /api/instances/test
func (h *InstanceHandler) TestConnection(c *gin.Context) {
	var input TestConnectionRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	host := strings.TrimRight(input.Host, "/")

	var version, name string
	var connErr error

	switch input.Type {
	case "radarr":
		client := clients.NewRadarrClient(host, input.APIKey)
		version, name, connErr = client.TestConnection()
	case "sonarr":
		client := clients.NewSonarrClient(host, input.APIKey)
		version, name, connErr = client.TestConnection()
	case "plex":
		client := clients.NewPlexClient(host, input.APIKey)
		version, name, connErr = client.TestConnection()
	default:
		c.JSON(http.StatusBadRequest, TestConnectionResponse{
			Success: false,
			Message: "type must be radarr, sonarr, or plex",
		})
		return
	}

	if connErr != nil {
		c.JSON(http.StatusOK, TestConnectionResponse{
			Success: false,
			Message: connErr.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, TestConnectionResponse{
		Success: true,
		Message: "Connection successful",
		Details: &struct {
			Version string `json:"version"`
			Name    string `json:"name"`
		}{
			Version: version,
			Name:    name,
		},
	})
}

// LockInstance handles POST /api/instances/:id/lock (plex only)
func (h *InstanceHandler) LockInstance(c *gin.Context) {
	id := c.Param("id")
	inst, err := h.getInstanceByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "instance not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if inst.Type != "plex" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "lock is only supported for plex instances"})
		return
	}

	// Get all media items for this instance and lock them
	rows, err := h.DB.Query(`SELECT remote_id FROM media_items WHERE instance_id = ? AND remote_id IS NOT NULL`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	plexClient := clients.NewPlexClient(inst.Host, inst.APIKey)
	var locked int
	for rows.Next() {
		var ratingKey string
		if err := rows.Scan(&ratingKey); err != nil {
			fmt.Printf("ERROR scanning rating key: %v\n", err)
			continue
		}
		if err := plexClient.LockMetadata(ratingKey); err != nil {
			fmt.Printf("ERROR locking metadata %s: %v\n", ratingKey, err)
		} else {
			locked++
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("locked %d items", locked), "locked_count": locked})
}

// GetSettings handles GET /api/instances/:id/settings
func (h *InstanceHandler) GetSettings(c *gin.Context) {
	id := c.Param("id")
	inst, err := h.getInstanceByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "instance not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Ensure non-nil maps in response for empty settings
	if inst.Settings.Quality == nil {
		inst.Settings.Quality = map[string]int{}
	}
	if inst.Settings.MaxWidth == nil {
		inst.Settings.MaxWidth = map[string]int{}
	}
	if inst.Settings.MinSizeKB == nil {
		inst.Settings.MinSizeKB = map[string]int64{}
	}

	c.JSON(http.StatusOK, inst.Settings)
}

// UpdateSettings handles PUT /api/instances/:id/settings
func (h *InstanceHandler) UpdateSettings(c *gin.Context) {
	id := c.Param("id")

	// Verify instance exists
	inst, err := h.getInstanceByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "instance not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var input models.InstanceSettings
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate quality values
	for role, q := range input.Quality {
		if q < 1 || q > 100 {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("quality for %q must be between 1 and 100", role)})
			return
		}
	}

	// Validate max_width values
	for role, w := range input.MaxWidth {
		if w < 100 || w > 8000 {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("max_width for %q must be between 100 and 8000", role)})
			return
		}
	}

	// Validate min_size_kb values
	for role, ms := range input.MinSizeKB {
		if ms < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("min_size_kb for %q must be >= 0", role)})
			return
		}
	}

	// Validate min_saving_kb
	if input.MinSavingKB != nil && *input.MinSavingKB < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "min_saving_kb must be >= 0"})
		return
	}

	// For Plex instances, prevent unlocking via settings (lock_plex: false when it was true)
	if inst.Type == "plex" && input.LockPlex != nil && !*input.LockPlex {
		// Allow setting lock_plex to false — the user may want to override
	}

	// Marshal and store
	settingsJSON, err := json.Marshal(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("marshal settings: %v", err)})
		return
	}

	_, err = h.DB.Exec(`UPDATE instances SET settings = ? WHERE id = ?`, string(settingsJSON), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("update settings: %v", err)})
		return
	}

	// Invalidate cache
	h.Cache.Invalidate("settings:" + id)

	// Ensure non-nil maps in response
	if input.Quality == nil {
		input.Quality = map[string]int{}
	}
	if input.MaxWidth == nil {
		input.MaxWidth = map[string]int{}
	}
	if input.MinSizeKB == nil {
		input.MinSizeKB = map[string]int64{}
	}

	c.JSON(http.StatusOK, input)
}

func (h *InstanceHandler) getInstanceByID(id string) (models.Instance, error) {
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

// Helper to convert string to int with default
func parseInt(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return v
}
