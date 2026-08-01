package compressor

import (
	"context"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"sync/atomic"
	"time"

	"github.com/disintegration/imaging"
	"github.com/google/uuid"

	"github.com/mediacrunch/mediacrunch/internal/clients"
	"github.com/mediacrunch/mediacrunch/internal/db"
	"github.com/mediacrunch/mediacrunch/internal/logger"
	"github.com/mediacrunch/mediacrunch/internal/models"
	"github.com/mediacrunch/mediacrunch/internal/util"
)

// Compressor handles image compression jobs.
type Compressor struct {
	DB     *db.DB
	Logger *logger.Logger
}

// NewCompressor creates a new Compressor.
func NewCompressor(database *db.DB, log *logger.Logger) *Compressor {
	return &Compressor{DB: database, Logger: log}
}

// RunCompressionJob processes a compression job asynchronously.
func (c *Compressor) RunCompressionJob(ctx context.Context, job *models.CompressionJob) {
	now := time.Now()
	job.StartedAt = &now
	job.Status = "running"
	c.updateJob(job)

	c.Logger.Infof("compressor", job.InstanceID, "Compression started for %d items", job.TotalItems)

	// Get the instance
	instance, err := c.getInstance(job.InstanceID)
	if err != nil {
		job.Status = "failed"
		now2 := time.Now()
		job.CompletedAt = &now2
		c.updateJob(job)
		c.Logger.Errorf("compressor", job.InstanceID, "Compression failed: %v", err)
		return
	}

	// If lock_plex is enabled and instance is plex, lock metadata first
	if job.Config.LockPlex && instance.Type == "plex" {
		c.lockPlexItems(ctx, instance, job)
	}

	// Get media items for this job
	items, err := c.getMediaItems(job)
	if err != nil {
		job.Status = "failed"
		now2 := time.Now()
		job.CompletedAt = &now2
		c.updateJob(job)
		return
	}

	// Count total images across all items for progress tracking
	for _, item := range items {
		job.TotalImages += int64(item.TotalImages)
	}
	c.updateJob(job)

	for _, item := range items {
		select {
		case <-ctx.Done():
			job.Status = "cancelled"
			now2 := time.Now()
			job.CompletedAt = &now2
			c.updateJob(job)
			return
		default:
		}

		err := c.compressItem(ctx, instance, job, item)
		if err != nil {
			job.ErrorCount++
			c.Logger.Errorf("compressor", job.InstanceID, "Error compressing \"%s\": %v", item.Title, err)
		}
		job.ProcessedItems++
		c.updateJob(job)
	}

	job.Status = "completed"
	now2 := time.Now()
	job.CompletedAt = &now2
	c.updateJob(job)
	c.Logger.Infof("compressor", job.InstanceID, "Compression completed: %d items, %d images, saved %s", job.ProcessedItems, job.ProcessedImages, models.FormatSize(job.SavedBytes))
}

func (c *Compressor) compressItem(ctx context.Context, instance models.Instance, job *models.CompressionJob, item models.MediaItem) error {
	if err := item.UnmarshalImages(); err != nil {
		return fmt.Errorf("unmarshal images: %w", err)
	}

	var totalNewSize int64
	var imagesUpdated bool

	for i, img := range item.Images {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		result := c.compressSingleImage(job, item, img, i)
		atomic.AddInt64(&job.ProcessedImages, 1)
		// Update job progress every 10 images or on every image if few remain
		if job.ProcessedImages%10 == 0 || job.ProcessedImages == job.TotalImages {
			c.updateJob(job)
		}
		if result.Status == "compressed" {
			item.Images[i].SizeBytes = result.NewBytes
			item.Images[i].Width = result.NewWidth
			item.Images[i].Height = result.NewHeight
			item.Images[i].Format = result.NewFormat
			totalNewSize += result.NewBytes
			imagesUpdated = true
			atomic.AddInt64(&job.SavedBytes, result.SavedBytes)
		} else if result.Status == "skipped" {
			totalNewSize += img.SizeBytes
			job.SkipCount++
		} else {
			totalNewSize += img.SizeBytes
		}

		c.saveResult(result)
	}

	if imagesUpdated {
		imagesJSON, err := item.MarshalImages()
		if err != nil {
			return fmt.Errorf("marshal updated images: %w", err)
		}

		_, err = c.DB.Exec(`
			UPDATE media_items SET images = ?, total_size = ?, original_size = ?, compressed = 1 WHERE id = ?
		`, imagesJSON, totalNewSize, item.TotalSize, item.ID)
		if err != nil {
			return fmt.Errorf("update media item: %w", err)
		}
	}

	return nil
}

// shouldPreserveAsPNG returns true for image roles that need transparency
// and should be kept as PNG rather than converted to JPEG.
func shouldPreserveAsPNG(role string) bool {
	switch role {
	case "clearLogo", "banner":
		return true
	default:
		return false
	}
}

func (c *Compressor) compressSingleImage(job *models.CompressionJob, item models.MediaItem, img models.ImageInfo, idx int) models.CompressionResult {
	result := models.CompressionResult{
		ID:            uuid.NewString(),
		JobID:         job.ID,
		MediaItemID:   item.ID,
		ImagePath:     img.Path,
		Role:          img.Role,
		OriginalBytes: img.SizeBytes,
		Status:        "error",
		CreatedAt:     time.Now(),
	}

	// Check minimum size threshold for this role
	// If the original file is already very small, skip compression to avoid quality degradation
	minSizeBytes := job.Config.MinSizeKB["default"] * 1024
	if ms, ok := job.Config.MinSizeKB[img.Role]; ok {
		minSizeBytes = ms * 1024
	}
	if minSizeBytes > 0 && img.SizeBytes < minSizeBytes {
		result.Status = "skipped"
		result.SkipReason = fmt.Sprintf("original size %s below minimum %s threshold for %s", models.FormatSize(img.SizeBytes), models.FormatSize(minSizeBytes), img.Role)
		result.NewBytes = img.SizeBytes
		result.SavedBytes = 0
		return result
	}

	// Open the source image
	srcImg, err := imaging.Open(img.Path)
	if err != nil {
		result.Error = fmt.Sprintf("open image: %v", err)
		return result
	}

	// Determine max width for this role
	maxWidth := job.Config.MaxWidth["default"]
	if w, ok := job.Config.MaxWidth[img.Role]; ok {
		maxWidth = w
	}
	if maxWidth <= 0 {
		maxWidth = 1920
	}

	// Determine quality for this role (JPEG only)
	quality := job.Config.Quality["default"]
	if q, ok := job.Config.Quality[img.Role]; ok {
		quality = q
	}
	if quality < 1 {
		quality = 80
	}
	if quality > 100 {
		quality = 100
	}

	// Determine output format based on source format and role
	preservePNG := img.Format == "png" || shouldPreserveAsPNG(img.Role)

	// Resize if wider than max
	bounds := srcImg.Bounds()
	width := bounds.Dx()
	dstImg := srcImg
	if width > maxWidth {
		if preservePNG {
			// For PNG: resize preserving alpha channel
			dstImg = imaging.Resize(srcImg, maxWidth, 0, imaging.Lanczos)
		} else {
			// For JPEG: composite onto white background, then resize
			dstImg = imaging.Resize(c.prepareImage(srcImg, img.Format), maxWidth, 0, imaging.Lanczos)
		}
	} else if !preservePNG {
		// No resize needed, but still need to composite PNG onto white for JPEG conversion
		dstImg = c.prepareImage(srcImg, img.Format)
	}

	// Backup original if requested
	if job.Config.Backup {
		backupPath := img.Path + ".bak." + job.ID
		if _, err := os.Stat(backupPath); os.IsNotExist(err) {
			if err := copyFile(img.Path, backupPath); err != nil {
				result.Error = fmt.Sprintf("backup: %v", err)
				return result
			}
		}
	}

	// Encode to destination format
	tmpPath := img.Path + ".tmp"
	f, err := os.Create(tmpPath)
	if err != nil {
		result.Error = fmt.Sprintf("create temp file: %v", err)
		return result
	}

	if preservePNG {
		// Encode as PNG (lossless, preserves transparency)
		err = png.Encode(f, dstImg)
	} else {
		// Encode as JPEG (lossy compression)
		err = jpeg.Encode(f, dstImg, &jpeg.Options{Quality: quality})
	}
	if err := f.Close(); err != nil {
		os.Remove(tmpPath)
		result.Error = fmt.Sprintf("close temp file: %v", err)
		return result
	}
	if err != nil {
		os.Remove(tmpPath)
		result.Error = fmt.Sprintf("encode image: %v", err)
		return result
	}

	// Check new size
	newInfo, err := os.Stat(tmpPath)
	if err != nil {
		os.Remove(tmpPath)
		result.Error = fmt.Sprintf("stat temp file: %v", err)
		return result
	}
	newBytes := newInfo.Size()
	if newBytes == 0 {
		os.Remove(tmpPath)
		result.Error = "encoded file is empty"
		return result
	}
	savedBytes := img.SizeBytes - newBytes

	// Check minimum saving
	minSaving := job.Config.MinSavingKB * 1024
	if minSaving <= 0 {
		minSaving = 50 * 1024
	}

	// For PNG: since PNG re-encoding is lossless, only replace if the file is actually smaller
	// For PNG that wasn't resized, the re-encoded file will be nearly the same size — skip it
	if preservePNG && width <= maxWidth {
		// No resize happened and format is PNG — nothing to gain
		os.Remove(tmpPath)
		result.Status = "skipped"
		result.SkipReason = "PNG not resized, preserving original"
		result.NewBytes = img.SizeBytes
		result.SavedBytes = 0
		return result
	}

	if savedBytes < 0 {
		os.Remove(tmpPath)
		result.Status = "skipped"
		result.SkipReason = "re-encoded file is larger than original"
		result.NewBytes = img.SizeBytes
		result.SavedBytes = 0
		return result
	}

	if savedBytes < minSaving {
		os.Remove(tmpPath)
		result.Status = "skipped"
		result.SkipReason = fmt.Sprintf("saving %d bytes < minimum %d bytes", savedBytes, minSaving)
		result.NewBytes = img.SizeBytes
		result.SavedBytes = 0
		return result
	}

	// Replace original with compressed
	if err := os.Rename(tmpPath, img.Path); err != nil {
		os.Remove(tmpPath)
		result.Error = fmt.Sprintf("replace file: %v", err)
		return result
	}

	result.Status = "compressed"
	result.NewBytes = newBytes
	result.SavedBytes = savedBytes
	result.NewWidth = dstImg.Bounds().Dx()
	result.NewHeight = dstImg.Bounds().Dy()
	if preservePNG {
		result.NewFormat = "png"
	} else {
		result.NewFormat = "jpeg"
	}
	c.Logger.Debugf("compressor", job.InstanceID, "%s: %s → %s (%s)", item.Title, models.FormatSize(img.SizeBytes), models.FormatSize(result.NewBytes), result.Role)
	return result
}

// prepareImage composites PNG transparency onto a white background for JPEG conversion.
func (c *Compressor) prepareImage(img image.Image, format string) image.Image {
	switch format {
	case "png":
		// Composite PNG with transparency onto a white background
		// so it can be saved as JPEG without black transparency artifacts
		bounds := img.Bounds()
		rgba := image.NewRGBA(bounds)
		draw.Draw(rgba, bounds, image.White, image.Point{}, draw.Src)
		draw.Draw(rgba, bounds, img, bounds.Min, draw.Over)
		return rgba
	default:
		return img
	}
}

func (c *Compressor) updateJob(job *models.CompressionJob) {
	configJSON, _ := job.MarshalConfig()
	_, err := c.DB.Exec(`
		UPDATE compression_jobs SET
			status = ?, config = ?, total_items = ?, processed_items = ?,
			total_images = ?, processed_images = ?,
			saved_bytes = ?, error_count = ?, skip_count = ?,
			started_at = ?, completed_at = ?
		WHERE id = ?
	`, job.Status, configJSON, job.TotalItems, job.ProcessedItems,
		job.TotalImages, job.ProcessedImages,
		job.SavedBytes, job.ErrorCount, job.SkipCount,
		job.StartedAt, job.CompletedAt, job.ID)
	if err != nil {
		fmt.Printf("ERROR updating job %s: %v\n", job.ID, err)
	}
}

func (c *Compressor) saveResult(result models.CompressionResult) {
	_, err := c.DB.Exec(`
		INSERT INTO compression_results (id, job_id, media_item_id, image_path, role, original_bytes, new_bytes, saved_bytes, status, skip_reason, error, new_width, new_height, new_format, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, result.ID, result.JobID, result.MediaItemID, result.ImagePath, result.Role,
		result.OriginalBytes, result.NewBytes, result.SavedBytes, result.Status,
		result.SkipReason, result.Error, result.NewWidth, result.NewHeight, result.NewFormat, result.CreatedAt)
	if err != nil {
		fmt.Printf("ERROR saving result: %v\n", err)
	}
}

func (c *Compressor) getInstance(instanceID string) (models.Instance, error) {
	var inst models.Instance
	var createdAt string
	err := c.DB.QueryRow(`
		SELECT id, type, name, host, api_key, path_prefix, settings, created_at FROM instances WHERE id = ?
	`, instanceID).Scan(&inst.ID, &inst.Type, &inst.Name, &inst.Host, &inst.APIKey, &inst.PathPrefix, &inst.SettingsJSON, &createdAt)
	if err != nil {
		return inst, fmt.Errorf("get instance: %w", err)
	}
	inst.CreatedAt = util.ParseTimestamp(createdAt)
	inst.UnmarshalSettings()
	return inst, nil
}

func (c *Compressor) getMediaItems(job *models.CompressionJob) ([]models.MediaItem, error) {
	query := `
		SELECT id, instance_id, media_type, title, year, remote_id, path, images, total_size, original_size, total_images, compressed, locked, scanned_at
		FROM media_items WHERE instance_id = ?`
	args := []interface{}{job.InstanceID}

	if len(job.MediaItemIDs) > 0 {
		placeholders := ""
		for i, id := range job.MediaItemIDs {
			if i > 0 {
				placeholders += ","
			}
			placeholders += "?"
			args = append(args, id)
		}
		query += ` AND id IN (` + placeholders + `)`
	}

	rows, err := c.DB.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query media items: %w", err)
	}
	defer rows.Close()

	var items []models.MediaItem
	for rows.Next() {
		var item models.MediaItem
		var year sqlNullInt64
		var locked sqlNullInt64
		var scannedAt string
		var compressedInt int

		err := rows.Scan(&item.ID, &item.InstanceID, &item.MediaType, &item.Title,
			&year, &item.RemoteID, &item.Path, &item.ImagesJSON,
			&item.TotalSize, &item.OriginalSize, &item.TotalImages, &compressedInt, &locked, &scannedAt)
		if err != nil {
			return nil, fmt.Errorf("scan media item: %w", err)
		}

		item.Year = int(year.Int64)
		item.Compressed = compressedInt == 1
		if locked.Valid {
			b := locked.Int64 == 1
			item.Locked = &b
		}
		item.ScannedAt = util.ParseTimestampPtr(scannedAt)

		items = append(items, item)
	}

	return items, nil
}

func (c *Compressor) lockPlexItems(ctx context.Context, instance models.Instance, job *models.CompressionJob) {
	client := clients.NewPlexClient(instance.Host, instance.APIKey)

	// Only lock items being compressed in this job
	query := `SELECT remote_id FROM media_items WHERE instance_id = ? AND compressed = 0`
	args := []interface{}{job.InstanceID}

	if len(job.MediaItemIDs) > 0 {
		placeholders := ""
		for i, id := range job.MediaItemIDs {
			if i > 0 {
				placeholders += ","
			}
			placeholders += "?"
			args = append(args, id)
		}
		query += ` AND id IN (` + placeholders + `)`
	}

	rows, err := c.DB.Query(query, args...)
	if err != nil {
		fmt.Printf("ERROR querying items for plex lock: %v\n", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		select {
		case <-ctx.Done():
			return
		default:
		}

		var ratingKey string
		if err := rows.Scan(&ratingKey); err != nil {
			continue
		}

		if err := client.LockMetadata(ratingKey); err != nil {
			fmt.Printf("WARN: lock plex metadata %s: %v\n", ratingKey, err)
		}
	}
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// sqlNullInt64 is a helper for scanning nullable int64 fields.
type sqlNullInt64 struct {
	Int64 int64
	Valid bool
}

func (n *sqlNullInt64) Scan(value interface{}) error {
	if value == nil {
		n.Valid = false
		return nil
	}
	n.Valid = true
	switch v := value.(type) {
	case int64:
		n.Int64 = v
	case float64:
		n.Int64 = int64(v)
	case int:
		n.Int64 = int64(v)
	case []byte:
		s := string(v)
		fmt.Sscanf(s, "%d", &n.Int64)
	default:
		n.Valid = false
	}
	return nil
}