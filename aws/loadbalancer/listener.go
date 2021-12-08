package loadbalancer

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/elbv2"
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
