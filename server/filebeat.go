package server

import (
	"errors"
	"io/ioutil"
	"strings"

	log "github.com/cantara/bragi"
)

type Filebeat struct {
	Name      string
	ConfigUrl string
	serv      Server
}

func NewFilebeat(name, configUrl string, serv Server) (f Filebeat, err error) {
	f = Filebeat{
		Name:      name,
		ConfigUrl: configUrl,
		serv:      serv,
	}
	return
}

func (f *Filebeat) Create() (id string, err error) {
	script, err := ioutil.ReadFile("./filebeat.sh")
	if err != nil {
		log.AddError(err).Warning("While reading in filebeat script")
		return
	}
	scripts := strings.ReplaceAll(string(script), "<filebeat_configuration>", f.ConfigUrl)
	_, err = f.serv.RunScript(scripts)
	return
}

func (f *Filebeat) Delete() (err error) {
	err = errors.New("NOT IMPLEMENTED") //TODO: FIXME: IMPLEMENT
	return
}
