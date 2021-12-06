package aws

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elbv2"
)

func (a AWS) CreateKeyPair(baseName string) (keyName, keyFingerprint, keyMaterial string, err error) {
	err = a.hasEC2Session()
	if err != nil {
		return
	}
	pairName := baseName + "-key"
	keyResult, err := a.ec2.CreateKeyPair(&ec2.CreateKeyPairInput{
		KeyName: aws.String(pairName),
		KeyType: aws.String("ed25519"),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == "InvalidKeyPair.Duplicate" {
			err = createError{
				text: "Duplicate key pair",
				err:  aerr,
			}
			return
		}
		err = createError{
			text: "Unable to create key pair: " + pairName,
			err:  err,
		}
		return
	}

	return aws.StringValue(keyResult.KeyName), aws.StringValue(keyResult.KeyFingerprint),
		aws.StringValue(keyResult.KeyMaterial), nil
}

func (a AWS) CreateSecurityGroup(serverName, vpcId string) (groupId string, err error) {
	err = a.hasEC2Session()
	if err != nil {
		return
	}
	secName := serverName + "-sg"
	secGroupRes, err := a.ec2.CreateSecurityGroup(&ec2.CreateSecurityGroupInput{
		GroupName:   aws.String(secName),
		Description: aws.String("Security group for server: " + serverName),
		VpcId:       aws.String(vpcId),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case "InvalidVpcID.NotFound":
				err = createError{
					text: fmt.Sprintf("Unable to find VPC with ID %q.", vpcId),
					err:  err,
				}
				return
			case "InvalidGroup.Duplicate":
				err = createError{
					text: fmt.Sprintf("Security group %q already exists.", secName),
					err:  err,
				}
				return
			}
		}
		err = createError{
			text: fmt.Sprintf("Unable to create security group %q.", secName),
			err:  err,
		}
		return
	}
	groupId = aws.StringValue(secGroupRes.GroupId)

	// Add tags to the created security group
	_, err = a.ec2.CreateTags(&ec2.CreateTagsInput{
		Resources: []*string{aws.String(groupId)},
		Tags: []*ec2.Tag{
			{
				Key:   aws.String("Name"),
				Value: aws.String(serverName),
			},
		},
	})
	if err != nil {
		err = createError{
			text: fmt.Sprintf("Could not create tags for sg %s.", groupId),
			err:  err,
		}
		return
	}
	return groupId, nil
}

func (a AWS) CreateServer(serverName, keyName, groupId string) (id string, err error) {
	err = a.hasEC2Session()
	if err != nil {
		return
	}
	// Specify the details of the instance that you want to create
	result, err := a.ec2.RunInstances(&ec2.RunInstancesInput{
		ImageId:          aws.String(os.Getenv("ami")), //ami-0142f6ace1c558c7d"),
		InstanceType:     aws.String("t3.micro"),
		MinCount:         aws.Int64(1),
		MaxCount:         aws.Int64(1),
		SecurityGroupIds: aws.StringSlice([]string{groupId}),
		KeyName:          aws.String(keyName),
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
						Value: aws.String(serverName),
					},
				},
			},
			{
				ResourceType: aws.String("volume"),
				Tags: []*ec2.Tag{
					{
						Key:   aws.String("Name"),
						Value: aws.String(serverName),
					},
				},
			},
		},
	})
	if err != nil {
		err = createError{
			text: fmt.Sprintf("Could not create instance with name %s.", serverName),
			err:  err,
		}
		return
	}

	return aws.StringValue(result.Instances[0].InstanceId), nil
}

func (a AWS) CreateTargetGroup(serverName, vpcId, uriPath string, port int) (anr string, err error) {
	err = a.hasELBSession()
	if err != nil {
		return
	}
	input := &elbv2.CreateTargetGroupInput{
		Name:                       aws.String(serverName + "-tg"),
		Port:                       aws.Int64(int64(port)),
		Protocol:                   aws.String("HTTP"),
		VpcId:                      aws.String(vpcId),
		TargetType:                 aws.String("instance"),
		ProtocolVersion:            aws.String("HTTP1"),
		HealthCheckIntervalSeconds: aws.Int64(5),
		HealthCheckPath:            aws.String(fmt.Sprintf("/%s/health", uriPath)),
		HealthCheckPort:            aws.String("traffic-port"),
		HealthCheckProtocol:        aws.String("HTTP"),
		HealthCheckTimeoutSeconds:  aws.Int64(2),
		HealthyThresholdCount:      aws.Int64(2),
	}

	result, err := a.elb.CreateTargetGroup(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case elbv2.ErrCodeDuplicateTargetGroupNameException:
				err = createError{
					text: "Duplicate target group name.",
					err:  aerr,
				}
			case elbv2.ErrCodeTooManyTargetGroupsException:
				err = createError{
					text: "Too many target groups",
					err:  aerr,
				}
			case elbv2.ErrCodeInvalidConfigurationRequestException:
				err = createError{
					text: "Invalid configuration",
					err:  aerr,
				}
			case elbv2.ErrCodeTooManyTagsException:
				err = createError{
					text: "To many tags",
					err:  aerr,
				}
			}
		}
		err = createError{
			text: fmt.Sprintf("Unable to create target group for server %s.", serverName),
			err:  err,
		}
		return
	}

	return aws.StringValue(result.TargetGroups[0].TargetGroupArn), nil
}

func (a AWS) AddRuleToLoadBalancer(targetGroupArn, elbListenerArn, uriPath string) (ruleArn string, err error) {
	err = a.hasELBSession()
	if err != nil {
		return
	}
	numRules, err := a.GetNumRules(elbListenerArn)
	if err != nil {
		return
	}
	createRuleInput := &elbv2.CreateRuleInput{
		Actions: []*elbv2.Action{
			{
				TargetGroupArn: aws.String(targetGroupArn),
				Type:           aws.String("forward"),
			},
		},
		Conditions: []*elbv2.RuleCondition{
			{
				Field: aws.String("path-pattern"),
				Values: []*string{
					aws.String(fmt.Sprintf("/%s/*", uriPath)),
				},
			},
		},
		ListenerArn: aws.String(elbListenerArn),
		Priority:    aws.Int64(int64(numRules)),
	}

	result, err := a.elb.CreateRule(createRuleInput)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case elbv2.ErrCodePriorityInUseException:
				err = createError{
					text: "Priority in use",
					err:  aerr,
				}
				return
			case elbv2.ErrCodeTooManyTargetGroupsException:
				err = createError{
					text: "Too many target groups",
					err:  aerr,
				}
				return
			case elbv2.ErrCodeTooManyRulesException:
				err = createError{
					text: "Too many rules",
					err:  aerr,
				}
				return
			case elbv2.ErrCodeTargetGroupAssociationLimitException:
				err = createError{
					text: "Target group association limit",
					err:  aerr,
				}
				return
			case elbv2.ErrCodeIncompatibleProtocolsException:
				err = createError{
					text: "Incompatible protocols",
					err:  aerr,
				}
				return
			case elbv2.ErrCodeListenerNotFoundException:
				err = createError{
					text: "Listener not found",
					err:  aerr,
				}
				return
			case elbv2.ErrCodeTargetGroupNotFoundException:
				err = createError{
					text: "Target group not found",
					err:  aerr,
				}
				return
			case elbv2.ErrCodeInvalidConfigurationRequestException:
				err = createError{
					text: "Invalid configuration",
					err:  aerr,
				}
				return
			case elbv2.ErrCodeTooManyRegistrationsForTargetIdException:
				err = createError{
					text: "Too many registrations for target id",
					err:  aerr,
				}
				return
			case elbv2.ErrCodeTooManyTargetsException:
				err = createError{
					text: "Too many targets",
					err:  aerr,
				}
				return
			case elbv2.ErrCodeUnsupportedProtocolException:
				err = createError{
					text: "Unsupported protocol",
					err:  aerr,
				}
				return
			case elbv2.ErrCodeTooManyActionsException:
				err = createError{
					text: "Too many actions",
					err:  aerr,
				}
				return
			case elbv2.ErrCodeInvalidLoadBalancerActionException:
				err = createError{
					text: "Incalid loadbalancer action",
					err:  aerr,
				}
				return
			case elbv2.ErrCodeTooManyUniqueTargetGroupsPerLoadBalancerException:
				err = createError{
					text: "Too many unique target groups per loadbalancer",
					err:  aerr,
				}
				return
			case elbv2.ErrCodeTooManyTagsException:
				err = createError{
					text: "Too many tags",
					err:  aerr,
				}
				return
			}
		}
		err = createError{
			text: fmt.Sprintf("Unable to add rule to loadbalancer %s.", uriPath),
			err:  err,
		}
	}

	return aws.StringValue(result.Rules[0].RuleArn), nil
}
