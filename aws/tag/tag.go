package tag

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	elbv2types "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	//"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	elbv2 "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/cantara/nerthus/aws/util"
)

type Tag interface {
	Create() (string, error)
	Delete() error
}

type tag struct {
	ec2Resources []string
	elbResources []string
	Key          string
	Value        string
	ec2          *ec2.Client
	elb          *elbv2.Client
	created      bool
}

func (t *tag) Create() (id string, err error) {
	if t.ec2Resources != nil {
		err = util.CheckEC2Session(t.ec2)
		if err != nil {
			return
		}
		_, err = t.ec2.CreateTags(context.Background(), &ec2.CreateTagsInput{
			Resources: t.ec2Resources,
			Tags: []ec2types.Tag{
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
			_, err = t.elb.AddTags(context.Background(), &elbv2.AddTagsInput{
				ResourceArns: []string{resource},
				Tags: []elbv2types.Tag{
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
		_, err = t.ec2.DeleteTags(context.Background(), &ec2.DeleteTagsInput{
			Resources: t.ec2Resources,
			Tags: []ec2types.Tag{
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
		_, err = t.elb.RemoveTags(context.Background(), &elbv2.RemoveTagsInput{
			ResourceArns: t.elbResources,
			TagKeys: []string{
				t.Key,
			},
		})

		if err != nil {
			return
		}
	}
	return
}
