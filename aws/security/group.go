package security

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/cantara/nerthus/aws/util"
	"github.com/cantara/nerthus/aws/vpc"
)

type Group struct {
	Scope   string `json:"-"`
	Name    string `json:"name"`
	Desc    string `json:"-"`
	Id      string `json:"id"`
	vpc     vpc.VPC
	ec2     *ec2.EC2
	created bool
}

func NewGroup(scope string, vpc vpc.VPC, e2 *ec2.EC2) (g Group, err error) {
	err = util.CheckEC2Session(e2)
	if err != nil {
		return
	}
	g = Group{
		Scope: scope,
		Name:  scope + "-sg",
		Desc:  "Security group for scope: " + g.Scope,
		vpc:   vpc,
		ec2:   e2,
	}
	return
}

func NewDBGroup(serviceName, scope string, vpc vpc.VPC, e2 *ec2.EC2) (g Group, err error) {
	err = util.CheckEC2Session(e2)
	if err != nil {
		return
	}
	g = Group{
		Scope: scope,
		Name:  fmt.Sprintf("%s-%s-db", g.Scope, serviceName),
		Desc:  "Database security group for scope: " + g.Scope + " " + serviceName,
		vpc:   vpc,
		ec2:   e2,
	}
	return
}

func (g *Group) Create() (groupId string, err error) {
	err = util.CheckEC2Session(g.ec2)
	if err != nil {
		return
	}
	secGroupRes, err := g.ec2.CreateSecurityGroup(&ec2.CreateSecurityGroupInput{
		GroupName:   aws.String(g.Name),
		Description: aws.String(g.Desc),
		VpcId:       aws.String(g.vpc.Id),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case "InvalidVpcID.NotFound":
				err = util.CreateError{
					Text: fmt.Sprintf("Unable to find VPC with ID %q.", g.vpc.Id),
					Err:  err,
				}
				return
			case "InvalidGroup.Duplicate":
				err = util.CreateError{
					Text: fmt.Sprintf("Security group %q already exists.", g.Name),
					Err:  err,
				}
				return
			}
		}
		err = util.CreateError{
			Text: fmt.Sprintf("Unable to create security group %q.", g.Name),
			Err:  err,
		}
		return
	}
	g.Id = aws.StringValue(secGroupRes.GroupId)
	groupId = g.Id

	// Add tags to the created security group
	_, err = g.ec2.CreateTags(&ec2.CreateTagsInput{
		Resources: []*string{aws.String(groupId)},
		Tags: []*ec2.Tag{
			{
				Key:   aws.String("Name"),
				Value: aws.String(g.Name),
			},
			{
				Key:   aws.String("Scope"),
				Value: aws.String(g.Scope),
			},
		},
	})
	if err != nil {
		err = util.CreateError{
			Text: fmt.Sprintf("Could not create tags for sg %s %s.", g.Id, g.Name),
			Err:  err,
		}
		return
	}
	g.created = true
	return groupId, nil
}

func (g Group) Wait() (err error) {
	if !g.created {
		return
	}
	err = util.CheckEC2Session(g.ec2)
	if err != nil {
		return
	}
	err = g.ec2.WaitUntilSecurityGroupExists(&ec2.DescribeSecurityGroupsInput{
		GroupIds: []*string{
			aws.String(g.Id),
		},
	})
	return
}

func (g *Group) Delete() (err error) {
	if !g.created {
		return
	}
	err = util.CheckEC2Session(g.ec2)
	if err != nil {
		return
	}
	_, err = g.ec2.DeleteSecurityGroup(&ec2.DeleteSecurityGroupInput{
		GroupId: aws.String(g.Id),
	})
	return
}

func (g Group) AddBaseAuthorization() (err error) {
	err = util.CheckEC2Session(g.ec2)
	if err != nil {
		return
	}
	input := &ec2.AuthorizeSecurityGroupIngressInput{
		GroupId: aws.String(g.Id),
		IpPermissions: []*ec2.IpPermission{
			{
				FromPort:   aws.Int64(22),
				IpProtocol: aws.String("tcp"),
				ToPort:     aws.Int64(22),
				IpRanges: []*ec2.IpRange{
					{
						CidrIp:      aws.String("0.0.0.0/0"),
						Description: aws.String("SSH access from everywhere"),
					},
				},
			},
		},
	}

	_, err = g.ec2.AuthorizeSecurityGroupIngress(input)
	if err != nil {
		err = util.CreateError{
			Text: fmt.Sprintf("Could not add base authorization to security group %s %s.", g.Id, g.Name),
			Err:  err,
		}
		return
	}

	return
}

func (g Group) AddDatabaseAuthorization(serverSgId string) (err error) {
	err = util.CheckEC2Session(g.ec2)
	if err != nil {
		return
	}
	input := &ec2.AuthorizeSecurityGroupIngressInput{
		GroupId: aws.String(g.Id),
		IpPermissions: []*ec2.IpPermission{
			{
				FromPort:   aws.Int64(5432),
				IpProtocol: aws.String("tcp"),
				ToPort:     aws.Int64(5432),
				UserIdGroupPairs: []*ec2.UserIdGroupPair{
					{
						Description: aws.String("Postgresql access from server"),
						GroupId:     aws.String(serverSgId),
					},
				},
			},
		},
	}

	_, err = g.ec2.AuthorizeSecurityGroupIngress(input)
	if err != nil {
		err = util.CreateError{
			Text: fmt.Sprintf("Could not add base authorization to security group %s %s.", g.Id, g.Name),
			Err:  err,
		}
		return
	}

	return
}

func (g Group) AddLoadbalancerAuthorization(loadbalancerId string, port int) (err error) {
	err = util.CheckEC2Session(g.ec2)
	if err != nil {
		return
	}
	input := &ec2.AuthorizeSecurityGroupIngressInput{
		GroupId: aws.String(g.Id),
		IpPermissions: []*ec2.IpPermission{
			{
				FromPort:   aws.Int64(int64(port)),
				IpProtocol: aws.String("tcp"),
				ToPort:     aws.Int64(int64(port)),
				UserIdGroupPairs: []*ec2.UserIdGroupPair{
					{
						Description: aws.String("HTTP access from loadbalancer"),
						GroupId:     aws.String(loadbalancerId),
					},
				},
			},
		},
	}

	_, err = g.ec2.AuthorizeSecurityGroupIngress(input)
	if err != nil {
		err = util.CreateError{
			Text: fmt.Sprintf("Could not add service loadbalancer authorization to security group %s %s.", g.Id, g.Name),
			Err:  err,
		}
		return
	}

	return
}

func (g Group) AddServerAccess(sgId string) (err error) {
	err = util.CheckEC2Session(g.ec2)
	if err != nil {
		return
	}
	input := &ec2.AuthorizeSecurityGroupIngressInput{
		GroupId: aws.String(g.Id),
		IpPermissions: []*ec2.IpPermission{
			{
				FromPort:   aws.Int64(5432),
				IpProtocol: aws.String("tcp"),
				ToPort:     aws.Int64(5432),
				UserIdGroupPairs: []*ec2.UserIdGroupPair{
					{
						Description: aws.String("PSQL access from servers in scope: " + g.Scope),
						GroupId:     aws.String(sgId),
					},
				},
			},
		},
	}

	_, err = g.ec2.AuthorizeSecurityGroupIngress(input)
	if err != nil {
		err = util.CreateError{
			Text: fmt.Sprintf("Could not add PSQL access to security group %s %s.", g.Id, g.Name),
			Err:  err,
		}
		return
	}

	return
}

func (g *Group) AuthorizeHazelcast() (err error) {
	err = util.CheckEC2Session(g.ec2)
	if err != nil {
		return
	}
	input := &ec2.AuthorizeSecurityGroupIngressInput{
		GroupId: aws.String(g.Id),
		IpPermissions: []*ec2.IpPermission{
			{
				FromPort:   aws.Int64(5700),
				IpProtocol: aws.String("tcp"),
				ToPort:     aws.Int64(5799),
				UserIdGroupPairs: []*ec2.UserIdGroupPair{
					{
						Description: aws.String("Hazelcast access"),
						GroupId:     aws.String(g.Id),
					},
				},
			},
		},
	}

	_, err = g.ec2.AuthorizeSecurityGroupIngress(input)
	if err != nil {
		err = util.CreateError{
			Text: fmt.Sprintf("Could not add Hazelcast authorization to security group %s %s.", g.Id, g.Name),
			Err:  err,
		}
		return
	}

	return
}

func (g Group) WithEC2(e *ec2.EC2) Group {
	g.ec2 = e
	return g
}
