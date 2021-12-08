package loadbalancer

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/cantara/nerthus/aws/util"
)

type Rule struct {
	ARN         string
	listener    Listener
	targetGroup TargetGroup
	elb         *elbv2.ELBV2
}

func NewRule(l Listener, tg TargetGroup, elb *elbv2.ELBV2) (r Rule, err error) {
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

func (r *Rule) Create() (id string, err error) {
	err = util.CheckELBV2Session(r.elb)
	if err != nil {
		return
	}
	numRules, err := r.listener.GetNumRules()
	if err != nil {
		return
	}
	path := fmt.Sprintf("/%s/*", r.targetGroup.UriPath)
	input := &elbv2.CreateRuleInput{
		Actions: []*elbv2.Action{
			{
				TargetGroupArn: aws.String(r.targetGroup.ARN),
				Type:           aws.String("forward"),
			},
		},
		Conditions: []*elbv2.RuleCondition{
			{
				Field: aws.String("path-pattern"),
				Values: []*string{
					aws.String(path),
				},
			},
		},
		ListenerArn: aws.String(r.listener.ARN),
		Priority:    aws.Int64(int64(numRules)),
	}

	result, err := r.elb.CreateRule(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case elbv2.ErrCodePriorityInUseException:
				err = util.CreateError{
					Text: "Priority in use",
					Err:  aerr,
				}
				return
			case elbv2.ErrCodeTooManyTargetGroupsException:
				err = util.CreateError{
					Text: "Too many target groups",
					Err:  aerr,
				}
				return
			case elbv2.ErrCodeTooManyRulesException:
				err = util.CreateError{
					Text: "Too many rules",
					Err:  aerr,
				}
				return
			case elbv2.ErrCodeTargetGroupAssociationLimitException:
				err = util.CreateError{
					Text: "Target group association limit",
					Err:  aerr,
				}
				return
			case elbv2.ErrCodeIncompatibleProtocolsException:
				err = util.CreateError{
					Text: "Incompatible protocols",
					Err:  aerr,
				}
				return
			case elbv2.ErrCodeListenerNotFoundException:
				err = util.CreateError{
					Text: "Listener not found",
					Err:  aerr,
				}
				return
			case elbv2.ErrCodeTargetGroupNotFoundException:
				err = util.CreateError{
					Text: "Target group not found",
					Err:  aerr,
				}
				return
			case elbv2.ErrCodeInvalidConfigurationRequestException:
				err = util.CreateError{
					Text: "Invalid configuration",
					Err:  aerr,
				}
				return
			case elbv2.ErrCodeTooManyRegistrationsForTargetIdException:
				err = util.CreateError{
					Text: "Too many registrations for target id",
					Err:  aerr,
				}
				return
			case elbv2.ErrCodeTooManyTargetsException:
				err = util.CreateError{
					Text: "Too many targets",
					Err:  aerr,
				}
				return
			case elbv2.ErrCodeUnsupportedProtocolException:
				err = util.CreateError{
					Text: "Unsupported protocol",
					Err:  aerr,
				}
				return
			case elbv2.ErrCodeTooManyActionsException:
				err = util.CreateError{
					Text: "Too many actions",
					Err:  aerr,
				}
				return
			case elbv2.ErrCodeInvalidLoadBalancerActionException:
				err = util.CreateError{
					Text: "Incalid loadbalancer action",
					Err:  aerr,
				}
				return
			case elbv2.ErrCodeTooManyUniqueTargetGroupsPerLoadBalancerException:
				err = util.CreateError{
					Text: "Too many unique target groups per loadbalancer",
					Err:  aerr,
				}
				return
			case elbv2.ErrCodeTooManyTagsException:
				err = util.CreateError{
					Text: "Too many tags",
					Err:  aerr,
				}
				return
			}
		}
		err = util.CreateError{
			Text: fmt.Sprintf("Unable to add rule to listener %s with target group %s and path %s.",
				r.listener.ARN, r.targetGroup.ARN, path),
			Err: err,
		}
	}
	r.ARN = aws.StringValue(result.Rules[0].RuleArn)
	id = r.ARN
	return
}

func (r *Rule) Delete() (err error) {
	err = util.CheckELBV2Session(r.elb)
	if err != nil {
		return
	}
	_, err = r.elb.DeleteRule(&elbv2.DeleteRuleInput{
		RuleArn: aws.String(r.ARN),
	})
	return
}
