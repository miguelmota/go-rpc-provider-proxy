package slack

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// SendNotificationInput ...
type SendNotificationInput struct {
	WebhookURL string
	Message    string
	Channel    string
	Username   string
	IconEmoji  string
}

// RequestBody ...
type RequestBody struct {
	Text      string `json:"text"`
	Channel   string `json:"channel"`
	Username  string `json:"username"`
	IconEmoji string `json:"icon_emoji"`
}

// SendNotification ...
func SendNotification(input *SendNotificationInput) error {
	slackBody, err := json.Marshal(&RequestBody{
		Text:      input.Message,
		Channel:   input.Channel,
		Username:  input.Username,
		IconEmoji: input.IconEmoji,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, input.WebhookURL, bytes.NewBuffer(slackBody))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	resBody := string(buf)
	if resBody != "ok" {
		fmt.Println(resBody)
		return errors.New("Not OK")
	}

	return nil
}
