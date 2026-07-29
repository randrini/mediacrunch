package api

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/mediacrunch/mediacrunch/internal/clients"
)

const plexProduct = "MediaCrunch"

// pendingPINs stores active PINs waiting for user authentication
var pendingPINs sync.Map

// PlexPINResponse is returned when creating a PIN
type PlexPINResponse struct {
	PINID    int64  `json:"pin_id"`
	Code     string `json:"code"`
	AuthURL  string `json:"auth_url"`
	ClientID string `json:"client_id"`
}

// PlexPINStatusResponse is returned when checking PIN status
type PlexPINStatusResponse struct {
	Token    string `json:"token,omitempty"`
	Username string `json:"username,omitempty"`
	Claimed  bool   `json:"claimed"`
}

// CreatePlexPIN handles POST /api/plex/pin
func (h *InstanceHandler) CreatePlexPIN(c *gin.Context) {
	clientID := uuid.NewString()
	authClient := clients.NewPlexAuthClient(clientID)

	pin, err := authClient.CreatePIN()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Store the client identifier so we can check the PIN later
	pendingPINs.Store(pin.ID, &plexPINEntry{
		ClientID: clientID,
		Code:     pin.Code,
		Expires:  time.Now().Add(15 * time.Minute),
	})

	authURL := "https://app.plex.tv/auth#?clientID=" + clientID +
		"&code=" + pin.Code +
		"&context%5Bdevice%5D%5Bproduct%5D=" + plexProduct

	c.JSON(http.StatusOK, PlexPINResponse{
		PINID:    pin.ID,
		Code:     pin.Code,
		AuthURL:  authURL,
		ClientID: clientID,
	})
}

// CheckPlexPIN handles GET /api/plex/pin/:id
func (h *InstanceHandler) CheckPlexPIN(c *gin.Context) {
	pinIDStr := c.Param("id")
	var pinID int64
	if _, err := fmt.Sscanf(pinIDStr, "%d", &pinID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pin ID"})
		return
	}

	entry, ok := pendingPINs.Load(pinID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "PIN not found or expired"})
		return
	}
	pe := entry.(*plexPINEntry)

	// Check with plex.tv
	authClient := clients.NewPlexAuthClient(pe.ClientID)
	pin, err := authClient.CheckPIN(pinID, pe.Code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if pin.AuthToken == nil || *pin.AuthToken == "" {
		c.JSON(http.StatusOK, PlexPINStatusResponse{
			Claimed: false,
		})
		return
	}

	// Token acquired — verify it and get username
	token := *pin.AuthToken
	username, _ := authClient.VerifyToken(token)

	// Clean up
	pendingPINs.Delete(pinID)

	c.JSON(http.StatusOK, PlexPINStatusResponse{
		Token:    token,
		Username: username,
		Claimed:  true,
	})
}

type plexPINEntry struct {
	ClientID string
	Code     string
	Expires  time.Time
}
