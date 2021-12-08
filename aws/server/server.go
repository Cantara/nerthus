package server

import (
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	log "github.com/cantara/bragi"
	"github.com/cantara/nerthus/aws/key"
	"github.com/cantara/nerthus/aws/security"
	"github.com/cantara/nerthus/aws/util"
)

type Server struct {
	Name      string
	Id        string
	PublicDNS string
	key       key.Key
	group     security.Group
	ec2       *ec2.EC2
}

func NewServer(name string, key key.Key, group security.Group, e2 *ec2.EC2) (s Server, err error) {
	err = util.CheckEC2Session(e2)
	if err != nil {
		return
	}
	s = Server{
		Name:  name,
		key:   key,
		group: group,
		ec2:   e2,
	}
	return
}

func (s *Server) Create() (id string, err error) {
	// Specify the details of the instance that you want to create
	result, err := s.ec2.RunInstances(&ec2.RunInstancesInput{
		ImageId:          aws.String(os.Getenv("ami")), //ami-0142f6ace1c558c7d"),
		InstanceType:     aws.String("t3.micro"),
		MinCount:         aws.Int64(1),
		MaxCount:         aws.Int64(1),
		SecurityGroupIds: aws.StringSlice([]string{s.group.Id}),
		KeyName:          aws.String(s.key.Name),
		BlockDeviceMappings: []*ec2.BlockDeviceMapping{
			{
				DeviceName: aws.String("/dev/xvda"),
				Ebs: &ec2.EbsBlockDevice{
					VolumeSize: aws.Int64(20),
					VolumeType: aws.String("gp3"),
				},
			},
		},
		TagSpecifications: []*ec2.TagSpecification{
			{
				ResourceType: aws.String("instance"),
				Tags: []*ec2.Tag{
					{
						Key:   aws.String("Name"),
						Value: aws.String(s.Name),
					},
				},
			},
			{
				ResourceType: aws.String("volume"),
				Tags: []*ec2.Tag{
					{
						Key:   aws.String("Name"),
						Value: aws.String(s.Name),
					},
				},
			},
		},
	})
	if err != nil {
		err = util.CreateError{
			Text: fmt.Sprintf("Could not create instance with name %s.", s.Name),
			Err:  err,
		}
		return
	}
	s.Id = aws.StringValue(result.Instances[0].InstanceId)
	id = s.Id
	return
}

func (s *Server) Delete() (err error) {
	_, err = s.ec2.TerminateInstances(&ec2.TerminateInstancesInput{
		InstanceIds: []*string{
			aws.String(s.Id),
		},
	})
	if err != nil {
		return
	}
	s.WaitForTerminated()
	return
}

func (s Server) WaitForRunning() (err error) {
	status, err := GetServerStatus(s.Id, s.ec2)
	for count := 0; (err != nil || status != "running") && count < 60; count++ {
		if err != nil {
			log.AddError(err).Warning("Getting running state of new instance")
		}
		time.Sleep(1 * time.Second)
		status, err = GetServerStatus(s.Id, s.ec2)
	}
	return
}

func (s Server) WaitForTerminated() (err error) {
	status, err := GetServerStatus(s.Id, s.ec2)
	for count := 0; (err != nil || status != "terminated") && count < 60; count++ {
		if err != nil {
			log.AddError(err).Warning("Getting state of instance")
		}
		time.Sleep(10 * time.Second)
		status, err = GetServerStatus(s.Id, s.ec2)
	}
	return
}

func (s *Server) GetPublicDNS() (publicDNS string, err error) {
	if s.PublicDNS != "" {
		publicDNS = s.PublicDNS
		return
	}
	result, err := s.ec2.DescribeInstances(&ec2.DescribeInstancesInput{
		InstanceIds: []*string{aws.String(s.Id)},
	})
	if err != nil {
		err = util.CreateError{
			Text: "Unable to describe Instance",
			Err:  err,
		}
		return
	}
	s.PublicDNS = aws.StringValue(result.Reservations[0].Instances[0].PublicDnsName)
	publicDNS = s.PublicDNS
	return
}

func GetServerStatus(id string, e *ec2.EC2) (status string, err error) {
	result, err := e.DescribeInstances(&ec2.DescribeInstancesInput{
		InstanceIds: []*string{aws.String(id)},
	})
	if err != nil {
		return
	}
	status = aws.StringValue(result.Reservations[0].Instances[0].State.Name)
	return
}
