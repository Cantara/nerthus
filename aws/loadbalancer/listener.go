package loadbalancer

import (
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/elbv2"
	log "github.com/cantara/bragi"
	"github.com/cantara/nerthus/aws/util"
)

type Listener struct {
	ARN string
	elb *elbv2.ELBV2
}

func GetListener(arn string, elb *elbv2.ELBV2) (l Listener, err error) {
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
	result, err := l.elb.DescribeListeners(&elbv2.DescribeListenersInput{
		ListenerArns: []*string{
			aws.String(l.ARN),
		},
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case elbv2.ErrCodeListenerNotFoundException:
				fmt.Println(elbv2.ErrCodeListenerNotFoundException, aerr.Error())
				err = util.CreateError{
					Text: "Listener not found",
					Err:  aerr,
				}
			case elbv2.ErrCodeLoadBalancerNotFoundException:
				fmt.Println(elbv2.ErrCodeLoadBalancerNotFoundException, aerr.Error())
				err = util.CreateError{
					Text: "Loadbalancer not found",
					Err:  aerr,
				}
			case elbv2.ErrCodeUnsupportedProtocolException:
				fmt.Println(elbv2.ErrCodeUnsupportedProtocolException, aerr.Error())
				err = util.CreateError{
					Text: "Unsupported protocol",
					Err:  aerr,
				}
			}
		}
		return
	}
	loadbalancer = aws.StringValue(result.Listeners[0].LoadBalancerArn)
	return
}

func GetListeners(loadbalancerARN string, elb *elbv2.ELBV2) (l []Listener, err error) {
	err = util.CheckELBV2Session(elb)
	if err != nil {
		return
	}
	input := &elbv2.DescribeListenersInput{
		LoadBalancerArn: aws.String(loadbalancerARN),
	}

	result, err := elb.DescribeListeners(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case elbv2.ErrCodeListenerNotFoundException:
				err = util.CreateError{
					Text: "Listener not found",
					Err:  aerr,
				}
			case elbv2.ErrCodeLoadBalancerNotFoundException:
				err = util.CreateError{
					Text: "Loadbalancer not found",
					Err:  aerr,
				}
			case elbv2.ErrCodeUnsupportedProtocolException:
				err = util.CreateError{
					Text: "Unsupported protocol",
					Err:  aerr,
				}
			}
		}
		return
	}

	for _, listener := range result.Listeners {
		if *listener.Port != 443 {
			continue
		}
		if aws.StringValue(listener.Protocol) != "HTTPS" {
			continue
		}
		l = append(l, Listener{
			ARN: aws.StringValue(listener.ListenerArn),
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

	result, err := l.elb.DescribeRules(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case elbv2.ErrCodeListenerNotFoundException:
				err = util.CreateError{
					Text: "Listener not found",
					Err:  aerr,
				}
				return
			case elbv2.ErrCodeRuleNotFoundException:
				err = util.CreateError{
					Text: "Rule not found",
					Err:  aerr,
				}
				return
			case elbv2.ErrCodeUnsupportedProtocolException:
				err = util.CreateError{
					Text: "Unsupported protocol",
					Err:  aerr,
				}
				return
			}
		}
		err = util.CreateError{
			Text: "Unable to describe rules",
			Err:  err,
		}
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

	result, err := l.elb.DescribeRules(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case elbv2.ErrCodeListenerNotFoundException:
				err = util.CreateError{
					Text: "Listener not found",
					Err:  aerr,
				}
				return
			case elbv2.ErrCodeRuleNotFoundException:
				err = util.CreateError{
					Text: "Rule not found",
					Err:  aerr,
				}
				return
			case elbv2.ErrCodeUnsupportedProtocolException:
				err = util.CreateError{
					Text: "Unsupported protocol",
					Err:  aerr,
				}
				return
			}
		}
		err = util.CreateError{
			Text: "Unable to describe rules",
			Err:  err,
		}
		return
	}

	for _, rule := range result.Rules {
		priString := aws.StringValue(rule.Priority)
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
