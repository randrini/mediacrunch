package clients

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const plexProduct = "MediaCrunch"
const plexPINURL = "https://plex.tv/api/v2/pins"

// PlexPIN represents a PIN from plex.tv
type PlexPIN struct {
	ID        int64   `json:"id"`
	Code      string  `json:"code"`
	AuthToken *string `json:"authToken"`
	ExpiresAt string  `json:"expiresAt"`
}

// PlexAuthClient handles plex.tv authentication
type PlexAuthClient struct {
	ClientIdentifier string
	HTTPClient       *http.Client
}

// NewPlexAuthClient creates a new auth client with a persistent client identifier
func NewPlexAuthClient(clientIdentifier string) *PlexAuthClient {
	return &PlexAuthClient{
		ClientIdentifier: clientIdentifier,
		HTTPClient:       &http.Client{Timeout: 15 * time.Second},
	}
}

// CreatePIN generates a new PIN for the OAuth flow
func (c *PlexAuthClient) CreatePIN() (*PlexPIN, error) {
	data := url.Values{}
	data.Set("strong", "true")
	data.Set("X-Plex-Product", plexProduct)
	data.Set("X-Plex-Client-Identifier", c.ClientIdentifier)

	req, err := http.NewRequest("POST", plexPINURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("plex PIN API returned status %d", resp.StatusCode)
	}

	var pin PlexPIN
	if err := json.NewDecoder(resp.Body).Decode(&pin); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &pin, nil
}

// CheckPIN checks if a PIN has been claimed and returns the auth token
func (c *PlexAuthClient) CheckPIN(pinID int64, pinCode string) (*PlexPIN, error) {
	u := fmt.Sprintf("%s/%d", plexPINURL, pinID)
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Plex-Client-Identifier", c.ClientIdentifier)

	q := req.URL.Query()
	q.Set("code", pinCode)
	q.Set("X-Plex-Product", plexProduct)
	q.Set("X-Plex-Client-Identifier", c.ClientIdentifier)
	req.URL.RawQuery = q.Encode()

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("plex PIN check returned status %d", resp.StatusCode)
	}

	var pin PlexPIN
	if err := json.NewDecoder(resp.Body).Decode(&pin); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &pin, nil
}

// VerifyToken checks if a Plex token is still valid
func (c *PlexAuthClient) VerifyToken(token string) (username string, err error) {
	req, err := http.NewRequest("GET", "https://plex.tv/api/v2/user", nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Plex-Token", token)
	req.Header.Set("X-Plex-Product", plexProduct)
	req.Header.Set("X-Plex-Client-Identifier", c.ClientIdentifier)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return "", fmt.Errorf("token is invalid or expired")
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("plex user API returned status %d", resp.StatusCode)
	}

	var user struct {
		Username string `json:"username"`
		Email    string `json:"email"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return "", fmt.Errorf("decode user: %w", err)
	}
	return user.Username, nil
}
