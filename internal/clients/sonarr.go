package clients

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// SonarrSeries represents a series from the Sonarr API.
type SonarrSeries struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Year     int    `json:"year"`
	Path     string `json:"path"`
	Images   []struct {
		CoverType string `json:"coverType"`
		URL       string `json:"url"`
	} `json:"images"`
	Seasons  []struct {
		SeasonNumber int `json:"seasonNumber"`
	} `json:"seasons"`
}

// SonarrClient communicates with a Sonarr instance.
type SonarrClient struct {
	Host   string
	APIKey string
	Client *http.Client
}

// NewSonarrClient creates a new SonarrClient.
func NewSonarrClient(host, apiKey string) *SonarrClient {
	return &SonarrClient{
		Host:   host,
		APIKey: apiKey,
		Client: &http.Client{},
	}
}

// TestConnection tests connectivity to a Sonarr instance and returns version info.
func (c *SonarrClient) TestConnection() (version, appName string, err error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v3/system/status", c.Host), nil)
	if err != nil {
		return "", "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("X-Api-Key", c.APIKey)

	resp, err := c.Client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var status struct {
		Version string `json:"version"`
		AppName string `json:"appName"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return "", "", fmt.Errorf("decode response: %w", err)
	}
	return status.Version, status.AppName, nil
}

// GetSeries fetches all series from Sonarr.
func (c *SonarrClient) GetSeries() ([]SonarrSeries, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v3/series", c.Host), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("X-Api-Key", c.APIKey)

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get series: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sonarr API returned status %d", resp.StatusCode)
	}

	var series []SonarrSeries
	if err := json.NewDecoder(resp.Body).Decode(&series); err != nil {
		return nil, fmt.Errorf("decode series: %w", err)
	}
	return series, nil
}
