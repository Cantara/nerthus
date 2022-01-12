package server

import (
	"errors"
	"fmt"
	"io"
	"os/exec"

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
	cmd := exec.Command("ssh", "-o", "UserKnownHostsFile=/dev/null", "-o", "StrictHostKeyChecking=no",
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
			log.Crit(string(eerr.Stderr))
		}

		return
	}
	stdout = string(stdoutB)
	return
}

func (s Server) Update() (err error) {
	script := "sudo yum update -y"
	_, err = s.RunScript(script)
	return
}
