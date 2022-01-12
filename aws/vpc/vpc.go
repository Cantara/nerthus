package vpc

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/cantara/nerthus/aws/util"
)

type VPC struct {
	Id string `json:"id"`
}

func GetVPC(e2 *ec2.EC2) (vpc VPC, err error) {
	err = util.CheckEC2Session(e2)
	if err != nil {
		return
	}
	result, err := e2.DescribeVpcs(nil)
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
	vpc = VPC{
		Id: aws.StringValue(result.Vpcs[0].VpcId),
	}
	return
}
