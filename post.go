package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
	"net/url"
	"os"
	//log "github.com/cantara/bragi"
)

type slackMessage struct {
	SlackId string `json:"recepientId"`
	Message string `json:"message"`
	//	Username    string   `json:"username"`
	Pinned bool `json:"pinned"`
	//	Attachments []string `json:"attachments"`
}

func sendSlackMessage(message string) (err error) {
	if token == "" {
		token, err = getWhydahAuthToken()
		for count := 0; err != nil && count < 10; count++ {
			token, err = getWhydahAuthToken()
		}
	}
	return postAuth(os.Getenv("entraos_api_uri")+"/slack/api/message", slackMessage{
		SlackId: os.Getenv("slack_channel"),
		Message: message,
		Pinned:  false,
	}, nil, token)
}

type applicationcredential struct {
	Params applicationCredentialParams `xml:"params"`
}

type applicationCredentialParams struct {
	AppId     string `xml:"applicationID"`
	AppName   string `xml:"applicationName"`
	AppSecret string `xml:"applicationSecret"`
}

type applicationtoken struct {
	Params applicationTokenParams `xml:"params"`
}

type applicationTokenParams struct {
	AppTokenId string `xml:"applicationtokenID"`
	AppId      string `xml:"applicationid"`
	AppName    string `xml:"applicationName"`
	expires    int    `xml:"expires"`
}

func getWhydahAuthToken() (token string, err error) {
	appCred := applicationcredential{
		Params: applicationCredentialParams{
			AppId:     os.Getenv("whydah_application_id"),
			AppName:   os.Getenv("whydah_application_name"),
			AppSecret: os.Getenv("whydah_application_secret"),
		},
	}
	appCredXML, err := xml.Marshal(appCred)
	if err != nil {
		return
	}
	data := url.Values{
		"applicationcredential": {string(appCredXML)},
	}
	resp, err := http.PostForm(os.Getenv("whydah_uri")+"/tokenservice/logon", data)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	var tokenData applicationtoken
	err = xml.Unmarshal(body, &tokenData)
	if err != nil {
		return
	}
	token = tokenData.Params.AppTokenId
	return
}

func postAuth(uri string, data interface{}, out interface{}, token string) (err error) {
	jsonValue, _ := json.Marshal(data)
	client := &http.Client{}
	req, err := http.NewRequest("POST", uri, bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
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
	return
}
