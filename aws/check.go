package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func CheckIfServiceExcistsInScope(scope, service string, e *ec2.EC2) (exists bool, err error) {
	result, err := e.DescribeTags(&ec2.DescribeTagsInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("tag:" + service),
				Values: []*string{
					aws.String(scope),
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
