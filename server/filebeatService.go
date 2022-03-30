package server

import (
	"embed"
	"strings"

	log "github.com/cantara/bragi"
)

type FilebeatService struct {
	serverName string
	artifactId string
	userName   string
	serv       Server
	added      bool
}

func NewFilebeatService(serverName, artifactId, userName string, serv Server) (f FilebeatService, err error) {
	f = FilebeatService{
		serverName: serverName,
		artifactId: artifactId,
		userName:   userName,
		serv:       serv,
	}
	return
}

//go:embed filebeat_service.sh
var fsFBS embed.FS

func (f *FilebeatService) Create() (id string, err error) {
	script, err := fsFBS.ReadFile("filebeat_service.sh")
	if err != nil {
		log.AddError(err).Warning("While reading in filebeat script")
		f.added = false
		return
	}
	scripts := strings.ReplaceAll(string(script), "<filebeat_server_name>", f.serverName)
	scripts = strings.ReplaceAll(scripts, "<filebeat_artifact_id>", f.artifactId)
	scripts = strings.ReplaceAll(scripts, "<filebeat_user_name>", f.userName)
	_, err = f.serv.RunScript(scripts)
	f.added = true
	return
}

func (f *FilebeatService) Delete() (err error) {
	if !f.added {
		return
	}
	script := `
sudo rm /etc/filebeat/inputs.d/<file_name>.yml
sudo service filebeat restart
`
	script = strings.ReplaceAll(script, "<file_name>", f.serverName)
	_, err = f.serv.RunScript(script)
	return
}
