package loadbalancer

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/cantara/nerthus/aws/server"
	"github.com/cantara/nerthus/aws/util"
)

type Target struct {
	targetGroup TargetGroup
	server      server.Server
	elb         *elbv2.ELBV2
}

func NewTarget(tg TargetGroup, s server.Server, elb *elbv2.ELBV2) (t Target, err error) {
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
		Targets: []*elbv2.TargetDescription{
			{
				Id: aws.String(t.server.Id),
			},
		},
	}

	_, err = t.elb.RegisterTargets(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case elbv2.ErrCodeTargetGroupNotFoundException:
				err = util.CreateError{
					Text: "Target group not found",
					Err:  aerr,
				}
			case elbv2.ErrCodeTooManyTargetsException:
				err = util.CreateError{
					Text: "Too many targets",
					Err:  aerr,
				}
			case elbv2.ErrCodeInvalidTargetException:
				err = util.CreateError{
					Text: "Invalid target",
					Err:  aerr,
				}
			case elbv2.ErrCodeTooManyRegistrationsForTargetIdException:
				err = util.CreateError{
					Text: "To many registrations for target id",
					Err:  aerr,
				}
			}
		}
		err = util.CreateError{
			Text: fmt.Sprintf("Unable to register target for server %s in targetgroup %s.", t.server.Id, t.targetGroup.ARN),
			Err:  err,
		}
		return
	}
	return
}

func (t *Target) Delete() (err error) {
	err = util.CheckELBV2Session(t.elb)
	if err != nil {
		return
	}
	input := &elbv2.DeregisterTargetsInput{
		TargetGroupArn: aws.String(t.targetGroup.ARN),
		Targets: []*elbv2.TargetDescription{
			{
				Id: aws.String(t.server.Id),
			},
		},
	}

	_, err = t.elb.DeregisterTargets(input)
	return
}
