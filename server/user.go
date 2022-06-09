package server

import (
	"errors"
	"strings"
)

type User struct {
	Name  string
	serv  Server
	added bool
}

func NewUser(name string, serv Server) (u User, err error) {
	u = User{
		Name: ToFriendlyName(name),
		serv: serv,
	}
	return
}

func (u User) exist() (exist bool, err error) {
	script := "cat /etc/passwd | grep " + u.Name
	stdout, err := u.serv.RunScript(script)
	if err != nil {
		return
	}
	exist = stdout != ""
	return
}

func (u *User) Create() (id string, err error) {
	exist, err := u.exist()
	if err != nil || exist {
		u.added = false
		if err != nil {
			return
		}
		err = errors.New("User already exists")
		return
	}
	script := `
sudo adduser <username>

cat <<'EOF' > su_to_<username>.sh
#!/bin/env sh
sudo su - <username>
EOF

chmod +x su_to_<username>.sh
`
	script = strings.ReplaceAll(script, "<username>", u.Name)
	_, err = u.serv.RunScript(script)
	if err != nil {
		return
	}
	u.added = true
	id = u.Name
	return
}

func (u *User) Delete() (err error) {
	if !u.added {
		return
	}
	script := `
sudo userdel -rf <username>
rm su_to_<username>.sh
`
	script = strings.ReplaceAll(script, "<username>", u.Name)
	_, err = u.serv.RunScript(script)
	return
}
