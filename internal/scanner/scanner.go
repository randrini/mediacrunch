package scanner

import (
	"context"
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

// Scan dispatches to the correct scanner implementation.
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

	// Delete old items for this instance before inserting fresh scan results
	_, err = d.DB.Exec(`DELETE FROM compression_results WHERE media_item_id IN (SELECT id FROM media_items WHERE instance_id = ?)`, instance.ID)
	if err != nil {
		return nil, fmt.Errorf("delete old compression results: %w", err)
	}
	_, err = d.DB.Exec(`DELETE FROM media_items WHERE instance_id = ?`, instance.ID)
	if err != nil {
		return nil, fmt.Errorf("delete old media items: %w", err)
	}

	fmt.Printf("INFO: Cleared old items for instance %s, inserting %d new items\n", instance.Name, len(items))

	// Persist scanned items to database
	for _, item := range items {
		imagesJSON, err := item.MarshalImages()
		if err != nil {
			return nil, fmt.Errorf("marshal images for %s: %w", item.Title, err)
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
			INSERT INTO media_items (id, instance_id, media_type, title, year, remote_id, path, images, total_size, total_images, compressed, locked, scanned_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		`, item.ID, item.InstanceID, item.MediaType, item.Title, item.Year, item.RemoteID, item.Path, imagesJSON, item.TotalSize, item.TotalImages, boolToInt(item.Compressed), lockedVal)
		if err != nil {
			return nil, fmt.Errorf("insert media item %s: %w", item.Title, err)
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

	return items, nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
