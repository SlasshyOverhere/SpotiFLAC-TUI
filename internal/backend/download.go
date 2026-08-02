package backend

import (
	"fmt"

	gobackend "github.com/zarz/spotiflac_android/go_backend"
)

type ProgressItem struct {
	ItemID        string  `json:"item_id"`
	BytesTotal    int64   `json:"bytes_total"`
	BytesReceived int64   `json:"bytes_received"`
	Progress      float64 `json:"progress"`
	SpeedMBps     float64 `json:"speed_mbps"`
	IsDownloading bool    `json:"is_downloading"`
	Status        string  `json:"status"`
	Stage         string  `json:"stage"`
}

func AllProgress() map[string]ProgressItem {
	var mp struct {
		Items map[string]ProgressItem `json:"items"`
	}
	_ = unmarshal(gobackend.GetAllDownloadProgress(), &mp)
	if mp.Items == nil {
		return map[string]ProgressItem{}
	}
	return mp.Items
}

func CancelDownload(itemID string) {
	gobackend.CancelDownload(itemID)
}

type AuthRequest struct {
	ExtensionID string `json:"extension_id"`
	AuthURL     string `json:"auth_url"`
	CallbackURL string `json:"callback_url"`
}

func PendingAuthRequests() ([]AuthRequest, error) {
	raw, err := gobackend.GetAllPendingAuthRequestsJSON()
	if err != nil {
		return nil, err
	}
	var reqs []AuthRequest
	if err := unmarshal(raw, &reqs); err != nil {
		return nil, fmt.Errorf("parse auth requests: %w", err)
	}
	return reqs, nil
}

func SetAuthCode(extensionID, code string) {
	gobackend.SetExtensionAuthCodeByID(extensionID, code)
}

func IsAuthenticated(extensionID string) bool {
	return gobackend.IsExtensionAuthenticatedByID(extensionID)
}
