package pushover

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/kabukky/homeautomation/utils"
)

type MessagePayload struct {
	Token   string `json:"token"`
	User    string `json:"user"`
	Message string `json:"message"`
}

var (
	httpClient = http.Client{
		Timeout: 10 * time.Second,
	}
	pushoverURL = "https://api.pushover.net/1/messages.json"
)

func SendPush(message string) error {
	payload := MessagePayload{
		Token:   utils.PushoverToken,
		User:    utils.PushoverUserKey,
		Message: message,
	}
	reqBody, err := json.Marshal(&payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, pushoverURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respData, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return errors.New("pushover reponded with status code " + strconv.Itoa(resp.StatusCode) + " and message " + string(respData))

	}
	return nil
}
