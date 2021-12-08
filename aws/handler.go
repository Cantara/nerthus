package aws

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"time"

	log "github.com/cantara/bragi"
	keylib "github.com/cantara/nerthus/aws/key"
	loadbalancerlib "github.com/cantara/nerthus/aws/loadbalancer"
	securitylib "github.com/cantara/nerthus/aws/security"
	serverlib "github.com/cantara/nerthus/aws/server"
	vpclib "github.com/cantara/nerthus/aws/vpc"
	"github.com/cantara/nerthus/crypto"
	"github.com/cantara/nerthus/slack"
)

func (c AWS) CreateFromScratch(service, path string, port int, ELBListenerArn, ELBSecurityGroup string) {
	shouldCleanUp := false
	deleters := NewStack()
	defer func() {
		if a := recover(); a != nil {
			log.Warning("Recovered: ", a)
		}
		if !shouldCleanUp {
			return
		}
		log.Info("Cleanup started.")
		slack.SendStatus("Something went wrong starting cleanup.")
		for delFunc := deleters.Pop(); delFunc != nil; delFunc = deleters.Pop() {
			delFunc()
		}
		log.Info("Cleanup is \"done\", exiting.")
		slack.SendStatus("Cleanup is \"done\".")
	}()

	// Create a new key
	key, err := keylib.NewKey(service, c.ec2)
	_, err = key.Create()
	if err != nil {
		log.AddError(err).Fatal("While creating keypair")
	}
	deleters.Push(cleanup("Key pair", "while deleting created key pair", &key))
	s := fmt.Sprintf("Created key pair %s %s", key.Name, key.Fingerprint)
	log.Info(s)
	slack.SendStatus(s)
	keyEncrypted, err := crypto.Encrypt(key.Material)
	if err != nil {
		shouldCleanUp = true
		log.AddError(err).Fatal("while encrypting key")
	}
	pemName := fmt.Sprintf("%s.pem", key.Name)
	pem, err := os.OpenFile("./"+pemName, os.O_WRONLY|os.O_CREATE, 0600)
	if err == nil {
		fmt.Fprint(pem, key.Material)
		pem.Close()
	}
	defer os.Remove(pemName)

	// Get a list of VPCs so we can associate the group with the first VPC.
	vpc, err := vpclib.GetVPC(c.ec2)
	if err != nil {
		shouldCleanUp = true
		log.AddError(err).Fatal("While getting vpcId")
	}
	s = fmt.Sprintf("Found VPCId: %s.", vpc.Id)
	log.Info(s)
	slack.SendStatus(s)

	securityGroup, err := securitylib.NewGroup(service, vpc, c.ec2)
	_, err = securityGroup.Create()
	if err != nil {
		shouldCleanUp = true
		log.AddError(err).Fatal("While creating security group")
	}
	deleters.Push(cleanup("Security group", "while deleting created security group",
		&securityGroup))
	s = fmt.Sprintf("Created security group %s with VPC %s.",
		securityGroup.Id, vpc.Id)
	log.Info(s)
	slack.SendStatus(s)

	err = securityGroup.AddBaseAuthorization(ELBSecurityGroup, port)
	if err != nil {
		shouldCleanUp = true
		log.AddError(err).Fatal("Could not add base authorization")
	}
	s = fmt.Sprintf("Added base authorization to security group: %s.", securityGroup.Id)
	log.Info(s)
	slack.SendStatus(s)

	server, err := serverlib.NewServer(service, key, securityGroup, c.ec2)
	_, err = server.Create()
	if err != nil {
		shouldCleanUp = true
		log.Fatal("Could not create server", err)
	}
	deleters.Push(cleanup("Server", "while deleting created server", &server))
	s = fmt.Sprintf("Created server: %s.", server.Id)
	log.Info(s)
	slack.SendStatus(s)

	if false { // Enable hazelcast
		err = securityGroup.AuthorizeHazelcast()
		if err != nil {
			shouldCleanUp = true
			log.AddError(err).Fatal("Could not add hazelcast authorization")
		}
		s = fmt.Sprintf("Added hazelcast authorization to security group: %s.", securityGroup.Id)
		log.Info(s)
		slack.SendStatus(s)
	}

	err = server.WaitForRunning()
	s = fmt.Sprintf("Server %s is now in running state.", server.Id)
	log.Info(s)
	slack.SendStatus(s)
	_, err = server.GetPublicDNS()
	if err != nil {
		shouldCleanUp = true
		log.AddError(err).Fatal("While getting public dns name")
	}
	s = fmt.Sprintf("Got server %s's public dns %s.", server.Id, server.PublicDNS)
	log.Info(s)
	slack.SendStatus(s)

	targetGroup, err := loadbalancerlib.NewTargetGroup(service, path, port, vpc, c.elb)
	_, err = targetGroup.Create()
	if err != nil {
		shouldCleanUp = true
		log.AddError(err).Fatal(fmt.Sprintf("While creating target group for %s", server.Name))
	}
	deleters.Push(cleanup("Target group", "while deleting created target group", &targetGroup))
	s = fmt.Sprintf("Created target group: %s.", targetGroup.ARN)
	log.Info(s)
	slack.SendStatus(s)

	target, err := loadbalancerlib.NewTarget(targetGroup, server, c.elb)
	_, err = target.Create()
	if err != nil {
		shouldCleanUp = true
		log.AddError(err).Fatal(fmt.Sprintf("While adding target to target group %s", targetGroup.ARN))
	}
	deleters.Push(cleanup("Target in targetgroup", "while removing registered targetgroup", &target))
	s = fmt.Sprintf("Registered server %s as target for target group %s.", server.Id, targetGroup.ARN)
	log.Info(s)
	slack.SendStatus(s)

	listener, err := loadbalancerlib.GetListener(ELBListenerArn, c.elb)
	rule, err := loadbalancerlib.NewRule(listener, targetGroup, c.elb)
	_, err = rule.Create()
	if err != nil {
		shouldCleanUp = true
		log.AddError(err).Fatal(fmt.Sprintf("While adding rule to elb %s", listener.ARN))
	}
	deleters.Push(cleanup("Rule", "while removing rule added to loadbalancer", &rule))
	s = fmt.Sprintf("Adding elastic load balancer rule: %s.", rule.ARN)
	log.Info(s)
	slack.SendStatus(s)
	s = fmt.Sprintf("Done setting up server in aws %s.", server.Id)
	log.Info(s)
	slack.SendStatus(s)

	s = fmt.Sprintf("Started waiting for elb rule to healthy %s.", rule.ARN)
	log.Info(s)
	slack.SendStatus(s)
	time.Sleep(30 * time.Second)
	s = fmt.Sprintf("Done waiting for elb rule to healthy %s.", rule.ARN)
	log.Info(s)
	slack.SendStatus(s)
	s = fmt.Sprintf("Starting to install stuff on server %s.", server.PublicDNS)
	log.Info(s)
	slack.SendStatus(s)
	script, err := ioutil.ReadFile("./new_devtest_server.sh")
	if err != nil {
		log.AddError(err).Fatal("While reading in base script")
	}
	cmd := exec.Command("ssh", "-o", "UserKnownHostsFile=/dev/null", "-o", "StrictHostKeyChecking=no",
		fmt.Sprintf("ec2-user@%s", server.PublicDNS), "-i", "./"+pemName, "/bin/bash -s")
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
	s = fmt.Sprintf("Done installing service on server %s.", server.PublicDNS)
	log.Info(s)
	slack.SendStatus(s)

	script, err = ioutil.ReadFile("./new_devtest_server.sh")
	if err != nil {
		log.AddError(err).Fatal("While reading in filebeat script")
	}
	cmd = exec.Command("ssh", "-o", "UserKnownHostsFile=/dev/null", "-o", "StrictHostKeyChecking=no",
		fmt.Sprintf("ec2-user@%s", server.PublicDNS), "-i", "./"+pemName, "/bin/bash -s")
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
	s = fmt.Sprintf("Done installing filebeat on server %s.", server.PublicDNS)
	log.Info(s)
	slack.SendStatus(s)
	err = slack.SendServer(fmt.Sprintf("`ssh ec2-user@%s -i %s`\n%[2]s\n```%s```", server.PublicDNS, pemName, keyEncrypted))
	if err != nil {
		shouldCleanUp = true
		log.Fatal(err)
	}
	s = fmt.Sprintf("Completed all opperations for creating the new server %s.", server.Name)
	log.Info(s)
	slack.SendStatus(s)
	shouldCleanUp = true
}
