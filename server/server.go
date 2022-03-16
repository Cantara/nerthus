package server

import (
	"errors"
	"fmt"
	"io"
	"os/exec"
	"time"

	log "github.com/cantara/bragi"
)

type Server struct {
	publicDNS string
	pemName   string
}

func NewServer(publicDNS, pemName string) (s Server, err error) {
	s = Server{
		publicDNS: publicDNS,
		pemName:   pemName,
	}
	return
}

func (s *Server) RunScript(script string) (stdout string, err error) {
	script = script + `
history -c
exit
`
	cmd := exec.Command("ssh", "-o", "ConnectTimeout=5", "-o", "ConnectionAttempts=3", "-o", "UserKnownHostsFile=/dev/null", "-o", "StrictHostKeyChecking=no",
		fmt.Sprintf("ec2-user@%s", s.publicDNS), "-i", "./"+s.pemName, "/bin/bash -s")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.AddError(err).Warning("While creating stdin pipe")
		return
	}
	defer stdin.Close()
	io.WriteString(stdin, script)

	stdoutB, err := cmd.Output()
	if err != nil {
		var eerr *exec.ExitError
		if errors.As(err, &eerr) {
			log.AddError(errors.New(string(eerr.Stderr))).Warning("Exit error from ssh run command for host %s with pemName %s and command %s", s.publicDNS, s.pemName, cmd.String())
		}

		return
	}
	stdout = string(stdoutB)
	return
}

func (s Server) WaitForConnection() (err error) {
	for i := 0; i < 30; i++ {
		_, err = s.RunScript(`echo "ping"`)
		if err == nil {
			return nil
		}
		log.AddError(err).Info("While waiting for connection to server %s", s.publicDNS)
		time.Sleep(10 * time.Second)
	}
	return errors.New("Unable to connect to server")
}

func (s Server) AddAutoUpdate() (err error) {
	script := `
cat <<'EOF' > ~/CRON
MAILTO=""
*/30 * * * * sudo yum update -y > /dev/null
EOF

crontab ~/CRON
`
	_, err = s.RunScript(script)
	return
}

func (s Server) Update() (err error) {
	script := "sudo yum update -y"
	_, err = s.RunScript(script)
	return
}
