package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	log "github.com/cantara/bragi"
)

type slackMessage struct {
	SlackId string `json:"channel"`
	TS      string `json:"thread_ts"`
	Text    string `json:"text"`
	//	Username    string   `json:"username"`
	Pinned bool `json:"pinned"`
	//	Attachments []string `json:"attachments"`
}

type slackRespons struct {
	Ok               bool        `json:"ok"`
	SlackId          string      `json:"channel"`
	TS               string      `json:"ts"`
	Message          Message     `json:"message"`
	Warning          string      `json:"warning"`
	ResponseMetadata interface{} `json:"response_metadata"`
}

type Message struct {
	BotId      string     `json:"bot_id"`
	Type       string     `json:"type"`
	Text       string     `json:"text"`
	User       string     `json:"user"`
	TS         string     `json:"ts"`
	Team       string     `json:"team"`
	BotProfile BotProfile `json:"bot_profile"`
	Deleted    bool       `json:"deleted"`
	Updated    int        `json:"updated"`
	TeamId     string     `json:"team_id"`
}

type BotProfile struct {
	/*
	   "id": "B02V186UMM5",
	   "app_id": "A02V959QU94",
	   "name": "Nerthus",
	   "icons": {
	     "image_36": "https:\\/\\/a.slack-edge.com\\/80588\\/img\\/plugins\\/app\\/bot_36.png",
	     "image_48": "https:\\/\\/a.slack-edge.com\\/80588\\/img\\/plugins\\/app\\/bot_48.png",
	     "image_72": "https:\\/\\/a.slack-edge.com\\/80588\\/img\\/plugins\\/app\\/service_72.png"
	*/
}

func sendMessage(message, slackId, ts string) (resp slackRespons, err error) {
	return resp, PostAuth("https://slack.com/api/chat.postMessage", slackMessage{
		SlackId: slackId,
		TS:      ts,
		Text:    ":ghost:" + message,
		Pinned:  false,
	}, &resp)
}

func SendBase(message string) (id string, err error) {
	return SendFollowup(message, "")
}

func SendFollowup(message, id string) (idOut string, err error) {
	resp, err := sendMessage(message, os.Getenv("slack_channel_secret"), id)
	if err != nil {
		return
	}
	idOut = resp.TS
	return
}

func SendStatus(message string) (err error) {
	_, err = sendMessage(message, os.Getenv("slack_channel_status"), "")
	return
}

func PostAuth(uri string, data interface{}, out interface{}) (err error) {
	jsonValue, _ := json.Marshal(data)
	client := &http.Client{}
	req, err := http.NewRequest("POST", uri, bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("slack_token"))
	resp, err := client.Do(req)
	if err != nil || out == nil {
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, out)
	if err != nil {
		log.AddError(err).Warning(fmt.Sprintf("%s\t%s", body, data))
	}
	return
}

func SendCommand(endpoint, body string) (err error) {
	_, err = sendMessage(fmt.Sprintf(`curl --header "Content-Type: application/json" \
	--header "Authorization: Basic <base64 username and password>" \
  --request POST \
  --data '%s' \
	baseurl/%s`, body, endpoint), os.Getenv("slack_channel_commands"), "")
	return
}
