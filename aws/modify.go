package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elbv2"
)

func (a AWS) AddBaseAuthorization(groupId, loadbalancerId string, port int) (err error) {
	err = a.hasEC2Session()
	if err != nil {
		return
	}
	input := &ec2.AuthorizeSecurityGroupIngressInput{
		GroupId: aws.String(groupId),
		IpPermissions: []*ec2.IpPermission{
			{
				FromPort:   aws.Int64(int64(port)),
				IpProtocol: aws.String("tcp"),
				ToPort:     aws.Int64(int64(port)),
				UserIdGroupPairs: []*ec2.UserIdGroupPair{
					{
						Description: aws.String("HTTP access from loadbalancer"),
						GroupId:     aws.String(loadbalancerId),
					},
				},
			},
			{
				FromPort:   aws.Int64(22),
				IpProtocol: aws.String("tcp"),
				ToPort:     aws.Int64(22),
				IpRanges: []*ec2.IpRange{
					{
						CidrIp:      aws.String("0.0.0.0/0"),
						Description: aws.String("SSH access from everywhere"),
					},
				},
			},
		},
	}

	_, err = a.ec2.AuthorizeSecurityGroupIngress(input)
	if err != nil {
		err = createError{
			text: fmt.Sprintf("Could not add base authorization to security group %s.", groupId),
			err:  err,
		}
		return
	}

	return nil
}

func (a AWS) AuthorizeHazelcast(groupId string) (err error) {
	err = a.hasEC2Session()
	if err != nil {
		return
	}
	input := &ec2.AuthorizeSecurityGroupIngressInput{
		GroupId: aws.String(groupId),
		IpPermissions: []*ec2.IpPermission{
			{
				FromPort:   aws.Int64(5701),
				IpProtocol: aws.String("tcp"),
				ToPort:     aws.Int64(5799),
				UserIdGroupPairs: []*ec2.UserIdGroupPair{
					{
						Description: aws.String("Hazelcast access"),
						GroupId:     aws.String(groupId),
					},
				},
			},
		},
	}

	_, err = a.ec2.AuthorizeSecurityGroupIngress(input)
	if err != nil {
		err = createError{
			text: fmt.Sprintf("Could not add Hazelcast authorization to security group %s.", groupId),
			err:  err,
		}
		return
	}

	return nil
}

func (a AWS) RegisterTarget(targetGroupArn, serverId string) (err error) {
	err = a.hasELBSession()
	if err != nil {
		return
	}
	input := &elbv2.RegisterTargetsInput{
		TargetGroupArn: aws.String(targetGroupArn),
		Targets: []*elbv2.TargetDescription{
			{
				Id: aws.String(serverId),
			},
		},
	}

	_, err = a.elb.RegisterTargets(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case elbv2.ErrCodeTargetGroupNotFoundException:
				err = createError{
					text: "Target group not found",
					err:  aerr,
				}
			case elbv2.ErrCodeTooManyTargetsException:
				err = createError{
					text: "Too many targets",
					err:  aerr,
				}
			case elbv2.ErrCodeInvalidTargetException:
				err = createError{
					text: "Invalid target",
					err:  aerr,
				}
			case elbv2.ErrCodeTooManyRegistrationsForTargetIdException:
				err = createError{
					text: "To many registrations for target id",
					err:  aerr,
				}
			}
		}
		err = createError{
			text: fmt.Sprintf("Unable to target for server %s.", serverId),
			err:  err,
		}
		return
	}
	return
}
