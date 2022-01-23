package aws

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elbv2"
	log "github.com/cantara/bragi"
	"github.com/cantara/nerthus/aws/key"
	"github.com/cantara/nerthus/aws/security"
	serverlib "github.com/cantara/nerthus/aws/server"
	"github.com/cantara/nerthus/aws/util"
	"github.com/cantara/nerthus/aws/vpc"
	"github.com/cantara/nerthus/crypto"
	"github.com/cantara/nerthus/slack"
)

func CheckNameLen(name string) error {
	const maxNameLen = 29
	const minNameLen = 3
	if len(name) < minNameLen {
		return fmt.Errorf("Min name len is: %d provided name len is %d.", minNameLen, len(name))
	}
	if len(name) > maxNameLen {
		return fmt.Errorf("Max name len in aws is: %d provided name len is %d.", maxNameLen, len(name))
	}
	return nil
}

type AWS struct {
	ec2 *ec2.EC2
	elb *elbv2.ELBV2
}

func (a AWS) GetELB() *elbv2.ELBV2 {
	return a.elb
}

func (a *AWS) NewEC2(c client.ConfigProvider) {
	if a.ec2 != nil {
		return
	}
	a.ec2 = ec2.New(c)
}

func (a AWS) hasEC2Session() error {
	if a.ec2 == nil {
		return fmt.Errorf("No ec2 session found")
	}
	return nil
}

func (a *AWS) NewELB(c client.ConfigProvider) {
	if a.elb != nil {
		return
	}
	a.elb = elbv2.New(c)
}

func (a AWS) hasELBSession() error {
	if a.elb == nil {
		return fmt.Errorf("No elb session found")
	}
	return nil
}

func cleanup(object, logMessage string, obj util.AWSObject) func() {
	return func() {
		s := fmt.Sprintf(":ghost: Cleaning up: %s", object)
		log.Info(s)
		slack.SendStatus(s)
		err := obj.Delete()
		if err != nil {
			log.AddError(err).Crit(logMessage)
		}
	}
}

func NewStack() Stack {
	return Stack{}
}

type Stack struct {
	funcs []func()
}

func (s *Stack) Push(fun func()) {
	s.funcs = append(s.funcs, fun)
}

func (s *Stack) Pop() func() {
	if s.Empty() {
		return nil
	}
	fun := s.Last()
	s.funcs = s.funcs[:s.Len()-1]
	return fun
}

func (s Stack) Len() int {
	return len(s.funcs)
}

func (s Stack) Last() func() {
	if s.Empty() {
		return nil
	}
	return s.funcs[s.Len()-1]
}

func (s Stack) First() func() {
	if s.Empty() {
		return nil
	}
	return s.funcs[0]
}

func (s Stack) Empty() bool {
	return s.Len() == 0
}

type cryptoData struct {
	Scope         string         `json:"scope"`
	VPC           vpc.VPC        `json:"vpc"`
	Key           key.Key        `json:"key"`
	SecurityGroup security.Group `json:"security_group"`
}

func Encrypt(scope string, v vpc.VPC, k key.Key, sg security.Group) (encrypted string, err error) {
	data := cryptoData{
		VPC:           v,
		Key:           k,
		Scope:         scope,
		SecurityGroup: sg,
	}
	fmt.Println(data)
	b, err := json.Marshal(data)
	if err != nil {
		return
	}
	encrypted, err = crypto.Encrypt(b)
	return
}

func Decrypt(dataCrypt string, a *AWS) (scope string, v vpc.VPC, k key.Key, sg security.Group, err error) {
	data, err := crypto.Decrypt(dataCrypt)
	if err != nil {
		return
	}
	var cd cryptoData
	err = json.Unmarshal(data, &cd)
	if err != nil {
		return
	}
	scope = cd.Scope
	v = cd.VPC
	k = cd.Key.WithEC2(a.ec2)
	k.Scope = scope
	sg = cd.SecurityGroup.WithEC2(a.ec2)
	sg.Scope = scope
	fmt.Println(cd)
	return
}

func GetPublicDNS(server, scope string, a *AWS) (publicDNS string, err error) {
	serv, err := serverlib.GetServer(server, scope, key.Key{}, security.Group{}, a.ec2)
	if err != nil {
		return
	}
	publicDNS = serv.PublicDNS
	return
}
