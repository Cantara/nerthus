package aws

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elbv2"
	log "github.com/cantara/bragi"
)

func (a AWS) GetVPCId() (vpcId string, err error) {
	err = a.hasEC2Session()
	if err != nil {
		return
	}
	result, err := a.ec2.DescribeVpcs(nil)
	if err != nil {
		err = createError{
			text: "Unable to describe VPCs",
			err:  err,
		}
		return
	}
	if len(result.Vpcs) == 0 {
		err = fmt.Errorf("No VPCs found to associate security group with.")
		return
	}
	return aws.StringValue(result.Vpcs[0].VpcId), nil
}

func (a AWS) WaitForRunning(instanceId string) (err error) {
	err = a.hasEC2Session()
	if err != nil {
		return
	}
	input := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{aws.String(instanceId)},
	}
	result, err := a.ec2.DescribeInstances(input)

	for count := 0; (err != nil || *result.Reservations[0].Instances[0].State.Name != "running") && count < 60; count++ {
		if err != nil {
			log.AddError(err).Warning("Getting running state of new instance")
		}
		result, err = a.ec2.DescribeInstances(input)
		time.Sleep(1 * time.Second)
	}
	return
}

func (a AWS) WaitForTerminated(instanceId string) (err error) {
	err = a.hasEC2Session()
	if err != nil {
		return
	}
	input := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{aws.String(instanceId)},
	}
	result, err := a.ec2.DescribeInstances(input)

	for count := 0; (err != nil || *result.Reservations[0].Instances[0].State.Name != "terminated") && count < 60; count++ {
		if err != nil {
			log.AddError(err).Warning("Getting state of instance")
		}
		result, err = a.ec2.DescribeInstances(input)
		time.Sleep(15 * time.Second)
	}
	return
}

func (a AWS) GetPublicDNS(instanceId string) (publicDNS string, err error) {
	err = a.hasEC2Session()
	if err != nil {
		return
	}
	input := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{aws.String(instanceId)},
	}
	result, err := a.ec2.DescribeInstances(input)
	if err != nil {
		err = createError{
			text: "Unable to describe Instance",
			err:  err,
		}
		return
	}
	return aws.StringValue(result.Reservations[0].Instances[0].PublicDnsName), nil
}

func (a AWS) GetNumRules(listenerArn string) (numRules int, err error) {
	err = a.hasELBSession()
	if err != nil {
		return
	}
	input := &elbv2.DescribeRulesInput{
		ListenerArn: aws.String(listenerArn),
	}

	result, err := a.elb.DescribeRules(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case elbv2.ErrCodeListenerNotFoundException:
				err = createError{
					text: "Listener not found",
					err:  aerr,
				}
				return
			case elbv2.ErrCodeRuleNotFoundException:
				err = createError{
					text: "Rule not found",
					err:  aerr,
				}
				return
			case elbv2.ErrCodeUnsupportedProtocolException:
				err = createError{
					text: "Unsupported protocol",
					err:  aerr,
				}
				return
			}
		}
		err = createError{
			text: "Unable to describe rules",
			err:  err,
		}
		return
	}

	return len(result.Rules), nil
}
