package server

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/cantara/nerthus/aws/key"
	"github.com/cantara/nerthus/aws/security"
	"github.com/cantara/nerthus/aws/util"
)

type Server struct {
	Name               string
	Scope              string
	Id                 string
	PublicDNS          string
	VolumeId           string `json:"volume_id"`
	NetworkInterfaceId string `json:"network_interface_id"`
	ImageId            string `json:"image_id"`
	key                key.Key
	group              security.Group
	ec2                *ec2.EC2
	created            bool
}

func NewServer(name, scope string, key key.Key, group security.Group, e2 *ec2.EC2) (s Server, err error) {
	err = util.CheckEC2Session(e2)
	if err != nil {
		return
	}
	s = Server{
		Name:    name,
		Scope:   scope,
		ImageId: os.Getenv("ami"),
		key:     key,
		group:   group,
		ec2:     e2,
	}
	return
}

func GetServer(name, scope string, key key.Key, group security.Group, e2 *ec2.EC2) (s Server, err error) {
	result, err := e2.DescribeInstances(&ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("tag:Name"),
				Values: []*string{
					aws.String(name),
				},
			},
		},
	})
	if err != nil {
		return
	}
	if len(result.Reservations) < 1 {
		err = fmt.Errorf("No servers with name %s", name)
	}
	if len(result.Reservations) > 1 {
		err = fmt.Errorf("Too many servers with name %s", name)
	}
	for _, tag := range result.Reservations[0].Instances[0].Tags {
		if aws.StringValue(tag.Key) == "Scope" && aws.StringValue(tag.Value) == scope {
			s = Server{
				Name:               name,
				Scope:              scope,
				key:                key,
				group:              group,
				Id:                 aws.StringValue(result.Reservations[0].Instances[0].InstanceId),
				PublicDNS:          aws.StringValue(result.Reservations[0].Instances[0].PublicDnsName),
				VolumeId:           aws.StringValue(result.Reservations[0].Instances[0].BlockDeviceMappings[0].Ebs.VolumeId),
				NetworkInterfaceId: aws.StringValue(result.Reservations[0].Instances[0].NetworkInterfaces[0].NetworkInterfaceId),
				ImageId:            aws.StringValue(result.Reservations[0].Instances[0].ImageId),
				ec2:                e2,
			}
			return
		}
	}
	err = fmt.Errorf("Server name %s was not in scope %s", name, scope)
	return
}

func NameAvailable(name string, e2 *ec2.EC2) (available bool, err error) {
	result, err := e2.DescribeInstances(&ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("tag:Name"),
				Values: []*string{
					aws.String(name),
				},
			},
		},
	})
	if err != nil {
		return
	}
	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			if aws.StringValue(instance.State.Name) == "terminated" {
				continue
			}
			available = false
			return
		}
	}
	available = true
	return
}

func (s *Server) Create() (id string, err error) {
	// Specify the details of the instance that you want to create
	result, err := s.ec2.RunInstances(&ec2.RunInstancesInput{
		ImageId:          aws.String(s.ImageId), //ami-0142f6ace1c558c7d"),
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
					{
						Key:   aws.String("Scope"),
						Value: aws.String(s.Scope),
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
					{
						Key:   aws.String("Scope"),
						Value: aws.String(s.Scope),
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
	s.NetworkInterfaceId = aws.StringValue(result.Instances[0].NetworkInterfaces[0].NetworkInterfaceId)
	//s.VolumeId = aws.StringValue(result.Instances[0].BlockDeviceMappings[0].Ebs.VolumeId)
	id = s.Id
	s.created = true
	return
}

func (s *Server) Delete() (err error) {
	if !s.created {
		return
	}
	err = util.CheckEC2Session(s.ec2)
	if err != nil {
		return
	}
	_, err = s.ec2.TerminateInstances(&ec2.TerminateInstancesInput{
		InstanceIds: []*string{
			aws.String(s.Id),
		},
	})
	if err != nil {
		return
	}
	s.WaitUntilTerminated()
	return
}

func (s Server) WaitUntilRunning() (err error) {
	err = s.ec2.WaitUntilInstanceRunning(&ec2.DescribeInstancesInput{
		InstanceIds: []*string{aws.String(s.Id)},
	})
	if err != nil {
		return
	}
	return
}

func (s Server) WaitUntilTerminated() (err error) {
	err = s.ec2.WaitUntilInstanceTerminated(&ec2.DescribeInstancesInput{
		InstanceIds: []*string{aws.String(s.Id)},
	})
	return
}

func (s Server) WaitUntilNetworkAvailable() (err error) {
	err = s.ec2.WaitUntilNetworkInterfaceAvailable(&ec2.DescribeNetworkInterfacesInput{
		NetworkInterfaceIds: []*string{aws.String(s.NetworkInterfaceId)},
	})
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

func (s *Server) GetVolumeId() (volumeId string, err error) {
	if s.VolumeId != "" {
		volumeId = s.VolumeId
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
	s.VolumeId = aws.StringValue(result.Reservations[0].Instances[0].BlockDeviceMappings[0].Ebs.VolumeId)
	volumeId = s.VolumeId
	return
}
