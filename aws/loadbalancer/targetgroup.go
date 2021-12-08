package loadbalancer

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/cantara/nerthus/aws/util"
	"github.com/cantara/nerthus/aws/vpc"
)

type TargetGroup struct {
	Service string
	Name    string
	UriPath string
	Port    int
	ARN     string
	vpc     vpc.VPC
	elb     *elbv2.ELBV2
}

func NewTargetGroup(service, uriPath string, port int, vpc vpc.VPC, elb *elbv2.ELBV2) (tg TargetGroup, err error) {
	err = util.CheckELBV2Session(elb)
	if err != nil {
		return
	}
	tg = TargetGroup{
		Service: service,
		Name:    service + "-tg",
		UriPath: uriPath,
		Port:    port,
		vpc:     vpc,
		elb:     elb,
	}
	return
}

func (tg *TargetGroup) Create() (id string, err error) {
	err = util.CheckELBV2Session(tg.elb)
	if err != nil {
		return
	}
	input := &elbv2.CreateTargetGroupInput{
		Name:                       aws.String(tg.Name),
		Port:                       aws.Int64(int64(tg.Port)),
		Protocol:                   aws.String("HTTP"),
		VpcId:                      aws.String(tg.vpc.Id),
		TargetType:                 aws.String("instance"),
		ProtocolVersion:            aws.String("HTTP1"),
		HealthCheckIntervalSeconds: aws.Int64(5),
		HealthCheckPath:            aws.String(fmt.Sprintf("/%s/health", tg.UriPath)), //FIXME: This is shady
		HealthCheckPort:            aws.String("traffic-port"),
		HealthCheckProtocol:        aws.String("HTTP"),
		HealthCheckTimeoutSeconds:  aws.Int64(2),
		HealthyThresholdCount:      aws.Int64(2),
	}

	result, err := tg.elb.CreateTargetGroup(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case elbv2.ErrCodeDuplicateTargetGroupNameException:
				err = util.CreateError{
					Text: "Duplicate target group name.",
					Err:  aerr,
				}
			case elbv2.ErrCodeTooManyTargetGroupsException:
				err = util.CreateError{
					Text: "Too many target groups",
					Err:  aerr,
				}
			case elbv2.ErrCodeInvalidConfigurationRequestException:
				err = util.CreateError{
					Text: "Invalid configuration",
					Err:  aerr,
				}
			case elbv2.ErrCodeTooManyTagsException:
				err = util.CreateError{
					Text: "To many tags",
					Err:  aerr,
				}
			}
		}
		err = util.CreateError{
			Text: fmt.Sprintf("Unable to create target group for service %s.", tg.Service),
			Err:  err,
		}
		return
	}
	tg.ARN = aws.StringValue(result.TargetGroups[0].TargetGroupArn)
	id = tg.ARN
	return
}

func (tg *TargetGroup) Delete() (err error) {
	err = util.CheckELBV2Session(tg.elb)
	if err != nil {
		return
	}
	input := &elbv2.DeleteTargetGroupInput{
		TargetGroupArn: aws.String(tg.ARN),
	}

	_, err = tg.elb.DeleteTargetGroup(input)
	return
}
