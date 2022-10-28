package server

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
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
	ec2                *ec2.Client
	created            bool
}

func NewServer(name, scope string, key key.Key, group security.Group, e2 *ec2.Client) (s Server, err error) {
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

func GetServer(name, scope string, key key.Key, group security.Group, e2 *ec2.Client) (s Server, err error) {
	result, err := e2.DescribeInstances(context.Background(), &ec2.DescribeInstancesInput{
		Filters: []ec2types.Filter{
			{
				Name: aws.String("tag:Name"),
				Values: []string{
					name,
				},
			},
		},
	})
	if err != nil {
		return
	}
	if len(result.Reservations) < 1 {
		err = fmt.Errorf("No servers with name %s", name)
		return
	}
	/* if len(result.Reservations) > 1 {
		err = fmt.Errorf("Too many servers with name %s", name)
	} */
	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			if instance.State.Name != ec2types.InstanceStateNameRunning {
				continue
			}
			for _, tag := range instance.Tags {
				if aws.ToString(tag.Key) == "Scope" && aws.ToString(tag.Value) == scope {
					if len(instance.BlockDeviceMappings) < 1 || len(instance.NetworkInterfaces) < 1 {
						continue
					}
					s = Server{
						Name:               name,
						Scope:              scope,
						key:                key,
						group:              group,
						Id:                 aws.ToString(instance.InstanceId),
						PublicDNS:          aws.ToString(instance.PublicDnsName),
						VolumeId:           aws.ToString(instance.BlockDeviceMappings[0].Ebs.VolumeId),
						NetworkInterfaceId: aws.ToString(instance.NetworkInterfaces[0].NetworkInterfaceId),
						ImageId:            aws.ToString(instance.ImageId),
						ec2:                e2,
					}
					return
				}
			}
		}
	}
	err = fmt.Errorf("Server name %s was not in scope %s", name, scope)
	return
}

func NameAvailable(name string, e2 *ec2.Client) (available bool, err error) {
	result, err := e2.DescribeInstances(context.Background(), &ec2.DescribeInstancesInput{
		Filters: []ec2types.Filter{
			{
				Name: aws.String("tag:Name"),
				Values: []string{
					name,
				},
			},
		},
	})
	if err != nil {
		return
	}
	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			if instance.State.Name == ec2types.InstanceStateNameTerminated {
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
	result, err := s.ec2.RunInstances(context.Background(), &ec2.RunInstancesInput{
		ImageId:          aws.String(s.ImageId), //ami-0142f6ace1c558c7d"),
		InstanceType:     "t3.micro",
		MinCount:         aws.Int32(1),
		MaxCount:         aws.Int32(1),
		SecurityGroupIds: []string{s.group.Id},
		KeyName:          aws.String(s.key.Name),
		BlockDeviceMappings: []ec2types.BlockDeviceMapping{
			{
				DeviceName: aws.String("/dev/xvda"),
				Ebs: &ec2types.EbsBlockDevice{
					VolumeSize: aws.Int32(20),
					VolumeType: "gp3",
				},
			},
		},
		MetadataOptions: &ec2types.InstanceMetadataOptionsRequest{
			HttpTokens: ec2types.HttpTokensStateRequired,
		},
		TagSpecifications: []ec2types.TagSpecification{
			{
				ResourceType: "instance",
				Tags: []ec2types.Tag{
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
				ResourceType: "volume",
				Tags: []ec2types.Tag{
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
	s.Id = aws.ToString(result.Instances[0].InstanceId)
	s.NetworkInterfaceId = aws.ToString(result.Instances[0].NetworkInterfaces[0].NetworkInterfaceId)
	//s.VolumeId = aws.ToString(result.Instances[0].BlockDeviceMappings[0].Ebs.VolumeId)
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
	_, err = s.ec2.TerminateInstances(context.Background(), &ec2.TerminateInstancesInput{
		InstanceIds: []string{
			s.Id,
		},
	})
	if err != nil {
		return
	}
	err = s.WaitUntilTerminated()
	return
}

func (s Server) WaitUntilRunning() (err error) {
	err = ec2.NewInstanceRunningWaiter(s.ec2).Wait(context.Background(), &ec2.DescribeInstancesInput{
		InstanceIds: []string{s.Id},
	}, 5*time.Minute)
	if err != nil {
		return
	}
	return
}

func (s Server) WaitUntilTerminated() (err error) {
	err = ec2.NewInstanceTerminatedWaiter(s.ec2).Wait(context.Background(), &ec2.DescribeInstancesInput{
		InstanceIds: []string{s.Id},
	}, 5*time.Minute)
	return
}

func (s Server) WaitUntilNetworkAvailable() (err error) {
	err = ec2.NewNetworkInterfaceAvailableWaiter(s.ec2).Wait(context.Background(), &ec2.DescribeNetworkInterfacesInput{
		NetworkInterfaceIds: []string{s.NetworkInterfaceId},
	}, 5*time.Minute)
	return
}

func (s *Server) GetPublicDNS() (publicDNS string, err error) {
	if s.PublicDNS != "" {
		publicDNS = s.PublicDNS
		return
	}
	result, err := s.ec2.DescribeInstances(context.Background(), &ec2.DescribeInstancesInput{
		InstanceIds: []string{s.Id},
	})
	if err != nil {
		err = util.CreateError{
			Text: "Unable to describe Instance",
			Err:  err,
		}
		return
	}
	s.PublicDNS = aws.ToString(result.Reservations[0].Instances[0].PublicDnsName)
	publicDNS = s.PublicDNS
	return
}

func (s *Server) GetVolumeId() (volumeId string, err error) {
	if s.VolumeId != "" {
		volumeId = s.VolumeId
		return
	}
	result, err := s.ec2.DescribeInstances(context.Background(), &ec2.DescribeInstancesInput{
		InstanceIds: []string{s.Id},
	})
	if err != nil {
		err = util.CreateError{
			Text: "Unable to describe Instance",
			Err:  err,
		}
		return
	}
	s.VolumeId = aws.ToString(result.Reservations[0].Instances[0].BlockDeviceMappings[0].Ebs.VolumeId)
	volumeId = s.VolumeId
	return
}
