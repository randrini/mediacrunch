package scanner

import (
	"context"
	"crypto/sha1"
	"database/sql"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	_ "image/jpeg"
	_ "image/png"

	_ "github.com/mattn/go-sqlite3"

	"github.com/mediacrunch/mediacrunch/internal/models"
)

// plexTitleInfo holds title info from the Plex database.
type plexTitleInfo struct {
	Title      string
	Year       int
	MediaType  int // 1=movie, 2=show, 3=season, 4=episode, 9=collection
	GUID       string
	RatingKey  string
	Deleted    bool // true if the metadata item was deleted (orphaned bundle on disk)
}

// PlexScanner scans Plex Media Server instances.
type PlexScanner struct{}

// Scan implements the Scanner interface for Plex.
func (s *PlexScanner) Scan(ctx context.Context, instance models.Instance) ([]models.MediaItem, error) {
	// Resolve metadata root
	metadataRoot, err := s.resolveMetadataRoot(instance.PathPrefix)
	if err != nil {
		return nil, fmt.Errorf("resolve metadata root: %w", err)
	}
	fmt.Printf("SCAN: Plex metadata root: %s\n", metadataRoot)

	// Read titles from the Plex database for enrichment
	titleMap := s.readPlexDB(instance.PathPrefix)

	// Build a SHA1 -> ratingKey map from the Plex DB for fast matching
	// The Plex DB hash is SHA1(guid), and the bundle path is hash[0]/hash[1:].bundle
	hashToInfo := map[string]plexTitleInfo{}
	for rk, info := range titleMap {
		if info.GUID != "" && !info.Deleted {
			h := fmt.Sprintf("%x", sha1.Sum([]byte(info.GUID)))
			hashToInfo[h] = info
			// Also store the ratingKey-based lookup
			hashToInfo[rk] = info
		}
	}
	fmt.Printf("SCAN: Loaded %d items from Plex DB, %d GUID hashes\n", len(titleMap), len(hashToInfo))

	// Walk filesystem: discover all .bundle directories
	var items []models.MediaItem
	matchedBundles := 0

	for _, mediaDir := range []string{"Movies", "TV Shows", "Collections"} {
		select {
		case <-ctx.Done():
			return items, ctx.Err()
		default:
		}

		dirPath := filepath.Join(metadataRoot, mediaDir)
		entries, err := os.ReadDir(dirPath)
		if err != nil {
			fmt.Printf("SCAN: No %s directory: %v\n", mediaDir, err)
			continue
		}

		for _, shard := range entries {
			if !shard.IsDir() {
				continue
			}
			shardPath := filepath.Join(dirPath, shard.Name())
			bundleEntries, err := os.ReadDir(shardPath)
			if err != nil {
				continue
			}

			for _, bundle := range bundleEntries {
				if !bundle.IsDir() || !strings.HasSuffix(bundle.Name(), ".bundle") {
					continue
				}

				// Bundle name is hash[1:] (39 chars), shard dir is hash[0]
				// Reconstruct full 40-char hash
				bundleName := strings.TrimSuffix(bundle.Name(), ".bundle")
				fullHash := shard.Name() + bundleName

				bundlePath := filepath.Join(shardPath, bundle.Name(), "Contents", "_combined")

				if _, err := os.Stat(bundlePath); err != nil {
					continue
				}

				mediaType := s.mediaTypeFromDir(mediaDir)
				images := s.walkBundle(bundlePath)
				if len(images) == 0 {
					continue
				}

				matchedBundles++

				// Look up title from Plex DB using the full hash
				title := bundleName // fallback
				year := 0
				ratingKey := bundleHashToRatingKey(fullHash)

				if info, ok := hashToInfo[fullHash]; ok {
					title = info.Title
					year = info.Year
					ratingKey = info.RatingKey
				} else {
					// Bundle exists on disk but has no matching (non-deleted) item in the
					// Plex DB. This is an orphaned bundle left behind when a title was
					// removed from Plex. Skip it instead of showing the raw hash as a title.
					fmt.Printf("SCAN: skipping orphaned bundle %s (no active Plex item)\n", fullHash)
					continue
				}

				item := models.MediaItem{
					ID:         uuid.NewString(),
					InstanceID: instance.ID,
					MediaType:  mediaType,
					Title:      title,
					Year:       year,
					RemoteID:   ratingKey,
					Path:       bundlePath,
					Images:     images,
					Compressed: false,
				}

				s.computeTotals(&item)
				items = append(items, item)
			}
		}
	}

	fmt.Printf("SCAN: Matched %d bundles with images, %d total items\n", matchedBundles, len(items))
	return items, nil
}

// bundleHashToRatingKey converts a 40-char SHA1 hash to a short ratingKey for storage.
// Since the hash is unique, we can use a truncated version as the remote_id.
func bundleHashToRatingKey(fullHash string) string {
	if len(fullHash) >= 16 {
		return fullHash[:16]
	}
	return fullHash
}

// readPlexDB reads titles and metadata from the Plex SQLite database.
func (s *PlexScanner) readPlexDB(pathPrefix string) map[string]plexTitleInfo {
	dbPath := filepath.Join(pathPrefix, "config", "Plug-in Support", "Databases", "com.plexapp.plugins.library.db")
	titleMap := map[string]plexTitleInfo{}

	db, err := sql.Open("sqlite3", dbPath+"?mode=ro")
	if err != nil {
		fmt.Printf("SCAN: Cannot open Plex DB at %s: %v\n", dbPath, err)
		return titleMap
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, title, year, guid, metadata_type, deleted_at FROM metadata_items")
	if err != nil {
		fmt.Printf("SCAN: Cannot query Plex DB: %v\n", err)
		return titleMap
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var title string
		var year sql.NullInt64
		var guid string
		var metadataType int
		var deletedAt sql.NullInt64
		if err := rows.Scan(&id, &title, &year, &guid, &metadataType, &deletedAt); err != nil {
			continue
		}
		y := 0
		if year.Valid {
			y = int(year.Int64)
		}
		ratingKey := fmt.Sprintf("%d", id)
		titleMap[ratingKey] = plexTitleInfo{
			Title:      title,
			Year:       y,
			MediaType:  metadataType,
			GUID:       guid,
			RatingKey:  ratingKey,
			Deleted:    deletedAt.Valid,
		}
	}

	fmt.Printf("SCAN: Loaded %d titles from Plex DB\n", len(titleMap))
	return titleMap
}

// resolveMetadataRoot finds the Plex metadata directory.
func (s *PlexScanner) resolveMetadataRoot(pathPrefix string) (string, error) {
	candidates := []string{
		filepath.Join(pathPrefix, "config", "Metadata"),
		filepath.Join(pathPrefix, "Metadata"),
	}
	for _, c := range candidates {
		if info, err := os.Stat(c); err == nil && info.IsDir() {
			return c, nil
		}
	}
	return "", fmt.Errorf("no Metadata directory found under %s (tried config/Metadata and Metadata)", pathPrefix)
}

// mediaTypeFromDir infers the media type from the Plex directory name.
func (s *PlexScanner) mediaTypeFromDir(mediaDir string) string {
	switch mediaDir {
	case "Movies":
		return "movie"
	case "TV Shows":
		return "series"
	case "Collections":
		return "collection"
	default:
		return "unknown"
	}
}

// plexBundleHash computes the SHA1 hash of a GUID for Plex path construction.
func plexBundleHash(guid string) string {
	h := sha1.Sum([]byte(guid))
	return fmt.Sprintf("%x", h)
}

func (s *PlexScanner) walkBundle(bundlePath string) []models.ImageInfo {
	var images []models.ImageInfo

	subdirs := []string{"posters", "art", "clearLogos", "banners", "squareArt"}
	for _, subdir := range subdirs {
		dir := filepath.Join(bundlePath, subdir)
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			path := filepath.Join(dir, entry.Name())
			info, err := entry.Info()
			if err != nil {
				continue
			}

			// Skip theme audio files and very small files
			if info.Size() < 1024 {
				continue
			}

			role := s.roleFromSubdir(subdir)
			width, height, format := s.getImageDimensions(path)

			images = append(images, models.ImageInfo{
				Role:      role,
				Path:      path,
				SizeBytes: info.Size(),
				Width:     width,
				Height:    height,
				Format:    format,
			})
		}
	}

	return images
}

func (s *PlexScanner) roleFromSubdir(subdir string) string {
	switch subdir {
	case "posters":
		return "poster"
	case "art":
		return "art"
	case "clearLogos":
		return "clearLogo"
	case "banners":
		return "banner"
	case "squareArt":
		return "squareArt"
	default:
		return "unknown"
	}
}

func (s *PlexScanner) getImageDimensions(path string) (width, height int, format string) {
	f, err := os.Open(path)
	if err != nil {
		return 0, 0, ""
	}
	defer f.Close()

	cfg, imgFormat, err := image.DecodeConfig(f)
	if err != nil {
		return 0, 0, ""
	}

	return cfg.Width, cfg.Height, imgFormat
}

func (s *PlexScanner) computeTotals(item *models.MediaItem) {
	var totalSize int64
	for _, img := range item.Images {
		totalSize += img.SizeBytes
	}
	item.TotalSize = totalSize
	item.TotalImages = len(item.Images)
}