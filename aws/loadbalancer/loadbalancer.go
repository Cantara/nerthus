package loadbalancer

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/elbv2"
)

type Loadbalancer struct {
	ARN           string   `json:"arn"`
	SecurityGroup string   `json:"security_group"`
	DNSName       string   `json:"dns_name"`
	ListenerARN   string   `json:"listener_arn"`
	Paths         []string `json:"paths"`
}

func GetLoadbalancers(svc *elbv2.ELBV2) (loadbalancers []Loadbalancer, err error) {
	input := &elbv2.DescribeLoadBalancersInput{}

	result, err := svc.DescribeLoadBalancers(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case elbv2.ErrCodeLoadBalancerNotFoundException:
				fmt.Println(elbv2.ErrCodeLoadBalancerNotFoundException, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return
	}

	for _, loadbalancer := range result.LoadBalancers {
		if aws.StringValue(loadbalancer.Type) != "application" {
			continue
		}
		if aws.StringValue(loadbalancer.Scheme) != "internet-facing" {
			//We could continue here if we dont want internal loadbalancers
			continue
		}
		input2 := &elbv2.DescribeListenersInput{
			LoadBalancerArn: loadbalancer.LoadBalancerArn,
		}

		result2, err := svc.DescribeListeners(input2)
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				switch aerr.Code() {
				case elbv2.ErrCodeListenerNotFoundException:
					fmt.Println(elbv2.ErrCodeListenerNotFoundException, aerr.Error())
				case elbv2.ErrCodeLoadBalancerNotFoundException:
					fmt.Println(elbv2.ErrCodeLoadBalancerNotFoundException, aerr.Error())
				case elbv2.ErrCodeUnsupportedProtocolException:
					fmt.Println(elbv2.ErrCodeUnsupportedProtocolException, aerr.Error())
				default:
					fmt.Println(aerr.Error())
				}
			} else {
				// Print the error, cast err to awserr.Error to get the Code and
				// Message from an error.
				fmt.Println(err.Error())
			}
			return nil, err
		}
		for _, listener := range result2.Listeners {
			if *listener.Port != 443 {
				continue
			}
			if aws.StringValue(listener.Protocol) != "HTTPS" {
				fmt.Println(aws.StringValue(listener.Protocol))
				continue
			}
			input3 := &elbv2.DescribeRulesInput{
				ListenerArn: listener.ListenerArn,
			}

			result3, err := svc.DescribeRules(input3)
			if err != nil {
				if aerr, ok := err.(awserr.Error); ok {
					switch aerr.Code() {
					case elbv2.ErrCodeListenerNotFoundException:
						fmt.Println(elbv2.ErrCodeListenerNotFoundException, aerr.Error())
					case elbv2.ErrCodeRuleNotFoundException:
						fmt.Println(elbv2.ErrCodeRuleNotFoundException, aerr.Error())
					case elbv2.ErrCodeUnsupportedProtocolException:
						fmt.Println(elbv2.ErrCodeUnsupportedProtocolException, aerr.Error())
					default:
						fmt.Println(aerr.Error())
					}
				} else {
					// Print the error, cast err to awserr.Error to get the Code and
					// Message from an error.
					fmt.Println(err.Error())
				}
				return nil, err
			}
			paths := []string{}
			for _, rule := range result3.Rules {
				for _, condition := range rule.Conditions {
					if aws.StringValue(condition.Field) != "path-pattern" {
						continue
					}
					for _, path := range condition.PathPatternConfig.Values {
						paths = append(paths, aws.StringValue(path))
					}
				}
			}
			loadbalancers = append(loadbalancers, Loadbalancer{
				ARN:           aws.StringValue(loadbalancer.LoadBalancerArn),
				SecurityGroup: aws.StringValue(loadbalancer.SecurityGroups[0]),
				DNSName:       aws.StringValue(loadbalancer.DNSName),
				ListenerARN:   aws.StringValue(listener.ListenerArn),
				Paths:         paths,
			})
			break
		}
	}
	return
}
