package checker

import (
	"bytes"
	"encoding/json"
	"net/http"
)

// NotifyMsg represents the message used to notify the notify url endpoint of a new update
type NotifyMsg struct {
	AccessToken  string  `json:"access_token"`
	DeviceType   string  `json:"device_type"`
	StatusUpdate bool    `json:"status_update"`
	Message      string  `json:"message"`
	APIInfo      APIInfo `json:"api_info"`
}

// APIInfo contains api information to be used by NotifyMsg
type APIInfo struct {
	Key        string `json:"key"`
	OldVersion string `json:"old_version"`
	NewVersion string `json:"new_version"`
}

// NotifyUpdate performs a POST request to the notify url with the update information
func NotifyUpdate(url, service, oldVersion, newVersion, accessToken string) error {
	notifyMsg := NotifyMsg{APIInfo: APIInfo{Key: service, OldVersion: oldVersion, NewVersion: newVersion}}
	marshaledObject, err := json.MarshalIndent(notifyMsg, "", "  ")
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(marshaledObject))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	c := &http.Client{}
	_, err = c.Do(req)
	if err != nil {
		return err
	}

	return nil
}
