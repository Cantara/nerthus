package main

import (
	"context"
	"embed"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
)

//go:embed nerthus_role.json nerthus_policy.json
var fsFB embed.FS

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}
	var opts []func(*config.LoadOptions) error
	if os.Getenv("aws.profile") != "" {
		opts = append(opts, config.WithSharedConfigProfile(os.Getenv("aws.profile")))
	} else {
		opts = append(opts, config.WithDefaultRegion(os.Getenv("region")))
	}
	sess, err := config.LoadDefaultConfig(context.TODO(), opts...,
	)
	if err != nil {
		log.Fatal("While creating aws session", err)
	}
	script, err := fsFB.ReadFile("nerthus_role.json")
	if err != nil {
		log.Fatal("While reading in nerthus policy", err)
		return
	}
	tmp := string(script)
	fmt.Println(tmp)
	svc := iam.NewFromConfig(sess)
	inputRole := &iam.CreateRoleInput{
		AssumeRolePolicyDocument: aws.String(tmp),
		Path:                     aws.String("/"),
		RoleName:                 aws.String("Nerthus-Role"),
	}

	resultRole, err := svc.CreateRole(context.Background(), inputRole)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(resultRole)

	scriptPol, err := fsFB.ReadFile("nerthus_policy.json")
	if err != nil {
		log.Fatal("While reading in nerthus policy", err)
		return
	}
	tmpPol := string(scriptPol)
	fmt.Println(tmpPol)
	inputPol := &iam.CreatePolicyInput{
		PolicyDocument: aws.String(tmpPol),
		PolicyName:     aws.String("Nerthus-Policy"),
	}
	resultPol, err := svc.CreatePolicy(context.Background(), inputPol)
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println(resultPol)

	inputAttach := &iam.AttachRolePolicyInput{
		PolicyArn: resultPol.Policy.Arn,
		RoleName:  inputRole.RoleName,
	}

	resultAttach, err := svc.AttachRolePolicy(context.Background(), inputAttach)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(resultAttach)

	inputProfile := &iam.CreateInstanceProfileInput{
		InstanceProfileName: aws.String("Nerthus"),
	}
	resultProfile, err := svc.CreateInstanceProfile(context.Background(), inputProfile)
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println(resultProfile)

	inputAdd := &iam.AddRoleToInstanceProfileInput{
		InstanceProfileName: inputProfile.InstanceProfileName,
		RoleName:            inputRole.RoleName,
	}
	result, err := svc.AddRoleToInstanceProfile(context.Background(), inputAdd)
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println(result)
}
