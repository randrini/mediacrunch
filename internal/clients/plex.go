package clients

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// PlexLibrary represents a library section from Plex.
type PlexLibrary struct {
	Key     string `json:"key"`
	Title   string `json:"title"`
	Type    string `json:"type"` // movie, show
	Agent   string `json:"agent"`
	Scanner string `json:"scanner"`
}

// PlexMediaContainer wraps the Plex API response.
type PlexMediaContainer struct {
	MediaContainer struct {
		Directory    []PlexLibrary `json:"Directory"`
		Metadata     []PlexItem    `json:"Metadata"`
		Size         int           `json:"size"`
		TotalSize    int           `json:"totalSize"`
		Version      string        `json:"version"`
		FriendlyName string        `json:"friendlyName"`
	} `json:"MediaContainer"`
}

// PlexGUID represents an alternative GUID for a Plex item.
type PlexGUID struct {
	ID string `json:"id"`
}

// PlexItem represents a media item from Plex.
type PlexItem struct {
	RatingKey        string     `json:"ratingKey"`
	Key              string     `json:"key"`
	Title            string     `json:"title"`
	Type             string     `json:"type"` // movie, show, season, episode, collection
	Year             int        `json:"year"`
	GUID             string     `json:"guid"`
	Guids            []PlexGUID `json:"Guid"` // alternative GUIDs (capital G in Plex API)
	LibrarySectionID string    `json:"librarySectionID"`
}

// PlexIdentityContainer wraps the root identity response.
type PlexIdentityContainer struct {
	MediaContainer struct {
		Version      string `json:"version"`
		FriendlyName string `json:"friendlyName"`
	} `json:"MediaContainer"`
}

// PlexClient communicates with a Plex Media Server instance.
type PlexClient struct {
	Host   string
	Token  string
	Client *http.Client
}

// NewPlexClient creates a new PlexClient.
func NewPlexClient(host, token string) *PlexClient {
	return &PlexClient{
		Host:   host,
		Token:  token,
		Client: &http.Client{},
	}
}

func (c *PlexClient) setHeaders(req *http.Request) {
	req.Header.Set("X-Plex-Token", c.Token)
	req.Header.Set("X-Plex-Product", "MediaCrunch")
	req.Header.Set("X-Plex-Version", "1.0")
	req.Header.Set("X-Plex-Client-Identifier", "mediacrunch")
	req.Header.Set("Accept", "application/json")
}

// TestConnection tests connectivity to a Plex instance and returns version info.
func (c *PlexClient) TestConnection() (version, appName string, err error) {
	// Use root / endpoint which always returns version
	u, _ := url.JoinPath(c.Host, "/")
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return "", "", fmt.Errorf("create request: %w", err)
	}
	c.setHeaders(req)

	resp, err := c.Client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var container PlexIdentityContainer
	if err := json.NewDecoder(resp.Body).Decode(&container); err != nil {
		return "", "", fmt.Errorf("decode response: %w", err)
	}

	ver := container.MediaContainer.Version
	if ver == "" {
		ver = "unknown"
	}
	name := container.MediaContainer.FriendlyName
	if name == "" {
		name = "Plex Media Server"
	}
	return ver, name, nil
}

// GetLibraries fetches all library sections from Plex.
func (c *PlexClient) GetLibraries() ([]PlexLibrary, error) {
	u, _ := url.JoinPath(c.Host, "/library/sections")
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	c.setHeaders(req)

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get libraries: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("plex API returned status %d", resp.StatusCode)
	}

	var container PlexMediaContainer
	if err := json.NewDecoder(resp.Body).Decode(&container); err != nil {
		return nil, fmt.Errorf("decode libraries: %w", err)
	}
	return container.MediaContainer.Directory, nil
}

// GetLibraryItems fetches all items in a library section with pagination.
func (c *PlexClient) GetLibraryItems(sectionKey string) ([]PlexItem, error) {
	var allItems []PlexItem
	containerStart := 0
	containerSize := 100

	for {
		u, _ := url.JoinPath(c.Host, "/library/sections", sectionKey, "all")
		req, err := http.NewRequest("GET", u, nil)
		if err != nil {
			return nil, fmt.Errorf("create request: %w", err)
		}
		c.setHeaders(req)

		q := req.URL.Query()
		q.Set("X-Plex-Container-Start", strconv.Itoa(containerStart))
		q.Set("X-Plex-Container-Size", strconv.Itoa(containerSize))
		req.URL.RawQuery = q.Encode()

		resp, err := c.Client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("get library items: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("plex API returned status %d for library section %s", resp.StatusCode, sectionKey)
		}

		var container PlexMediaContainer
		if err := json.NewDecoder(resp.Body).Decode(&container); err != nil {
			return nil, fmt.Errorf("decode library items: %w", err)
		}

		items := container.MediaContainer.Metadata
		if len(items) == 0 {
			break
		}
		allItems = append(allItems, items...)

		if len(items) < containerSize {
			break // no more pages
		}
		containerStart += containerSize
		resp.Body.Close()
	}

	return allItems, nil
}

// LockMetadata locks thumb and art fields for a Plex item.
func (c *PlexClient) LockMetadata(ratingKey string) error {
	u, _ := url.JoinPath(c.Host, "/library/metadata", ratingKey)
	req, err := http.NewRequest("PUT", u, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	c.setHeaders(req)

	q := req.URL.Query()
	q.Set("thumb.locked", "1")
	q.Set("art.locked", "1")
	req.URL.RawQuery = q.Encode()

	resp, err := c.Client.Do(req)
	if err != nil {
		return fmt.Errorf("lock metadata: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("plex lock returned status %d", resp.StatusCode)
	}
	return nil
}

// UnlockMetadata unlocks thumb and art fields for a Plex item.
func (c *PlexClient) UnlockMetadata(ratingKey string) error {
	u, _ := url.JoinPath(c.Host, "/library/metadata", ratingKey)
	req, err := http.NewRequest("PUT", u, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	c.setHeaders(req)

	q := req.URL.Query()
	q.Set("thumb.locked", "0")
	q.Set("art.locked", "0")
	req.URL.RawQuery = q.Encode()

	resp, err := c.Client.Do(req)
	if err != nil {
		return fmt.Errorf("unlock metadata: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("plex unlock returned status %d", resp.StatusCode)
	}
	return nil
}