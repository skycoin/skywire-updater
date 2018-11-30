package checker

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type NotifyMsg struct {
	AccessToken string `json:"access_token"`
	DeviceType string `json:"device_type"`
	StatusUpdate bool `json:"status_update"`
	Message string `json:"message"`
	ApiInfo ApiInfo `json:"api_info"`
}

type ApiInfo struct {
		Key string `json:"key"`
		OldVersion string `json:"old_version"`
		NewVersion string `json:"new_version"`
}

func NotifyUpdate(url, service, oldVersion, newVersion, accessToken string) error {
	notifyMsg := NotifyMsg{ApiInfo: ApiInfo{Key:service, OldVersion: oldVersion, NewVersion: newVersion}}
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
