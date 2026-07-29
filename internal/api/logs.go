package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/mediacrunch/mediacrunch/internal/db"
	"github.com/mediacrunch/mediacrunch/internal/models"
	"github.com/mediacrunch/mediacrunch/internal/util"
)

// LogHandler handles log-related API endpoints.
type LogHandler struct {
	DB *db.DB
}

// NewLogHandler creates a new LogHandler.
func NewLogHandler(database *db.DB) *LogHandler {
	return &LogHandler{DB: database}
}

// GetLogs handles GET /api/logs
// Query params: level (filter by level), source (filter by source), limit (default 100), offset (default 0), search (search in message)
func (h *LogHandler) GetLogs(c *gin.Context) {
	limit := 100
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 1000 {
			limit = v
		}
	}
	offset := 0
	if o := c.Query("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}

	query := `SELECT id, level, source, COALESCE(instance_id, ''), message, COALESCE(details, ''), created_at FROM logs`
	args := []interface{}{}
	conditions := []string{}

	if level := c.Query("level"); level != "" && level != "all" {
		conditions = append(conditions, "level = ?")
		args = append(args, level)
	}
	if source := c.Query("source"); source != "" {
		conditions = append(conditions, "source = ?")
		args = append(args, source)
	}
	if search := c.Query("search"); search != "" {
		conditions = append(conditions, "message LIKE ?")
		args = append(args, "%"+search+"%")
	}
	if instID := c.Query("instance_id"); instID != "" {
		conditions = append(conditions, "instance_id = ?")
		args = append(args, instID)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := h.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	logs := make([]models.LogEntry, 0)
	for rows.Next() {
		var log models.LogEntry
		var createdAt string
		if err := rows.Scan(&log.ID, &log.Level, &log.Source, &log.InstanceID, &log.Message, &log.Details, &createdAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		log.CreatedAt = util.ParseTimestamp(createdAt)
		logs = append(logs, log)
	}

	// Get total count for pagination
	countQuery := "SELECT COUNT(*) FROM logs"
	if len(conditions) > 0 {
		countQuery += " WHERE " + strings.Join(conditions, " AND ")
	}
	var total int
	h.DB.QueryRow(countQuery, args[:len(args)-2]...).Scan(&total)

	c.JSON(http.StatusOK, gin.H{
		"logs":   logs,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// ClearLogs handles DELETE /api/logs
func (h *LogHandler) ClearLogs(c *gin.Context) {
	result, err := h.DB.Exec("DELETE FROM logs")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	rowsAffected, _ := result.RowsAffected()
	c.JSON(http.StatusOK, gin.H{"message": "logs cleared", "deleted": rowsAffected})
}
