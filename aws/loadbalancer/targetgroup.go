package loadbalancer

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/cantara/nerthus/aws/util"
	"github.com/cantara/nerthus/aws/vpc"
)

type TargetGroup struct {
	Scope   string `json:"-"`
	Name    string `json:"name"`
	UriPath string `json:"path"`
	Port    int    `json:"port"`
	ARN     string `json:"arn"`
	vpc     vpc.VPC
	elb     *elbv2.ELBV2
	created bool
}

func createTargetGroupName(scope, name string) (string, error) {
	tgName := strings.Split(scope, "-")[0] + "-" + strings.ToLower(name)
	tgName = strings.TrimSuffix(tgName, "api")
	tgName = strings.ReplaceAll(tgName, "-", " ")
	tgName = strings.TrimSpace(tgName)
	tgName = strings.ReplaceAll(tgName, " ", "-")
	tgName = tgName + "-tg"
	if len(tgName) > 32 {
		return "", fmt.Errorf("Calculated targetgroup name (%s) is to long based on input scope (%s) and name (%s). Max len 32.",
			tgName, scope, name)
	}
	return tgName, nil
}

func NewTargetGroup(scope, name, uriPath string, port int, vpc vpc.VPC, elb *elbv2.ELBV2) (tg TargetGroup, err error) {
	err = util.CheckELBV2Session(elb)
	if err != nil {
		return
	}
	name, err = createTargetGroupName(scope, name)
	if err != nil {
		return
	}
	tg = TargetGroup{
		Scope:   scope,
		Name:    name,
		UriPath: uriPath,
		Port:    port,
		vpc:     vpc,
		elb:     elb,
	}
	return
}

func GetTargetGroup(scope, name, uriPath string, port int, elb *elbv2.ELBV2) (tg TargetGroup, err error) {
	err = util.CheckELBV2Session(elb)
	if err != nil {
		return
	}
	name, err = createTargetGroupName(scope, name)
	if err != nil {
		return
	}
	result, err := elb.DescribeTargetGroups(&elbv2.DescribeTargetGroupsInput{
		Names: []*string{
			aws.String(name),
		},
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case elbv2.ErrCodeLoadBalancerNotFoundException:
				err = util.CreateError{
					Text: "Loadbalancer not found.",
					Err:  aerr,
				}
			case elbv2.ErrCodeTargetGroupNotFoundException:
				err = util.CreateError{
					Text: "Targetgroup not found.",
					Err:  aerr,
				}
			}
		}
		return
	}

	tg = TargetGroup{
		Scope:   scope,
		Name:    name,
		UriPath: uriPath,
		Port:    port,
		elb:     elb,
	}
	tg.ARN = aws.StringValue(result.TargetGroups[0].TargetGroupArn)
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
			Text: fmt.Sprintf("Unable to create target group for scope %s.", tg.Scope),
			Err:  err,
		}
		return
	}
	tg.ARN = aws.StringValue(result.TargetGroups[0].TargetGroupArn)
	id = tg.ARN
	tg.created = true
	return
}

func (tg *TargetGroup) Delete() (err error) {
	if !tg.created {
		return
	}
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

func (tg TargetGroup) WithELB(e *elbv2.ELBV2) TargetGroup {
	tg.elb = e
	return tg
}
