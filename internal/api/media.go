package api

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/mediacrunch/mediacrunch/internal/cache"
	"github.com/mediacrunch/mediacrunch/internal/db"
	"github.com/mediacrunch/mediacrunch/internal/models"
	"github.com/mediacrunch/mediacrunch/internal/util"
)

// MediaHandler handles media-related API endpoints.
type MediaHandler struct {
	DB    *db.DB
	Cache *cache.Cache
}

// NewMediaHandler creates a new MediaHandler.
func NewMediaHandler(database *db.DB, c *cache.Cache) *MediaHandler {
	return &MediaHandler{DB: database, Cache: c}
}

// ListMedia handles GET /api/instances/:id/media
func (h *MediaHandler) ListMedia(c *gin.Context) {
	instanceID := c.Param("id")

	// Parse query parameters
	mediaType := c.Query("type")
	search := c.Query("search")
	sort := c.Query("sort")
	order := c.Query("order")
	page := parseInt(c.Query("page"), 1)
	perPage := parseInt(c.Query("per_page"), 50)
	compressedFilter := c.Query("compressed")
	lockedFilter := c.Query("locked")
	includeImages := c.Query("include_images") == "1"

	// Build cache key
	cacheKey := fmt.Sprintf("media:%s:%s:%s:%s:%s:%s:%s:%d:%d:%t",
		instanceID, mediaType, search, sort, order, compressedFilter, lockedFilter, page, perPage, includeImages)

	// Check cache
	if cached, ok := h.Cache.Get(cacheKey); ok {
		c.JSON(http.StatusOK, cached)
		return
	}

	// Verify instance exists
	var count int
	err := h.DB.QueryRow(`SELECT COUNT(*) FROM instances WHERE id = ?`, instanceID).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "instance not found"})
		return
	}

	// Validate sort field and map to SQL expression
	validSorts := map[string]bool{"total_size": true, "original_size": true, "title": true, "year": true, "total_images": true}
	if !validSorts[sort] {
		sort = "total_size"
	}
	// original_size is 0 for uncompressed items, so fall back to total_size for sorting
	sortExpr := sort
	if sort == "original_size" {
		sortExpr = "COALESCE(NULLIF(original_size, 0), total_size)"
	}

	// Validate order
	if order != "asc" && order != "desc" {
		order = "desc"
	}

	// Build query
	where := []string{"instance_id = ?"}
	args := []interface{}{instanceID}

	if mediaType != "" {
		where = append(where, "media_type = ?")
		args = append(args, mediaType)
	}

	if search != "" {
		where = append(where, "title LIKE ?")
		args = append(args, "%"+search+"%")
	}

	if compressedFilter == "0" {
		where = append(where, "compressed = 0")
	} else if compressedFilter == "1" {
		where = append(where, "compressed = 1")
	}

	if lockedFilter == "0" {
		where = append(where, "(locked IS NULL OR locked = 0)")
	} else if lockedFilter == "1" {
		where = append(where, "locked = 1")
	}

	whereClause := strings.Join(where, " AND ")

	// Count total
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM media_items WHERE %s", whereClause)
	err = h.DB.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Pagination
	offset := (page - 1) * perPage

	// Select columns — images column is parsed server-side for per-role sizes,
	// but the full image array is only included in the response when requested
	selectCols := "id, instance_id, media_type, title, year, remote_id, path, images, total_size, original_size, total_images, compressed, locked, scanned_at"

	query := fmt.Sprintf(`
		SELECT %s FROM media_items WHERE %s ORDER BY %s %s LIMIT ? OFFSET ?
	`, selectCols, whereClause, sortExpr, order)

	args = append(args, perPage, offset)
	rows, err := h.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var items []models.MediaItem
	for rows.Next() {
		var item models.MediaItem
		var year sql.NullInt64
		var locked sql.NullInt64
		var scannedAt string
		var compressedInt int

		if err := rows.Scan(&item.ID, &item.InstanceID, &item.MediaType, &item.Title,
			&year, &item.RemoteID, &item.Path, &item.ImagesJSON,
			&item.TotalSize, &item.OriginalSize, &item.TotalImages, &compressedInt, &locked, &scannedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if year.Valid {
			item.Year = int(year.Int64)
		}
		item.Compressed = compressedInt == 1
		if locked.Valid {
			b := locked.Int64 == 1
			item.Locked = &b
		}

		// Parse scanned_at
		item.ScannedAt = util.ParseTimestampPtr(scannedAt)

		// Parse images for per-role sizes. The full array is only included in
		// the response when requested; per-role sizes are always computed.
		if err := item.UnmarshalImages(); err != nil {
			item.Images = []models.ImageInfo{}
		}
		item.ComputeRoleSizes()
		if !includeImages {
			item.Images = []models.ImageInfo{}
		}

		items = append(items, item)
	}

	if items == nil {
		items = []models.MediaItem{}
	}

	response := gin.H{
		"items":       items,
		"total":       total,
		"page":        page,
		"per_page":    perPage,
		"total_pages": (total + perPage - 1) / perPage,
	}

	// Cache the response
	h.Cache.Set(cacheKey, response)

	c.JSON(http.StatusOK, response)
}
