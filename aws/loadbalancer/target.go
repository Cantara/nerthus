package loadbalancer

import (
	"context"
	elbv2types "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"

	"github.com/aws/aws-sdk-go-v2/aws"
	//"github.com/aws/aws-sdk-go-v2/aws/awserr"
	elbv2 "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/cantara/nerthus/aws/server"
	"github.com/cantara/nerthus/aws/util"
)

type Target struct {
	targetGroup TargetGroup
	server      server.Server
	elb         *elbv2.Client
	created     bool
}

func NewTarget(tg TargetGroup, s server.Server, elb *elbv2.Client) (t Target, err error) {
	err = util.CheckELBV2Session(elb)
	if err != nil {
		return
	}
	t = Target{
		targetGroup: tg,
		server:      s,
		elb:         elb,
	}
	return
}

func (t *Target) Create() (id string, err error) {
	err = util.CheckELBV2Session(t.elb)
	if err != nil {
		return
	}
	input := &elbv2.RegisterTargetsInput{
		TargetGroupArn: aws.String(t.targetGroup.ARN),
		Targets: []elbv2types.TargetDescription{
			{
				Id: aws.String(t.server.Id),
			},
		},
	}

	_, err = t.elb.RegisterTargets(context.Background(), input)
	if err != nil {
		return
	}
	t.created = true
	return
}

func (t *Target) Delete() (err error) {
	if !t.created {
		return
	}
	err = util.CheckELBV2Session(t.elb)
	if err != nil {
		return
	}
	input := &elbv2.DeregisterTargetsInput{
		TargetGroupArn: aws.String(t.targetGroup.ARN),
		Targets: []elbv2types.TargetDescription{
			{
				Id: aws.String(t.server.Id),
			},
		},
	}

	_, err = t.elb.DeregisterTargets(context.Background(), input)
	return
}
