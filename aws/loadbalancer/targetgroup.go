package loadbalancer

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	//"github.com/aws/aws-sdk-go-v2/aws/awserr"
	elbv2 "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
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
	elb     *elbv2.Client
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

func NewTargetGroup(scope, name, uriPath string, port int, vpc vpc.VPC, elb *elbv2.Client) (tg TargetGroup, err error) {
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

func GetTargetGroup(scope, name, uriPath string, port int, elb *elbv2.Client) (tg TargetGroup, err error) {
	err = util.CheckELBV2Session(elb)
	if err != nil {
		return
	}
	name, err = createTargetGroupName(scope, name)
	if err != nil {
		return
	}
	result, err := elb.DescribeTargetGroups(context.Background(), &elbv2.DescribeTargetGroupsInput{
		Names: []string{
			name,
		},
	})
	if err != nil {
		return
	}

	tg = TargetGroup{
		Scope:   scope,
		Name:    name,
		UriPath: uriPath,
		Port:    port,
		elb:     elb,
	}
	tg.ARN = aws.ToString(result.TargetGroups[0].TargetGroupArn)
	return
}

func (tg *TargetGroup) Create() (id string, err error) {
	err = util.CheckELBV2Session(tg.elb)
	if err != nil {
		return
	}
	input := &elbv2.CreateTargetGroupInput{
		Name:                       aws.String(tg.Name),
		Port:                       aws.Int32(int32(tg.Port)),
		Protocol:                   "HTTP",
		VpcId:                      aws.String(tg.vpc.Id),
		TargetType:                 "instance",
		ProtocolVersion:            aws.String("HTTP1"),
		HealthCheckIntervalSeconds: aws.Int32(5),
		HealthCheckPath:            aws.String(fmt.Sprintf("/%s/health", tg.UriPath)), //FIXME: This is shady
		HealthCheckPort:            aws.String("traffic-port"),
		HealthCheckProtocol:        "HTTP",
		HealthCheckTimeoutSeconds:  aws.Int32(2),
		HealthyThresholdCount:      aws.Int32(2),
	}

	result, err := tg.elb.CreateTargetGroup(context.Background(), input)
	if err != nil {
		return
	}
	tg.ARN = aws.ToString(result.TargetGroups[0].TargetGroupArn)
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

	_, err = tg.elb.DeleteTargetGroup(context.Background(), input)
	return
}

func (tg TargetGroup) WithELB(e *elbv2.Client) TargetGroup {
	tg.elb = e
	return tg
}
