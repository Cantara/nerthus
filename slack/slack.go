package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"

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

type slackFile struct {
	Channels []string `json:"channels"`
	TS       string   `json:"ts"`
	Text     string   `json:"initial_comment"`
	File     []byte   `json:"file"`
	Name     string   `json:"name"`
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

type client struct {
	baseurl        string
	token          string
	secretChannel  string
	statusChennel  string
	commandChannel string
	messageChan    chan Message
}

var c client
var statusMessageChan chan string

func NewClient(authToken, secretChannel, statusChennel, commandChannel string) (err error) {
	c = client{
		baseurl:        "https://slack.com",
		token:          authToken,
		secretChannel:  secretChannel,
		statusChennel:  statusChennel,
		commandChannel: commandChannel,
		messageChan:    make(chan Message, 10),
	}
	statusMessageChan = make(chan string, 50)
	go statusMessageWatcher()
	return nil
}

func sendMessage(message, slackId, ts string) (resp slackRespons, err error) {
	return resp, PostAuth(c.baseurl+"/api/chat.postMessage", slackMessage{
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
	resp, err := sendMessage(message, c.secretChannel, id)
	if err != nil {
		return
	}
	idOut = resp.TS
	return
}

func SendFollowupWFile(name, message, id string, file []byte) (idOut string, err error) {
	resp, err := sendFile(c.secretChannel, id, message, name, file)
	if err != nil {
		return
	}
	idOut = resp.TS
	return
}

func sendFile(channel, ts, message, name string, file []byte) (resp slackRespons, err error) {
	return resp, PostFormAuth(c.baseurl+"/api/files.upload", slackFile{
		Channels: []string{channel},
		Text:     message,
		File:     file,
		TS:       ts,
		Name:     name,
	}, &resp)
}

func statusMessageWatcher() {
	ticker := time.NewTicker(5 * time.Second)
	builder := strings.Builder{}
	failCount := 0
	for {
		select {
		case <-ticker.C:
			if builder.Len() <= 0 {
				continue
			}
			_, err := sendMessage(builder.String(), c.statusChennel, "")
			if err != nil {
				if failCount >= 3 {
					failCount = 0
					builder.Reset()
					continue
				}
				failCount++
				time.Sleep(3 * time.Second)
				continue
			}
			builder.Reset()
		case message := <-statusMessageChan:
			builder.WriteString(message)
			builder.WriteString("\n")
		}
	}
}

func SendStatus(message string) {
	statusMessageChan <- message
}

func PostAuth(uri string, data interface{}, out interface{}) (err error) {
	jsonValue, _ := json.Marshal(data)
	client := &http.Client{}
	req, err := http.NewRequest("POST", uri, bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)
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

func PostFormAuth(uri string, file slackFile, out interface{}) (err error) {
	var b bytes.Buffer
	mp := multipart.NewWriter(&b)
	f, err := mp.CreateFormFile("file", file.Name)
	if err != nil {
		return
	}
	f.Write(file.File)
	mp.Close()
	data := url.Values{}
	data.Set("channels", strings.Join(file.Channels, ","))
	data.Set("initial_comment", file.Text)
	if file.TS != "" {
		data.Set("ts", file.TS)
	}
	req, err := http.NewRequest("POST", uri, &b)
	req.Header.Set("Content-Type", mp.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.URL.RawQuery = data.Encode()
	cl := &http.Client{}
	resp, err := cl.Do(req)
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
	_, err = sendMessage(fmt.Sprintf(`%[1]scurl --header "Content-Type: application/json" \
	--header "Authorization: Basic <base64 username and password>" \
  --request POST \
  --data '%s' \
	baseurl/%s%[1]s`, "```", body, endpoint), c.commandChannel, "")
	return
}
