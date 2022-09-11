package loadbalancer

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"

	elbv2 "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	elbv2types "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/aws/smithy-go"
)

type Loadbalancer struct {
	ARN           string   `json:"arn"`
	SecurityGroup string   `json:"security_group"`
	DNSName       string   `json:"dns_name"`
	ListenerARN   string   `json:"listener_arn"`
	Paths         []string `json:"paths"`
}

func GetLoadbalancers(svc *elbv2.Client) (loadbalancers []Loadbalancer, err error) {
	input := &elbv2.DescribeLoadBalancersInput{}

	result, err := svc.DescribeLoadBalancers(context.Background(), input)
	if err != nil {
		var lbnf *elbv2types.LoadBalancerNotFoundException
		var apiErr smithy.APIError
		if errors.As(err, &lbnf) {
			code := lbnf.ErrorCode()
			message := lbnf.ErrorMessage()
			fmt.Printf("%s[%s]:%v\n", "LoadBalancerNotFoundException", code, message)
			return
		} else if errors.As(err, &apiErr) {
			code := apiErr.ErrorCode()
			message := apiErr.ErrorMessage()
			fmt.Printf("%s[%s]:%v\n", "Default", code, message)
		} else {
			fmt.Println(err.Error())
		}
		return
	}

	for _, loadbalancer := range result.LoadBalancers {
		if loadbalancer.Type != "application" {
			continue
		}
		if loadbalancer.Scheme != "internet-facing" {
			//We could continue here if we don't want internal loadbalancers
			continue
		}
		input2 := &elbv2.DescribeListenersInput{
			LoadBalancerArn: loadbalancer.LoadBalancerArn,
		}

		result2, err := svc.DescribeListeners(context.Background(), input2)
		if err != nil {
			var apiErr smithy.APIError
			if errors.As(err, &apiErr) {
				code := apiErr.ErrorCode()
				message := apiErr.ErrorMessage()
				fmt.Printf("%s[%s]:%v\n", "Default", code, message)
			} else {
				fmt.Println(err.Error())
			}
			return nil, err
		}
		for _, listener := range result2.Listeners {
			if *listener.Port != 443 {
				continue
			}
			if listener.Protocol != "HTTPS" {
				fmt.Println(listener.Protocol)
				continue
			}
			input3 := &elbv2.DescribeRulesInput{
				ListenerArn: listener.ListenerArn,
			}

			result3, err := svc.DescribeRules(context.Background(), input3)
			if err != nil {
				var apiErr smithy.APIError
				if errors.As(err, &apiErr) {
					code := apiErr.ErrorCode()
					message := apiErr.ErrorMessage()
					fmt.Printf("%s[%s]:%v\n", "Default", code, message)
				} else {
					fmt.Println(err.Error())
				}
				return nil, err
			}
			var paths []string
			for _, rule := range result3.Rules {
				for _, condition := range rule.Conditions {
					if *condition.Field != "path-pattern" {
						continue
					}
					for _, path := range condition.PathPatternConfig.Values {
						paths = append(paths, path)
					}
				}
			}
			loadbalancers = append(loadbalancers, Loadbalancer{
				ARN:           aws.ToString(loadbalancer.LoadBalancerArn),
				SecurityGroup: loadbalancer.SecurityGroups[0],
				DNSName:       aws.ToString(loadbalancer.DNSName),
				ListenerARN:   aws.ToString(listener.ListenerArn),
				Paths:         paths,
			})
			break
		}
	}
	return
}
