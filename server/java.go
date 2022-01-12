package server

type Version int

const (
	JAVA_ONE_EIGHT Version = iota
	JAVA_ONE_ELEVEN
)

func (v Version) String() string {
	return []string{"java-1.8.0-openjdk", "java-11-openjdk"}[v]
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
		script = "yum list installed | grep java-1.8.0-openjdk"
	case JAVA_ONE_ELEVEN:
		script = "yum list installed | grep java-11-openjdk"
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
	script := ""
	switch j.Version {
	case JAVA_ONE_EIGHT:
		script = "sudo yum install -y java-1.8.0-openjdk"
	case JAVA_ONE_ELEVEN:
		script = "sudo amazon-linux-extras install -y java-openjdk11"
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
	script := ""
	switch j.Version {
	case JAVA_ONE_EIGHT:
		script = "sudo yum remove -y java-1.8.0-openjdk"
	case JAVA_ONE_ELEVEN:
		script = "sudo yum remove -y java-11-openjdk"
	}
	_, err = j.Server.RunScript(script)
	return
}
