package loadbalancer

import (
	"context"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	elbv2 "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	log "github.com/cantara/bragi"
	"github.com/cantara/nerthus/aws/util"
)

type Listener struct {
	ARN string
	elb *elbv2.Client
}

func GetListener(arn string, elb *elbv2.Client) (l Listener, err error) {
	err = util.CheckELBV2Session(elb)
	if err != nil {
		return
	}
	l = Listener{
		ARN: arn,
		elb: elb,
	}
	return
}

func (l Listener) GetLoadbalancer() (loadbalancer string, err error) {
	result, err := l.elb.DescribeListeners(context.Background(), &elbv2.DescribeListenersInput{
		ListenerArns: []string{
			l.ARN,
		},
	})
	if err != nil {
		return
	}
	loadbalancer = *result.Listeners[0].LoadBalancerArn
	return
}

func GetListeners(loadbalancerARN string, elb *elbv2.Client) (l []Listener, err error) {
	err = util.CheckELBV2Session(elb)
	if err != nil {
		return
	}
	input := &elbv2.DescribeListenersInput{
		LoadBalancerArn: aws.String(loadbalancerARN),
	}

	result, err := elb.DescribeListeners(context.Background(), input)
	if err != nil {
		return
	}

	for _, listener := range result.Listeners {
		if *listener.Port != 443 {
			continue
		}
		if listener.Protocol != "HTTPS" {
			continue
		}
		l = append(l, Listener{
			ARN: aws.ToString(listener.ListenerArn),
			elb: elb,
		})
	}
	return
}

func (l Listener) GetNumRules() (numRules int, err error) {
	err = util.CheckELBV2Session(l.elb)
	if err != nil {
		return
	}
	input := &elbv2.DescribeRulesInput{
		ListenerArn: aws.String(l.ARN),
	}

	result, err := l.elb.DescribeRules(context.Background(), input)
	if err != nil {
		return
	}

	return len(result.Rules), nil
}

func (l Listener) GetHighestPriority() (highestPri int, err error) {
	err = util.CheckELBV2Session(l.elb)
	if err != nil {
		return
	}
	input := &elbv2.DescribeRulesInput{
		ListenerArn: aws.String(l.ARN),
	}

	result, err := l.elb.DescribeRules(context.Background(), input)
	if err != nil {
		return
	}

	for _, rule := range result.Rules {
		priString := aws.ToString(rule.Priority)
		if priString == "default" {
			continue
		}
		pri, err := strconv.Atoi(priString)
		if err != nil {
			log.AddError(err).Notice("While paring priority as int")
			continue
		}
		if pri > highestPri {
			highestPri = pri
		}
	}

	return
}
