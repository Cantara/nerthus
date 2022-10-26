package metadata

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/cantara/nerthus/aws/util"
)

type VPC struct {
	Id string `json:"id"`
}

func GetAllServerIds(e2 *ec2.Client) (ids []string, err error) {
	first := true
	var nextToken *string
	for nextToken != nil || first {
		first = false
		result, err := e2.DescribeInstances(context.Background(), &ec2.DescribeInstancesInput{
			NextToken: nextToken,
		})
		if err != nil {
			return nil, err
		}
		nextToken = result.NextToken
		for _, reservation := range result.Reservations {
			for _, instance := range reservation.Instances {
				ids = append(ids, aws.ToString(instance.InstanceId))
			}
		}
	}
	return
}

func GetAllServersWithMetadataV1IDs(e2 *ec2.Client) (ids []string, err error) {
	first := true
	var nextToken *string
	for nextToken != nil || first {
		first = false
		result, err := e2.DescribeInstances(context.Background(), &ec2.DescribeInstancesInput{
			NextToken: nextToken,
		})
		if err != nil {
			return nil, err
		}
		nextToken = result.NextToken
		for _, reservation := range result.Reservations {
			for _, instance := range reservation.Instances {
				if instance.MetadataOptions.HttpTokens == ec2types.HttpTokensStateRequired {
					continue
				}
				ids = append(ids, aws.ToString(instance.InstanceId))
			}
		}
	}
	return
}

func SetMetadataV2(id string, e2 *ec2.Client) (err error) {
	err = util.CheckEC2Session(e2)
	if err != nil {
		return
	}
	_, err = e2.ModifyInstanceMetadataOptions(context.Background(), &ec2.ModifyInstanceMetadataOptionsInput{
		InstanceId: &id,
		HttpTokens: ec2types.HttpTokensStateRequired,
	})
	if err != nil {
		return
	}
	return
}
