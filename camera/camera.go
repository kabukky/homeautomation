package camera

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"
)

var (
	httpClient = http.Client{
		Timeout: 10 * time.Second,
	}
)

type CameraStatus struct {
	Framesize int `json:"framesize"`
	Quality   int `json:"quality"`
}

func GetImage(esp32CamAddress string) ([]byte, error) {
	// Check settings
	settings, err := getStatus(esp32CamAddress)
	if err != nil {
		return nil, err
	}
	if settings.Framesize != 13 {
		err = setStatus("framesize", "13", esp32CamAddress)
		if err != nil {
			return nil, err
		}
		time.Sleep(700 * time.Millisecond)
	}
	if settings.Quality != 4 {
		err = setStatus("quality", "4", esp32CamAddress)
		if err != nil {
			return nil, err
		}
		time.Sleep(700 * time.Millisecond)
	}

	resp, err := httpClient.Get(esp32CamAddress + "/capture")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func getStatus(esp32CamAddress string) (*CameraStatus, error) {
	// Get status
	resp, err := httpClient.Get(esp32CamAddress + "/status")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var status CameraStatus
	err = json.NewDecoder(resp.Body).Decode(&status)
	if err != nil {
		return nil, err
	}
	return &status, nil
}

func setStatus(key, value string, esp32CamAddress string) error {
	log.Println("Setting camera status for", esp32CamAddress)
	// Set status
	resp, err := httpClient.Get(esp32CamAddress + "/control?var=" + key + "&val=" + value)
	if err != nil {
		return err
	}
	io.ReadAll(resp.Body)
	resp.Body.Close()
	return nil
}
