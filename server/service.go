package server

import (
	"io/ioutil"
	"strconv"
	"strings"

	log "github.com/cantara/bragi"
)

type Service struct {
	Name          string
	UpdateProp    string
	LocalOverride string
	HealthReport  string
	Path          string
	Port          string
	user          User
	serv          Server
}

func NewService(name, updateProp, localOverride, healthReport, path string, port int, user User, serv Server) (s Service, err error) {
	s = Service{
		Name:          name,
		UpdateProp:    updateProp,
		LocalOverride: localOverride,
		HealthReport:  healthReport,
		Path:          path,
		Port:          strconv.Itoa(port),
		user:          user,
		serv:          serv,
	}
	return
}

func (s *Service) Create() (id string, err error) {
	script, err := ioutil.ReadFile("./new_devtest_server.sh")
	if err != nil {
		log.AddError(err).Warning("While reading in base script")
		return
	}
	scripts := strings.ReplaceAll(string(script), "<application>", s.Name)
	scripts = strings.ReplaceAll(scripts, "<username>", s.user.Name)
	scripts = strings.ReplaceAll(scripts, "<semantic_update_service_properties>", s.UpdateProp)
	scripts = strings.ReplaceAll(scripts, "<local_override_properties>", s.LocalOverride)
	scripts = strings.ReplaceAll(scripts, "<port>", s.Port)
	scripts = strings.ReplaceAll(scripts, "<path>", s.Path)
	scripts = strings.ReplaceAll(scripts, "<health_report_enpoint>", s.HealthReport)
	_, err = s.serv.RunScript(scripts)
	return
}

func (s *Service) Delete() (err error) {
	script := `
./su_to_<username>.sh
crontab -r
~/scripts/kill-vili.sh
pkill -9 Visuale
history -c
exit
`
	script = strings.ReplaceAll(script, "<username>", s.user.Name)
	_, err = s.serv.RunScript(script)
	return
}
