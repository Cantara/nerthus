package tag

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/cantara/nerthus/aws/util"
)

type Tag interface {
	Create() (string, error)
	Delete() error
}

type tag struct {
	ec2Resources []*string
	elbResources []*string
	Key          string
	Value        string
	ec2          *ec2.EC2
	elb          *elbv2.ELBV2
	created      bool
}

func (t *tag) Create() (id string, err error) {
	if t.ec2Resources != nil {
		err = util.CheckEC2Session(t.ec2)
		if err != nil {
			return
		}
		_, err = t.ec2.CreateTags(&ec2.CreateTagsInput{
			Resources: t.ec2Resources,
			Tags: []*ec2.Tag{
				{
					Key:   aws.String(t.Key),
					Value: aws.String(t.Value),
				},
			},
		})
		if err != nil {
			return
		}
	}
	if t.elbResources != nil {
		err = util.CheckELBV2Session(t.elb)
		if err != nil {
			return
		}
		for _, resource := range t.elbResources {
			_, err = t.elb.AddTags(&elbv2.AddTagsInput{
				ResourceArns: []*string{resource},
				Tags: []*elbv2.Tag{
					{
						Key:   aws.String(t.Key),
						Value: aws.String(t.Value),
					},
				},
			})
			if err != nil {
				if aerr, ok := err.(awserr.Error); ok {
					switch aerr.Code() {
					case elbv2.ErrCodeDuplicateTagKeysException:
						fmt.Println(elbv2.ErrCodeDuplicateTagKeysException, aerr.Error())
						err = util.CreateError{
							Text: "Duplicate tag key",
							Err:  aerr,
						}
					case elbv2.ErrCodeTooManyTagsException:
						fmt.Println(elbv2.ErrCodeTooManyTagsException, aerr.Error())
						err = util.CreateError{
							Text: "Too many tags",
							Err:  aerr,
						}
					case elbv2.ErrCodeLoadBalancerNotFoundException:
						fmt.Println(elbv2.ErrCodeLoadBalancerNotFoundException, aerr.Error())
						err = util.CreateError{
							Text: "Loadbalancer not found",
							Err:  aerr,
						}
					case elbv2.ErrCodeTargetGroupNotFoundException:
						fmt.Println(elbv2.ErrCodeTargetGroupNotFoundException, aerr.Error())
						err = util.CreateError{
							Text: "Target group not found",
							Err:  aerr,
						}
					case elbv2.ErrCodeListenerNotFoundException:
						fmt.Println(elbv2.ErrCodeListenerNotFoundException, aerr.Error())
						err = util.CreateError{
							Text: "Listener not found",
							Err:  aerr,
						}
					case elbv2.ErrCodeRuleNotFoundException:
						fmt.Println(elbv2.ErrCodeRuleNotFoundException, aerr.Error())
						err = util.CreateError{
							Text: "Rule not found",
							Err:  aerr,
						}
					}
				}
				return
			}
		}
	}
	t.created = true
	return
}

func (t *tag) Delete() (err error) {
	if !t.created {
		return
	}
	if t.ec2Resources != nil {
		err = util.CheckEC2Session(t.ec2)
		if err != nil {
			return
		}
		_, err = t.ec2.DeleteTags(&ec2.DeleteTagsInput{
			Resources: t.ec2Resources,
			Tags: []*ec2.Tag{
				{
					Key:   aws.String(t.Key),
					Value: aws.String(t.Value),
				},
			},
		})
		if err != nil {
			return
		}
	}
	if t.elbResources != nil {
		err = util.CheckELBV2Session(t.elb)
		if err != nil {
			return
		}
		_, err = t.elb.RemoveTags(&elbv2.RemoveTagsInput{
			ResourceArns: t.elbResources,
			TagKeys: []*string{
				aws.String(t.Key),
			},
		})

		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				switch aerr.Code() {
				case elbv2.ErrCodeDuplicateTagKeysException:
					fmt.Println(elbv2.ErrCodeDuplicateTagKeysException, aerr.Error())
					err = util.CreateError{
						Text: "Duplicate tag key",
						Err:  aerr,
					}
				case elbv2.ErrCodeTooManyTagsException:
					fmt.Println(elbv2.ErrCodeTooManyTagsException, aerr.Error())
					err = util.CreateError{
						Text: "Too many tags",
						Err:  aerr,
					}
				case elbv2.ErrCodeLoadBalancerNotFoundException:
					fmt.Println(elbv2.ErrCodeLoadBalancerNotFoundException, aerr.Error())
					err = util.CreateError{
						Text: "Loadbalancer not found",
						Err:  aerr,
					}
				case elbv2.ErrCodeTargetGroupNotFoundException:
					fmt.Println(elbv2.ErrCodeTargetGroupNotFoundException, aerr.Error())
					err = util.CreateError{
						Text: "Target group not found",
						Err:  aerr,
					}
				case elbv2.ErrCodeListenerNotFoundException:
					fmt.Println(elbv2.ErrCodeListenerNotFoundException, aerr.Error())
					err = util.CreateError{
						Text: "Listener not found",
						Err:  aerr,
					}
				case elbv2.ErrCodeRuleNotFoundException:
					fmt.Println(elbv2.ErrCodeRuleNotFoundException, aerr.Error())
					err = util.CreateError{
						Text: "Rule not found",
						Err:  aerr,
					}
				}
			}
			return
		}
	}
	return
}
