package aws

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func CheckIfServiceExcistsInScope(scope, service string, e *ec2.Client) (exists bool, err error) {
	result, err := e.DescribeTags(context.Background(), &ec2.DescribeTagsInput{
		Filters: []ec2types.Filter{
			{
				Name: aws.String("tag:" + service),
				Values: []string{
					scope,
				},
			},
		},
	})
	if err != nil {
		return
	}
	exists = len(result.Tags) > 0
	return
}
