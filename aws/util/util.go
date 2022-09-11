package util

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	elbv2 "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/rds"
)

type AWSObject interface {
	Create() (string, error)
	Delete() error
}

func CheckEC2Session(e2 *ec2.Client) error {
	if e2 == nil {
		return fmt.Errorf("No ec2 session found")
	}
	return nil
}

func CheckELBV2Session(elb *elbv2.Client) error {
	if elb == nil {
		return fmt.Errorf("No elbv2 session found")
	}
	return nil
}

func CheckRDSSession(db *rds.Client) error {
	if db == nil {
		return fmt.Errorf("No rds session found")
	}
	return nil
}

type CreateError struct {
	Text string
	Err  error
}

func (e CreateError) Error() string {
	return fmt.Sprintf("%s: %v", e.Text, e.Err)
}

func (e CreateError) Unwrap() error {
	return e.Err
}
