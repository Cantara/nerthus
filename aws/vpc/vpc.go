package vpc

import (
	"context"
	"fmt"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/cantara/nerthus/aws/util"
)

type VPC struct {
	Id string `json:"id"`
}

func GetVPC(e2 *ec2.Client) (vpc VPC, err error) {
	err = util.CheckEC2Session(e2)
	if err != nil {
		return
	}
	result, err := e2.DescribeVpcs(context.Background(), nil)
	if err != nil {
		err = util.CreateError{
			Text: "Unable to describe VPCs",
			Err:  err,
		}
		return
	}
	if len(result.Vpcs) == 0 {
		err = fmt.Errorf("No VPCs found.")
		return
	}
	var filtered []ec2types.Vpc

	for _, entry := range result.Vpcs {
		if *entry.IsDefault {
			filtered = append(filtered, entry)
			break
		}

	}
	vpc = VPC{
		Id: aws.ToString(filtered[0].VpcId),
	}
	return
}
