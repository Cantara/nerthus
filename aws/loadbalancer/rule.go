package loadbalancer

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	//"github.com/aws/aws-sdk-go-v2/aws/awserr"
	elbv2 "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	elbv2types "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/cantara/nerthus/aws/util"
)

type Rule struct {
	ARN         string
	listener    Listener
	targetGroup TargetGroup
	elb         *elbv2.Client
	created     bool
}

func NewRule(l Listener, tg TargetGroup, elb *elbv2.Client) (r Rule, err error) {
	err = util.CheckELBV2Session(elb)
	if err != nil {
		return
	}
	r = Rule{
		listener:    l,
		targetGroup: tg,
		elb:         elb,
	}
	return
}

func GetRules(listenerARN string, elb *elbv2.Client) (r []Rule, err error) {
	err = util.CheckELBV2Session(elb)
	if err != nil {
		return
	}
	input := &elbv2.DescribeRulesInput{
		ListenerArn: aws.String(listenerARN),
	}

	result, err := elb.DescribeRules(context.Background(), input)
	if err != nil {
		return
	}

	paths := []string{}
	for _, rule := range result.Rules {
		for _, condition := range rule.Conditions {
			if aws.ToString(condition.Field) != "path-pattern" {
				continue
			}
			for _, path := range condition.PathPatternConfig.Values {
				paths = append(paths, path)
			}
		}
	}
	return
}

func (r *Rule) Create() (id string, err error) {
	err = util.CheckELBV2Session(r.elb)
	if err != nil {
		return
	}
	highestPriority, err := r.listener.GetHighestPriority()
	if err != nil {
		return
	}
	path := fmt.Sprintf("/%s", r.targetGroup.UriPath)
	input := &elbv2.CreateRuleInput{
		Actions: []elbv2types.Action{
			{
				TargetGroupArn: aws.String(r.targetGroup.ARN),
				Type:           "forward",
			},
		},
		Conditions: []elbv2types.RuleCondition{
			{
				Field: aws.String("path-pattern"),
				Values: []string{
					path,
					path + "/*",
				},
			},
		},
		ListenerArn: aws.String(r.listener.ARN),
		Priority:    aws.Int32(int32(highestPriority + 1)),
	}

	result, err := r.elb.CreateRule(context.Background(), input)
	if err != nil {
		return
	}
	r.ARN = aws.ToString(result.Rules[0].RuleArn)
	id = r.ARN
	r.created = true
	return
}

func (r *Rule) Delete() (err error) {
	if !r.created {
		return
	}
	err = util.CheckELBV2Session(r.elb)
	if err != nil {
		return
	}
	_, err = r.elb.DeleteRule(context.Background(), &elbv2.DeleteRuleInput{
		RuleArn: aws.String(r.ARN),
	})
	return
}
