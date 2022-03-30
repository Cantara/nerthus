package server

import (
	"embed"
	"errors"
	"strings"

	log "github.com/cantara/bragi"
)

type Filebeat struct {
	url      string
	password string
	serv     Server
}

func NewFilebeat(password string, serv Server) (f Filebeat, err error) {
	f = Filebeat{
		url:      "https://cloud.humio.com:443/api/v1/ingest/elastic-bulk",
		password: password,
		serv:     serv,
	}
	return
}

//go:embed filebeat.sh
var fsFB embed.FS

func (f *Filebeat) Create() (id string, err error) {
	script, err := fsFB.ReadFile("filebeat.sh")
	if err != nil {
		log.AddError(err).Warning("While reading in filebeat script")
		return
	}
	scripts := strings.ReplaceAll(string(script), "<filebeat_url>", f.url)
	scripts = strings.ReplaceAll(scripts, "<filebeat_password>", f.password)
	_, err = f.serv.RunScript(scripts)
	return
}

func (f *Filebeat) Delete() (err error) {
	err = errors.New("NOT IMPLEMENTED") //TODO: FIXME: IMPLEMENT
	return
}
