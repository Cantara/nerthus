package aws

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elbv2"
)

func (a AWS) DeleteKeyPair(vals ...string) (err error) {
	if len(vals) != 1 {
		return errors.New("Wrong number of params, need keyName")
	}
	err = a.hasEC2Session()
	if err != nil {
		return
	}
	input := &ec2.DeleteKeyPairInput{
		KeyName: aws.String(vals[0]),
	}

	_, err = a.ec2.DeleteKeyPair(input)
	return
}

func (a AWS) DeleteSecurityGroup(vals ...string) (err error) {
	if len(vals) != 1 {
		return errors.New("Wrong number of params, need securityGroupId")
	}
	err = a.hasEC2Session()
	if err != nil {
		return
	}
	input := &ec2.DeleteSecurityGroupInput{
		GroupId: aws.String(vals[0]),
	}

	_, err = a.ec2.DeleteSecurityGroup(input)
	return
}

func (a AWS) DeleteServer(vals ...string) (err error) {
	if len(vals) != 1 {
		return errors.New("Wrong number of params, need instanceId")
	}
	err = a.hasEC2Session()
	if err != nil {
		return
	}
	input := &ec2.TerminateInstancesInput{
		InstanceIds: []*string{
			aws.String(vals[0]),
		},
	}

	_, err = a.ec2.TerminateInstances(input)
	if err != nil {
		return
	}
	a.WaitForTerminated(vals[0])
	return
}

func (a AWS) DeleteTargetGroup(vals ...string) (err error) {
	if len(vals) != 1 {
		return errors.New("Wrong number of params, need targetGroupArn")
	}
	err = a.hasELBSession()
	if err != nil {
		return
	}
	input := &elbv2.DeleteTargetGroupInput{
		TargetGroupArn: aws.String(vals[0]),
	}

	_, err = a.elb.DeleteTargetGroup(input)
	return
}

func (a AWS) DeleteRule(vals ...string) (err error) {
	if len(vals) != 1 {
		return errors.New("Wrong number of params, need ruleArn")
	}
	err = a.hasELBSession()
	if err != nil {
		return
	}
	input := &elbv2.DeleteRuleInput{
		RuleArn: aws.String(vals[0]),
	}

	_, err = a.elb.DeleteRule(input)
	return
}

func (a AWS) RemoveTargetGroupTarget(vals ...string) (err error) {
	if len(vals) != 2 {
		return errors.New("Wrong number of params, need targetGroupArn, serverId")
	}
	err = a.hasELBSession()
	if err != nil {
		return
	}
	input := &elbv2.DeregisterTargetsInput{
		TargetGroupArn: aws.String(vals[0]),
		Targets: []*elbv2.TargetDescription{
			{
				Id: aws.String(vals[1]),
			},
		},
	}

	_, err = a.elb.DeregisterTargets(input)
	return
}
