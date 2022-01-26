package server

type Version int

const (
	JAVA_ONE_EIGHT Version = iota
	JAVA_ONE_ELEVEN
)

func (v Version) String() string {
	return []string{"zulu8-jdk", "zulu11-jdk"}[v]
}

type Java struct {
	Server    Server
	Version   Version
	installed bool
}

func NewJava(version Version, serv Server) (j Java, err error) {
	j = Java{
		Server:  serv,
		Version: version,
	}
	return
}

func (j Java) isInstalled() (installed bool, err error) {
	script := ""
	switch j.Version {
	case JAVA_ONE_EIGHT:
		script = "yum list installed | grep zulu8"
	case JAVA_ONE_ELEVEN:
		script = "yum list installed | grep zulu11"
	default:
		return
	}
	stdout, err := j.Server.RunScript(script)
	if err != nil {
		return
	}
	installed = stdout != ""
	return
}

func (j *Java) Create() (id string, err error) {
	installed, err := j.isInstalled()
	if err != nil || installed {
		j.installed = false
		return
	}
	script := "sudo yum install -y https://cdn.azul.com/zulu/bin/zulu-repo-1.0.0-1.noarch.rpm\n"
	switch j.Version {
	case JAVA_ONE_EIGHT:
		script += "sudo yum install -y zulu8-jdk"
	case JAVA_ONE_ELEVEN:
		script += "sudo yum install -y zulu11-jdk"
	default:
		return
	}
	_, err = j.Server.RunScript(script)
	if err != nil {
		return
	}
	j.installed = true
	return
}

func (j *Java) Delete() (err error) {
	if !j.installed {
		return
	}
	script := "yum list installed | grep "
	switch j.Version {
	case JAVA_ONE_EIGHT:
		script += "zulu8"
	case JAVA_ONE_ELEVEN:
		script += "zulu11"
	default:
		return
	}
	script += " | xargs sudo yum autoremove -y"
	_, err = j.Server.RunScript(script)
	return
}
