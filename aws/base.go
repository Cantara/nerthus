package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elbv2"
)

type createError struct {
	text string
	err  error
}

func (e createError) Error() string {
	return fmt.Sprintf("%s: %v", e.text, e.err)
}

func (e createError) Unwrap() error {
	return e.err
}

type AWS struct {
	ec2 *ec2.EC2
	elb *elbv2.ELBV2
}

func (a *AWS) NewEC2(c client.ConfigProvider) {
	a.ec2 = ec2.New(c)
}

func (a AWS) hasEC2Session() error {
	if a.ec2 == nil {
		return fmt.Errorf("No ec2 session found")
	}
	return nil
}

func (a *AWS) NewELB(c client.ConfigProvider) {
	a.elb = elbv2.New(c)
}

func (a AWS) hasELBSession() error {
	if a.elb == nil {
		return fmt.Errorf("No elb session found")
	}
	return nil
}
