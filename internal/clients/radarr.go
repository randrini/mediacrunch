package clients

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// RadarrMovie represents a movie from the Radarr API.
type RadarrMovie struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Year     int    `json:"year"`
	Path     string `json:"path"`
	FolderPath string `json:"folderPath"`
	Images   []struct {
		CoverType string `json:"coverType"`
		URL       string `json:"url"`
	} `json:"images"`
}

// RadarrClient communicates with a Radarr instance.
type RadarrClient struct {
	Host    string
	APIKey  string
	Client  *http.Client
}

// NewRadarrClient creates a new RadarrClient.
func NewRadarrClient(host, apiKey string) *RadarrClient {
	return &RadarrClient{
		Host:   host,
		APIKey: apiKey,
		Client: &http.Client{},
	}
}

// TestConnection tests connectivity to a Radarr instance and returns version info.
func (c *RadarrClient) TestConnection() (version, appName string, err error) {
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

// GetMovies fetches all movies from Radarr.
func (c *RadarrClient) GetMovies() ([]RadarrMovie, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v3/movie", c.Host), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("X-Api-Key", c.APIKey)

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get movies: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("radarr API returned status %d", resp.StatusCode)
	}

	var movies []RadarrMovie
	if err := json.NewDecoder(resp.Body).Decode(&movies); err != nil {
		return nil, fmt.Errorf("decode movies: %w", err)
	}
	return movies, nil
}
