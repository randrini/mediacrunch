package scanner

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/mediacrunch/mediacrunch/internal/db"
	"github.com/mediacrunch/mediacrunch/internal/logger"
	"github.com/mediacrunch/mediacrunch/internal/models"
)

// Scanner defines the interface for scanning media items from an instance.
type Scanner interface {
	Scan(ctx context.Context, instance models.Instance) ([]models.MediaItem, error)
}

// Dispatcher routes scan requests to the appropriate scanner based on instance type.
type Dispatcher struct {
	DB     *db.DB
	Logger *logger.Logger
}

// NewDispatcher creates a new scanner dispatcher.
func NewDispatcher(database *db.DB, log *logger.Logger) *Dispatcher {
	return &Dispatcher{DB: database, Logger: log}
}

// existingItem holds the DB state for a media item matched by path.
type existingItem struct {
	ID           string
	Path         string
	TotalSize    int64
	OriginalSize int64
	Compressed   bool
	Locked       *bool
}

// Scan dispatches to the correct scanner implementation and upserts results.
func (d *Dispatcher) Scan(ctx context.Context, instance models.Instance) ([]models.MediaItem, error) {
	d.Logger.Infof("scanner", instance.ID, "Scan started for %s (%s)", instance.Name, instance.Type)

	var s Scanner
	switch instance.Type {
	case "radarr":
		s = &ArrScanner{InstanceType: "radarr"}
	case "sonarr":
		s = &ArrScanner{InstanceType: "sonarr"}
	case "plex":
		s = &PlexScanner{}
	default:
		return nil, fmt.Errorf("unknown instance type: %s", instance.Type)
	}

	items, err := s.Scan(ctx, instance)
	if err != nil {
		d.Logger.Errorf("scanner", instance.ID, "Scan failed: %v", err)
		return nil, fmt.Errorf("scan %s: %w", instance.Type, err)
	}

	// Load existing items for this instance into a map keyed by path
	existingItems := make(map[string]existingItem)
	rows, err := d.DB.Query(`SELECT id, path, total_size, original_size, compressed, locked FROM media_items WHERE instance_id = ?`, instance.ID)
	if err != nil {
		return nil, fmt.Errorf("query existing items: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var ei existingItem
		var locked sql.NullInt64
		if err := rows.Scan(&ei.ID, &ei.Path, &ei.TotalSize, &ei.OriginalSize, &ei.Compressed, &locked); err != nil {
			return nil, fmt.Errorf("scan existing item row: %w", err)
		}
		if locked.Valid {
			v := locked.Int64 != 0
			ei.Locked = &v
		}
		existingItems[ei.Path] = ei
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate existing items: %w", err)
	}

	// Build a set of scanned paths for stale-item detection
	scannedPaths := make(map[string]struct{}, len(items))

	var matched, newCount, resetCount int

	// Upsert each scanned item
	for i := range items {
		item := &items[i]
		scannedPaths[item.Path] = struct{}{}

		imagesJSON, err := item.MarshalImages()
		if err != nil {
			return nil, fmt.Errorf("marshal images for %s: %w", item.Title, err)
		}

		existing, wasExisting := existingItems[item.Path]
		if wasExisting {
			matched++
			// Preserve the existing ID so compression_results FK references stay valid
			item.ID = existing.ID

			// Preserve the locked field from the DB (scanner may not set it)
			if existing.Locked != nil {
				item.Locked = existing.Locked
			}

			if existing.Compressed {
				if item.TotalSize != existing.TotalSize {
					// Source replaced the compressed file — reset compression state
					item.Compressed = false
					item.OriginalSize = 0
					resetCount++
				} else {
					// Compressed file still in place — preserve compression state
					item.Compressed = true
					item.OriginalSize = existing.OriginalSize
				}
			} else {
				// Not previously compressed — use scanned values (already set)
				item.OriginalSize = 0
			}

			lockedVal := interface{}(nil)
			if item.Locked != nil {
				if *item.Locked {
					lockedVal = 1
				} else {
					lockedVal = 0
				}
			}

			_, err = d.DB.Exec(`
				UPDATE media_items SET
					media_type = ?, title = ?, year = ?, remote_id = ?,
					images = ?, total_size = ?, total_images = ?,
					compressed = ?, original_size = ?,
					locked = ?,
					scanned_at = CURRENT_TIMESTAMP
				WHERE id = ?
			`, item.MediaType, item.Title, item.Year, item.RemoteID,
				imagesJSON, item.TotalSize, item.TotalImages,
				boolToInt(item.Compressed), item.OriginalSize,
				lockedVal,
				item.ID)
			if err != nil {
				return nil, fmt.Errorf("update media item %s: %w", item.Title, err)
			}
		} else {
			newCount++
			// New item — use the UUID already assigned by the scanner
			item.Compressed = false
			item.OriginalSize = 0

			lockedVal := interface{}(nil)
			if item.Locked != nil {
				if *item.Locked {
					lockedVal = 1
				} else {
					lockedVal = 0
				}
			}

			_, err = d.DB.Exec(`
				INSERT INTO media_items (id, instance_id, media_type, title, year, remote_id, path, images, total_size, total_images, compressed, locked, scanned_at)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
			`, item.ID, item.InstanceID, item.MediaType, item.Year, item.RemoteID, item.Path, imagesJSON, item.TotalSize, item.TotalImages, boolToInt(item.Compressed), lockedVal)
			if err != nil {
				return nil, fmt.Errorf("insert media item %s: %w", item.Title, err)
			}
		}
	}

	// Delete stale items that exist in DB but were not in the scan results
	var removedCount int
	for path, ei := range existingItems {
		if _, found := scannedPaths[path]; !found {
			_, err := d.DB.Exec(`DELETE FROM compression_results WHERE media_item_id = ?`, ei.ID)
			if err != nil {
				return nil, fmt.Errorf("delete compression results for stale item %s: %w", ei.ID, err)
			}
			_, err = d.DB.Exec(`DELETE FROM media_items WHERE id = ?`, ei.ID)
			if err != nil {
				return nil, fmt.Errorf("delete stale media item %s: %w", ei.ID, err)
			}
			removedCount++
		}
	}

	// Compute totals for logging
	var itemCount, imageCount int
	var totalSize int64
	for _, item := range items {
		itemCount++
		imageCount += item.TotalImages
		totalSize += item.TotalSize
	}

	d.Logger.Infof("scanner", instance.ID, "Scan completed: %d items, %d images, %d bytes total", itemCount, imageCount, totalSize)
	d.Logger.Infof("scanner", instance.ID, "Scan upsert: %d items matched, %d new, %d removed, %d compression resets (source replaced files)", matched, newCount, removedCount, resetCount)

	return items, nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
