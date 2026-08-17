package models

import (
	"encoding/json"
	"fmt"
	"time"
)

// InstanceSettings holds per-instance compression defaults.
// Pointer types distinguish "not set" (nil) from "set to zero value".
type InstanceSettings struct {
	Quality     map[string]int   `json:"quality,omitempty"`
	MaxWidth    map[string]int   `json:"max_width,omitempty"`
	MinSizeKB   map[string]int64 `json:"min_size_kb,omitempty"`
	Backup      *bool            `json:"backup,omitempty"`
	MinSavingKB *int64           `json:"min_saving_kb,omitempty"`
	LockPlex    *bool            `json:"lock_plex,omitempty"`
}

// Instance represents a connected Radarr, Sonarr, or Plex instance.
type Instance struct {
	ID            string           `json:"id"`
	Type          string           `json:"type"` // radarr, sonarr, plex
	Name          string           `json:"name"`
	Host          string           `json:"host"`
	APIKey        string           `json:"api_key"`
	PathPrefix    string           `json:"path_prefix"`
	SettingsJSON  string           `json:"-"` // raw JSON from DB
	Settings      InstanceSettings `json:"settings"`
	CreatedAt     time.Time        `json:"created_at"`
}

// UnmarshalSettings parses the settings JSON column into the Settings field.
func (inst *Instance) UnmarshalSettings() error {
	if inst.SettingsJSON == "" || inst.SettingsJSON == "{}" {
		inst.Settings = InstanceSettings{}
		return nil
	}
	return json.Unmarshal([]byte(inst.SettingsJSON), &inst.Settings)
}

// MarshalSettings serializes the Settings field to JSON for storage.
func (inst *Instance) MarshalSettings() (string, error) {
	b, err := json.Marshal(inst.Settings)
	if err != nil {
		return "{}", err
	}
	return string(b), nil
}

// ImageInfo describes a single image file associated with a media item.
type ImageInfo struct {
	Role      string `json:"role"`      // poster, art, clearLogo, banner, season_poster, episode_thumb, fanart
	Path      string `json:"path"`      // absolute filesystem path
	SizeBytes int64  `json:"size_bytes"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Format    string `json:"format"`    // jpeg, png
}

// MediaItem represents a movie, series, season, episode, or collection.
type MediaItem struct {
	ID              string      `json:"id"`
	InstanceID      string      `json:"instance_id"`
	MediaType       string      `json:"media_type"` // movie, series, season, episode, collection
	Title           string      `json:"title"`
	Year            int         `json:"year"`
	RemoteID        string      `json:"remote_id"`
	Path            string      `json:"path"`
	ImagesJSON      string      `json:"-"` // raw JSON from DB
	Images          []ImageInfo `json:"images"`
	TotalSize       int64       `json:"total_size"`
	OriginalSize    int64       `json:"original_size"` // pre-compression size (0 if never compressed)
	TotalImages     int         `json:"total_images"`
	Compressed      bool        `json:"compressed"`
	Locked          *bool       `json:"locked,omitempty"` // plex only
	ScannedAt       *time.Time  `json:"scanned_at,omitempty"`
	PosterSize      int64       `json:"poster_size"`
	FanartSize      int64       `json:"fanart_size"`
	ClearLogoSize   int64       `json:"clear_logo_size"`
	SeasonPosterSize int64      `json:"season_poster_size"`
	BannerSize      int64       `json:"banner_size"`
}

// ComputeRoleSizes sums the size of every image per role into the per-role
// size fields used by the media list UI. Must be called after UnmarshalImages.
func (item *MediaItem) ComputeRoleSizes() {
	for _, img := range item.Images {
		switch img.Role {
		case "poster":
			item.PosterSize += img.SizeBytes
		case "fanart", "art":
			item.FanartSize += img.SizeBytes
		case "clearLogo":
			item.ClearLogoSize += img.SizeBytes
		case "season_poster":
			item.SeasonPosterSize += img.SizeBytes
		case "banner":
			item.BannerSize += img.SizeBytes
		}
	}
}

// UnmarshalImages parses the images JSON column into the Images slice.
func (m *MediaItem) UnmarshalImages() error {
	if m.ImagesJSON == "" || m.ImagesJSON == "[]" {
		m.Images = []ImageInfo{}
		return nil
	}
	return json.Unmarshal([]byte(m.ImagesJSON), &m.Images)
}

// MarshalImages serializes the Images slice to JSON for storage.
func (m *MediaItem) MarshalImages() (string, error) {
	b, err := json.Marshal(m.Images)
	if err != nil {
		return "[]", err
	}
	return string(b), nil
}

// LogEntry represents a single log event.
type LogEntry struct {
	ID         int64     `json:"id"`
	Level      string    `json:"level"`       // debug, info, warn, error
	Source     string    `json:"source"`      // scanner, compressor, api, system
	InstanceID string    `json:"instance_id,omitempty"` // optional, which instance
	Message    string    `json:"message"`
	Details    string    `json:"details,omitempty"` // optional JSON string with extra context
	CreatedAt  time.Time `json:"created_at"`
}

// FormatSize returns a human-readable size string.
func FormatSize(bytes int64) string {
	switch {
	case bytes >= 1<<30:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(1<<30))
	case bytes >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(1<<20))
	case bytes >= 1<<10:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(1<<10))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// CompressionJob represents a batch compression operation.
type CompressionJob struct {
	ID             string     `json:"id"`
	InstanceID     string     `json:"instance_id"`
	MediaItemIDs   []string   `json:"media_item_ids,omitempty"` // empty = all items for instance
	Status         string     `json:"status"` // pending, running, completed, failed, cancelled
	ConfigJSON     string     `json:"-"`      // raw JSON from DB
	Config         JobConfig  `json:"config"`
	TotalItems     int        `json:"total_items"`
	ProcessedItems int        `json:"processed_items"`
	TotalImages    int64      `json:"total_images"`
	ProcessedImages int64     `json:"processed_images"`
	SavedBytes     int64      `json:"saved_bytes"`
	ErrorCount     int        `json:"error_count"`
	SkipCount      int        `json:"skip_count"`
	CreatedAt      time.Time  `json:"created_at"`
	StartedAt      *time.Time `json:"started_at,omitempty"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
}

// JobConfig holds the per-job compression settings.
type JobConfig struct {
	Quality     map[string]int   `json:"quality,omitempty"`
	MaxWidth    map[string]int   `json:"max_width,omitempty"`
	MinSizeKB   map[string]int64 `json:"min_size_kb,omitempty"`
	Backup      bool             `json:"backup"`
	MinSavingKB int64            `json:"min_saving_kb"`
	LockPlex    bool             `json:"lock_plex"`
	Recompress  bool             `json:"recompress"`
}

// UnmarshalConfig parses the config JSON column.
func (j *CompressionJob) UnmarshalConfig() error {
	if j.ConfigJSON == "" || j.ConfigJSON == "{}" {
		j.Config = JobConfig{}
		return nil
	}
	return json.Unmarshal([]byte(j.ConfigJSON), &j.Config)
}

// MarshalConfig serializes the config to JSON for storage.
func (j *CompressionJob) MarshalConfig() (string, error) {
	b, err := json.Marshal(j.Config)
	if err != nil {
		return "{}", err
	}
	return string(b), nil
}

// CompressionResult records the outcome of compressing a single image.
type CompressionResult struct {
	ID            string    `json:"id"`
	JobID         string    `json:"job_id"`
	MediaItemID   string    `json:"media_item_id"`
	ImagePath     string    `json:"image_path"`
	Role          string    `json:"role"`
	OriginalBytes int64     `json:"original_bytes"`
	NewBytes      int64     `json:"new_bytes"`
	SavedBytes    int64     `json:"saved_bytes"`
	Status        string    `json:"status"` // compressed, skipped, error
	SkipReason    string    `json:"skip_reason,omitempty"`
	Error         string    `json:"error,omitempty"`
	NewWidth      int       `json:"new_width,omitempty"`
	NewHeight     int       `json:"new_height,omitempty"`
	NewFormat     string    `json:"new_format,omitempty"` // "jpeg" or "png"
	CreatedAt     time.Time `json:"created_at"`
}

// SavedPercent returns the percentage saved.
func (r *CompressionResult) SavedPercent() float64 {
	if r.OriginalBytes == 0 {
		return 0
	}
	return float64(r.SavedBytes) / float64(r.OriginalBytes) * 100
}

// FormatOriginalSize returns a human-readable original size.
func (r *CompressionResult) FormatOriginalSize() string {
	return FormatSize(r.OriginalBytes)
}

// FormatNewSize returns a human-readable new size.
func (r *CompressionResult) FormatNewSize() string {
	return FormatSize(r.NewBytes)
}

// FormatSaved returns a human-readable saved size.
func (r *CompressionResult) FormatSaved() string {
	return FormatSize(r.SavedBytes)
}
