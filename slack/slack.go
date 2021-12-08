package slack

import (
	"os"

	"github.com/cantara/nerthus/whydah"
)

type slackMessage struct {
	SlackId string `json:"recepientId"`
	Message string `json:"message"`
	//	Username    string   `json:"username"`
	Pinned bool `json:"pinned"`
	//	Attachments []string `json:"attachments"`
}

func SendMessage(message, slackId string) (err error) {
	return whydah.PostAuth(os.Getenv("entraos_api_uri")+"/slack/api/message", slackMessage{
		SlackId: slackId,
		Message: message,
		Pinned:  false,
	}, nil)
}

func SendServer(message string) (err error) {
	return SendMessage(message, os.Getenv("slack_channel"))
}

func SendStatus(message string) (err error) {
	return SendMessage(message, "C02PUP3QTL4")
}
