package scanner

import (
	"context"
	"fmt"
	"image"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/uuid"
	_ "image/jpeg"
	_ "image/png"

	"github.com/mediacrunch/mediacrunch/internal/clients"
	"github.com/mediacrunch/mediacrunch/internal/models"
)

// ArrScanner scans Radarr or Sonarr instances.
type ArrScanner struct {
	InstanceType string // "radarr" or "sonarr"
}

// Scan implements the Scanner interface for Radarr/Sonarr.
func (s *ArrScanner) Scan(ctx context.Context, instance models.Instance) ([]models.MediaItem, error) {
	var items []models.MediaItem

	switch s.InstanceType {
	case "radarr":
		client := clients.NewRadarrClient(instance.Host, instance.APIKey)
		movies, err := client.GetMovies()
		if err != nil {
			return nil, fmt.Errorf("get movies: %w", err)
		}
		fmt.Printf("INFO: Radarr scan found %d movies from %s\n", len(movies), instance.Host)

		for _, m := range movies {
			select {
			case <-ctx.Done():
				return items, ctx.Err()
			default:
			}

			item, err := s.scanMovie(instance, m)
			if err != nil {
				fmt.Printf("WARN: scan movie %s: %v\n", m.Title, err)
				continue
			}
			if len(item.Images) > 0 {
				items = append(items, item)
			}
		}
		fmt.Printf("INFO: Radarr scan: %d items with images out of %d movies\n", len(items), len(movies))

	case "sonarr":
		client := clients.NewSonarrClient(instance.Host, instance.APIKey)
		seriesList, err := client.GetSeries()
		if err != nil {
			return nil, fmt.Errorf("get series: %w", err)
		}
		fmt.Printf("INFO: Sonarr scan found %d series from %s\n", len(seriesList), instance.Host)

		for _, se := range seriesList {
			select {
			case <-ctx.Done():
				return items, ctx.Err()
			default:
			}

			item, err := s.scanSeries(instance, se)
			if err != nil {
				fmt.Printf("WARN: scan series %s: %v\n", se.Title, err)
				continue
			}
			if len(item.Images) > 0 {
				items = append(items, item)
			}

			// Scan season posters
			seasonItems, err := s.scanSeasons(instance, se)
			if err != nil {
				fmt.Printf("WARN: scan seasons for %s: %v\n", se.Title, err)
			}
			items = append(items, seasonItems...)
		}
		fmt.Printf("INFO: Sonarr scan: %d items with images out of %d series\n", len(items), len(seriesList))
	}

	return items, nil
}

func (s *ArrScanner) scanMovie(instance models.Instance, m clients.RadarrMovie) (models.MediaItem, error) {
	mediaCoverDir := filepath.Join(instance.PathPrefix, "MediaCover", strconv.Itoa(m.ID))
	images := s.walkMediaCoverDir(mediaCoverDir, "movie")

	item := models.MediaItem{
		ID:         uuid.NewString(),
		InstanceID: instance.ID,
		MediaType:  "movie",
		Title:      m.Title,
		Year:       m.Year,
		RemoteID:   strconv.Itoa(m.ID),
		Path:       mediaCoverDir,
		Images:     images,
		Compressed: false,
	}

	s.computeTotals(&item)
	return item, nil
}

func (s *ArrScanner) scanSeries(instance models.Instance, se clients.SonarrSeries) (models.MediaItem, error) {
	mediaCoverDir := filepath.Join(instance.PathPrefix, "MediaCover", strconv.Itoa(se.ID))
	images := s.walkMediaCoverDir(mediaCoverDir, "series")

	item := models.MediaItem{
		ID:         uuid.NewString(),
		InstanceID: instance.ID,
		MediaType:  "series",
		Title:      se.Title,
		Year:       se.Year,
		RemoteID:   strconv.Itoa(se.ID),
		Path:       mediaCoverDir,
		Images:     images,
		Compressed: false,
	}

	s.computeTotals(&item)
	return item, nil
}

func (s *ArrScanner) scanSeasons(instance models.Instance, se clients.SonarrSeries) ([]models.MediaItem, error) {
	var items []models.MediaItem

	for _, season := range se.Seasons {
		seasonDir := filepath.Join(instance.PathPrefix, "MediaCover", strconv.Itoa(se.ID), "seasons")
		images := s.walkSeasonDir(seasonDir, season.SeasonNumber)

		if len(images) == 0 {
			continue
		}

		item := models.MediaItem{
			ID:         uuid.NewString(),
			InstanceID: instance.ID,
			MediaType:  "season",
			Title:      fmt.Sprintf("%s - Season %d", se.Title, season.SeasonNumber),
			Year:       se.Year,
			RemoteID:   fmt.Sprintf("%d-s%d", se.ID, season.SeasonNumber),
			Path:       seasonDir,
			Images:     images,
			Compressed: false,
		}

		s.computeTotals(&item)
		items = append(items, item)
	}

	return items, nil
}

func (s *ArrScanner) walkMediaCoverDir(dir string, mediaType string) []models.ImageInfo {
	var images []models.ImageInfo

	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Printf("[scanner] warning: failed to read media cover directory %s: %v", dir, err)
		return images
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

		role := s.roleFromFilename(entry.Name(), mediaType)
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

	return images
}

func (s *ArrScanner) walkSeasonDir(dir string, seasonNumber int) []models.ImageInfo {
	var images []models.ImageInfo

	pattern := fmt.Sprintf("season%02d*", seasonNumber)
	entries, err := filepath.Glob(filepath.Join(dir, pattern))
	if err != nil {
		return images
	}

	for _, path := range entries {
		info, err := os.Stat(path)
		if err != nil {
			continue
		}

		width, height, format := s.getImageDimensions(path)

		images = append(images, models.ImageInfo{
			Role:      "season_poster",
			Path:      path,
			SizeBytes: info.Size(),
			Width:     width,
			Height:    height,
			Format:    format,
		})
	}

	return images
}

func (s *ArrScanner) roleFromFilename(name, mediaType string) string {
	lower := strings.ToLower(name)
	switch {
	case strings.Contains(lower, "poster"):
		return "poster"
	case strings.Contains(lower, "fanart"):
		return "fanart"
	case strings.Contains(lower, "banner"):
		return "banner"
	case strings.Contains(lower, "clearlogo"):
		return "clearLogo"
	default:
		return "unknown"
	}
}

func (s *ArrScanner) getImageDimensions(path string) (width, height int, format string) {
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

func (s *ArrScanner) computeTotals(item *models.MediaItem) {
	var totalSize int64
	for _, img := range item.Images {
		totalSize += img.SizeBytes
	}
	item.TotalSize = totalSize
	item.TotalImages = len(item.Images)
}
