package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	log "github.com/cantara/bragi"
	cloud "github.com/cantara/nerthus/aws"
	"github.com/joho/godotenv"
)

func loadEnv() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}
}

var token string

func main() {
	loadEnv()
	region := "us-west-2" //"eu-central-1"
	serverName := "devtest-entraos-notification3"
	port := 18840
	uriPath := "notifications"
	elbListenerArn := "arn:aws:elasticloadbalancing:us-west-2:493376950721:listener/app/devtest-events2-lb/a3807cba101b280b/90abaa841820e9b2"
	elbSecurityGroupId := "sg-1325864d"
	shouldCleanUp := false
	deleters := NewStack()
	defer func() {
		if !shouldCleanUp {
			return
		}
		log.Info("Cleanup started.")
		for delFunc := deleters.Pop(); delFunc != nil; delFunc = deleters.Pop() {
			delFunc()
		}
		log.Info("Cleanup is \"done\", exiting.")
	}()

	// Initialize a session in us-west-2 that the SDK will use to load
	// credentials from the shared credentials file ~/.aws/credentials.
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region)},
	)

	var c cloud.AWS
	// Create an EC2 service client.
	c.NewEC2(sess)

	// Create a new key
	keyName, keyFingerprint, keyMaterial, err := c.CreateKeyPair(serverName)
	if err != nil {
		shouldCleanUp = true
		log.AddError(err).Fatal("While creating keypair")
	}
	deleters.Push(lateExecuteDeletersWithErrorLogging("Key pair", "while deleting created key pair",
		c.DeleteKeyPair, keyName))
	log.Printf("Created key pair %s %s\n%s\n",
		keyName, keyFingerprint,
		keyMaterial)
	pemName := fmt.Sprintf("%s.pem", keyName)
	pem, err := os.OpenFile("./"+pemName, os.O_WRONLY|os.O_CREATE, 0600)
	if err == nil {
		fmt.Fprint(pem, keyMaterial)
		pem.Close()
	}
	defer os.Remove(pemName)

	// Get a list of VPCs so we can associate the group with the first VPC.
	vpcId, err := c.GetVPCId()
	if err != nil {
		shouldCleanUp = true
		log.AddError(err).Fatal("While getting vpcId")
	}

	securityGroupId, err := c.CreateSecurityGroup(serverName, vpcId)
	if err != nil {
		shouldCleanUp = true
		log.AddError(err).Fatal("While creating security group")
	}
	deleters.Push(lateExecuteDeletersWithErrorLogging("Security group", "while deleting created security group",
		c.DeleteSecurityGroup, securityGroupId))
	log.Printf("Created security group %s with VPC %s.\n",
		securityGroupId, vpcId)

	err = c.AddBaseAuthorization(securityGroupId, elbSecurityGroupId, port)
	if err != nil {
		shouldCleanUp = true
		log.AddError(err).Fatal("Could not add base authorization")
	}

	serverId, err := c.CreateServer(serverName, keyName, securityGroupId)
	if err != nil {
		shouldCleanUp = true
		log.Fatal("Could not create server", err)
	}
	deleters.Push(lateExecuteDeletersWithErrorLogging("Server", "while deleting created server",
		c.DeleteServer, serverId))
	log.Info("Created server", serverId)

	if false { // Enable hazelcast
		err = c.AuthorizeHazelcast(securityGroupId)
		if err != nil {
			shouldCleanUp = true
			log.AddError(err).Fatal("Could not add hazelcast authorization")
		}
	}

	err = c.WaitForRunning(serverId)
	publicDns, err := c.GetPublicDNS(serverId)
	if err != nil {
		shouldCleanUp = true
		log.AddError(err).Fatal("While getting public dns name")
	}

	c.NewELB(sess)

	targetGroupArn, err := c.CreateTargetGroup(serverName, vpcId, uriPath, port)
	if err != nil {
		shouldCleanUp = true
		log.AddError(err).Fatal(fmt.Sprintf("While creating target group for %s", serverName))
	}
	deleters.Push(lateExecuteDeletersWithErrorLogging("Target group", "while deleting created target group",
		c.DeleteTargetGroup, targetGroupArn))

	err = c.RegisterTarget(targetGroupArn, serverId)
	if err != nil {
		shouldCleanUp = true
		log.AddError(err).Fatal(fmt.Sprintf("While adding target to target group %s", targetGroupArn))
	}
	deleters.Push(lateExecuteDeletersWithErrorLogging("Target in targetgroup", "while removing registered targetgroup",
		c.RemoveTargetGroupTarget, targetGroupArn, serverId))

	ruleArn, err := c.AddRuleToLoadBalancer(targetGroupArn, elbListenerArn, uriPath)
	if err != nil {
		shouldCleanUp = true
		log.AddError(err).Fatal(fmt.Sprintf("While adding rule to elb %s", elbListenerArn))
	}
	deleters.Push(lateExecuteDeletersWithErrorLogging("Rule", "while removing rule added to loadbalancer",
		c.DeleteRule, ruleArn))

	log.Info("Done creating")

	time.Sleep(30 * time.Second)
	script, err := ioutil.ReadFile("./new_devtest_server.sh")
	if err != nil {
		log.AddError(err).Fatal("While reading in base script")
	}
	cmd := exec.Command("ssh", "-o", "UserKnownHostsFile=/dev/null", "-o", "StrictHostKeyChecking=no",
		fmt.Sprintf("ec2-user@%s", publicDns), "-i", "./"+pemName, "/bin/bash -s")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.AddError(err).Fatal("While creating stdin pipe")
	}
	defer stdin.Close()
	io.WriteString(stdin, string(script))

	err = cmd.Run()
	if err != nil {
		shouldCleanUp = true
		log.AddError(err).Fatal("While running ssh settup command")
	}

	script, err = ioutil.ReadFile("./new_devtest_server.sh")
	if err != nil {
		log.AddError(err).Fatal("While reading in filebeat script")
	}
	cmd = exec.Command("ssh", "-o", "UserKnownHostsFile=/dev/null", "-o", "StrictHostKeyChecking=no",
		fmt.Sprintf("ec2-user@%s", publicDns), "-i", "./"+pemName, "/bin/bash -s")
	stdin, err = cmd.StdinPipe()
	if err != nil {
		log.AddError(err).Fatal("While creating stdin pipe")
	}
	defer stdin.Close()
	io.WriteString(stdin, string(script))

	err = cmd.Run()
	if err != nil {
		shouldCleanUp = true
		log.AddError(err).Fatal("While running ssh filebeat settup")
	}

	err = sendSlackMessage(fmt.Sprintf("`ssh ec2-user@%s -i %s`\n%[2]s\n```%s```", publicDns, pemName, keyMaterial))
	if err != nil {
		shouldCleanUp = true
		log.Fatal(err)
	}

}

func lateExecuteDeletersWithErrorLogging(object, logMessage string, f func(...string) error, values ...string) func() {
	return func() {
		log.Info("Cleaning up: ", object)
		err := f(values...)
		if err != nil {
			log.AddError(err).Crit(logMessage)
		}
	}
}

func NewStack() Stack {
	return Stack{}
}

type Stack struct {
	funcs []func()
}

func (s *Stack) Push(fun func()) {
	s.funcs = append(s.funcs, fun)
}

func (s *Stack) Pop() func() {
	if s.Empty() {
		return nil
	}
	fun := s.Last()
	s.funcs = s.funcs[:s.Len()-1]
	return fun
}

func (s Stack) Len() int {
	return len(s.funcs)
}

func (s Stack) Last() func() {
	if s.Empty() {
		return nil
	}
	return s.funcs[s.Len()-1]
}

func (s Stack) First() func() {
	if s.Empty() {
		return nil
	}
	return s.funcs[0]
}

func (s Stack) Empty() bool {
	return s.Len() == 0
}
