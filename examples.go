package main

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	log "github.com/cantara/bragi"
)

func createinstance() {
	fmt.Println("vim-go")
	// Get credentials from ~/.aws/credentials
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-west-2")},
	)

	// Create EC2 service client
	svc := ec2.New(sess)

	// Specify the details of the instance that you want to create
	runResult, err := svc.RunInstances(&ec2.RunInstancesInput{
		ImageId:          aws.String(os.Getenv("AMI")),
		InstanceType:     aws.String(os.Getenv("INSTANCE_TYPE")),
		MinCount:         aws.Int64(1),
		MaxCount:         aws.Int64(1),
		SecurityGroupIds: aws.StringSlice([]string{os.Getenv("SECURITY_GROUP")}),
		KeyName:          aws.String(os.Getenv("KEYNAME")),
		SubnetId:         aws.String(os.Getenv("SUBNET_ID")),
	})

	if err != nil {
		fmt.Println("Could not create instance", err)
		return
	}

	fmt.Println("Created instance", *runResult.Instances[0].InstanceId)

	// Add tags to the created instance
	_, errtag := svc.CreateTags(&ec2.CreateTagsInput{
		Resources: []*string{runResult.Instances[0].InstanceId},
		Tags: []*ec2.Tag{
			{
				Key:   aws.String("Name"),
				Value: aws.String("GoInstance"),
			},
		},
	})
	if errtag != nil {
		log.Println("Could not create tags for instance", runResult.Instances[0].InstanceId, errtag)
		return
	}

	fmt.Println("Successfully tagged instance")
}

func setupSecurityGroup(name, desc, vpc string, ec2client *ec2.EC2) (string, error) {
	//Create the input struct with the appropriate settings, making sure to use the aws string pointer type
	sgReq := ec2.CreateSecurityGroupInput{
		GroupName:   aws.String(name),
		Description: aws.String(desc),
		VpcId:       aws.String(vpc),
	}

	//Attempt to create the security group
	sgResp, err := ec2client.CreateSecurityGroup(&sgReq)
	if err != nil {
		return "", err
	}

	authReq := ec2.AuthorizeSecurityGroupIngressInput{
		CidrIp:     aws.String("0.0.0.0/0"),
		FromPort:   aws.Int64(9443),
		ToPort:     aws.Int64(9443),
		IpProtocol: aws.String("tcp"),
		GroupId:    sgResp.GroupId,
	}
	_, err = ec2client.AuthorizeSecurityGroupIngress(&authReq)
	if err != nil {
		return "", err
	}

	return *sgResp.GroupId, nil
}

func getSecurityGroupIds(c *ec2.EC2, config *Config, secgroups []string) []*string {

	//secgroups := make([]*string,0)
	secgroupids := make([]*string, 0)
	for i := range secgroups {
		filters := make([]*ec2.Filter, 0)

		keyname := "group-name"
		keyname2 := "vpc-id"
		filter := ec2.Filter{
			Name: &keyname, Values: []*string{&secgroups[i]}}
		filter2 := ec2.Filter{
			Name: &keyname2, Values: []*string{&config.VpcId}}
		filters = append(filters, &filter)
		filters = append(filters, &filter2)

		//fmt.Println("Filters ", filters)

		dsgi := &ec2.DescribeSecurityGroupsInput{Filters: filters}
		dsgo, err := c.DescribeSecurityGroups(dsgi)
		if err != nil {
			fmt.Println("Describe security groups failed.")
			panic(err)
		}

		for i := range dsgo.SecurityGroups {
			secgroupids = append(secgroupids, dsgo.SecurityGroups[i].GroupId)
		}

	}

	//fmt.Println("Security Groups!", secgroupids)
	return secgroupids

}

func getInstancesWithTag(ec2Client *ec2.EC2, key string, value string) ([]*ec2.Instance, error) {
	var instances []*ec2.Instance

	err := ec2Client.DescribeInstancesPages(&ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			&ec2.Filter{
				Name:   aws.String("instance-state-name"),
				Values: []*string{aws.String("running")},
			},
			&ec2.Filter{
				Name:   aws.String(fmt.Sprintf("tag:%s", key)),
				Values: []*string{aws.String(value)},
			},
		},
	}, func(result *ec2.DescribeInstancesOutput, _ bool) bool {
		for _, reservation := range result.Reservations {
			instances = append(instances, reservation.Instances...)
		}
		return true // keep going
	})
	if err != nil {
		return nil, err
	}

	return instances, nil
}

func findInstance(ec2HostConfig *config.Ec2HostConfig, svc *ec2.EC2) (*ec2.Instance, error) {
	region, name := ec2HostConfig.Region, ec2HostConfig.Name

	log.Printf("Searching for instance named %q in region %s", name, region)

	params := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("tag:Name"),
				Values: []*string{
					aws.String(ec2HostConfig.Name),
				},
			},
		},
	}
	resp, err := svc.DescribeInstances(params)
	if err != nil {
		return nil, err
	}

	if len(resp.Reservations) < 1 {
		return nil, fmt.Errorf("no instance found in %s with name %s", region, name)
	}
	if len(resp.Reservations) > 1 || len(resp.Reservations[0].Instances) != 1 {
		return nil, fmt.Errorf("multiple instances found in %s with name %s", region, name)
	}
	return resp.Reservations[0].Instances[0], nil
}

func createSecurityGroups(c *ec2.EC2, config *Config) error {
	for j := range config.AllSecurityGroups {
		csgi := &ec2.CreateSecurityGroupInput{GroupName: &config.AllSecurityGroups[j].Name, VpcId: &config.VpcId, Description: &config.AllSecurityGroups[j].Name}
		csgo, err := c.CreateSecurityGroup(csgi)
		//fmt.Println(err)
		if err != nil {
			if !strings.Contains(fmt.Sprintf("%s", err), "InvalidGroup.Duplicate") {
				fmt.Println("Failed to create security group.")
				return err
			}
			continue
		}

		everywhere := "0.0.0.0/0"
		proto := "tcp"
		//var fromPort int64
		//fromPort = -1
		asgii := &ec2.AuthorizeSecurityGroupIngressInput{CidrIp: &everywhere, FromPort: &config.AllSecurityGroups[j].TcpPort, ToPort: &config.AllSecurityGroups[j].TcpPort, GroupId: csgo.GroupId, IpProtocol: &proto}
		_, err = c.AuthorizeSecurityGroupIngress(asgii)
		//fmt.Println("Adding security group", asgii)
		if err != nil {
			fmt.Println("Failed to add rule to security group: ", err)
			return err
		}
	}

	return nil

}

// setTags is a helper to set the tags for a resource. It expects the
// tags field to be named "tags"
func setTags(conn *ec2.EC2, d *schema.ResourceData) error {
	if d.HasChange("tags") {
		oraw, nraw := d.GetChange("tags")
		o := oraw.(map[string]interface{})
		n := nraw.(map[string]interface{})
		create, remove := diffTags(tagsFromMap(o), tagsFromMap(n))

		// Set tags
		if len(remove) > 0 {
			log.Printf("[DEBUG] Removing tags: %#v from %s", remove, d.Id())
			_, err := conn.DeleteTags(&ec2.DeleteTagsInput{
				Resources: []*string{aws.String(d.Id())},
				Tags:      remove,
			})
			if err != nil {
				return err
			}
		}
		if len(create) > 0 {
			log.Printf("[DEBUG] Creating tags: %s for %s", awsutil.StringValue(create), d.Id())
			_, err := conn.CreateTags(&ec2.CreateTagsInput{
				Resources: []*string{aws.String(d.Id())},
				Tags:      create,
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func awsTerminateInstance(conn *ec2.EC2, id string) error {
	log.Printf("[INFO] Terminating instance: %s", id)
	req := &ec2.TerminateInstancesInput{
		InstanceIds: []*string{aws.String(id)},
	}
	if _, err := conn.TerminateInstances(req); err != nil {
		return fmt.Errorf("Error terminating instance: %s", err)
	}

	log.Printf("[DEBUG] Waiting for instance (%s) to become terminated", id)

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"pending", "running", "shutting-down", "stopped", "stopping"},
		Target:     []string{"terminated"},
		Refresh:    InstanceStateRefreshFunc(conn, id),
		Timeout:    10 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err := stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf(
			"Error waiting for instance (%s) to terminate: %s", id, err)
	}

	return nil
}

// Returns a single insance when given it's ID
func getInstanceByID(instanceid string, ec2client *ec2.EC2) (ec2.Instance, error) {
	instanceReq := ec2.DescribeInstancesInput{
		InstanceIds: []*string{
			aws.String(instanceid),
		},
	}

	instanceResp, err := ec2client.DescribeInstances(&instanceReq)
	if err != nil {
		return ec2.Instance{}, err
	}

	//We only requested one instance, so we should only get one
	if len(instanceResp.Reservations) != 1 {
		return ec2.Instance{}, errors.New("The total number of reservations did not match the request")
	}
	reservation := instanceResp.Reservations[0]

	// Now let's make sure we only got one instance in this reservation
	if len(reservation.Instances) != 1 {
		return ec2.Instance{}, errors.New("The total number of instances did not match the request for full instance data")
	}

	instance := reservation.Instances[0]
	return *instance, nil
}
