package main

import (
	"embed"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
)

//go:embed nerthus_role.json nerthus_policy.json
var fsFB embed.FS

func main() {
	/*region := "ap-northeast-1"
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region)},
	)
	if err != nil {
		log.Fatal("While creating aws session", err)
	}*/
	script, err := fsFB.ReadFile("nerthus_role.json")
	if err != nil {
		log.Fatal("While reading in nerthus policy", err)
		return
	}
	tmp := string(script)
	fmt.Println(tmp)
	svc := iam.New(session.New())
	inputRole := &iam.CreateRoleInput{
		AssumeRolePolicyDocument: aws.String(tmp),
		Path:                     aws.String("/"),
		RoleName:                 aws.String("Test-Role"),
	}

	resultRole, err := svc.CreateRole(inputRole)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case iam.ErrCodeLimitExceededException:
				fmt.Println(iam.ErrCodeLimitExceededException, aerr.Error())
			case iam.ErrCodeInvalidInputException:
				fmt.Println(iam.ErrCodeInvalidInputException, aerr.Error())
			case iam.ErrCodeEntityAlreadyExistsException:
				fmt.Println(iam.ErrCodeEntityAlreadyExistsException, aerr.Error())
			case iam.ErrCodeMalformedPolicyDocumentException:
				fmt.Println(iam.ErrCodeMalformedPolicyDocumentException, aerr.Error())
			case iam.ErrCodeConcurrentModificationException:
				fmt.Println(iam.ErrCodeConcurrentModificationException, aerr.Error())
			case iam.ErrCodeServiceFailureException:
				fmt.Println(iam.ErrCodeServiceFailureException, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
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
		PolicyName:     aws.String("Test-Policy"),
	}
	resultPol, err := svc.CreatePolicy(inputPol)
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println(resultPol)

	input := &iam.AttachRolePolicyInput{
		PolicyArn: resultPol.Policy.Arn,
		RoleName:  inputRole.RoleName,
	}

	result, err := svc.AttachRolePolicy(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case iam.ErrCodeNoSuchEntityException:
				fmt.Println(iam.ErrCodeNoSuchEntityException, aerr.Error())
			case iam.ErrCodeLimitExceededException:
				fmt.Println(iam.ErrCodeLimitExceededException, aerr.Error())
			case iam.ErrCodeInvalidInputException:
				fmt.Println(iam.ErrCodeInvalidInputException, aerr.Error())
			case iam.ErrCodeUnmodifiableEntityException:
				fmt.Println(iam.ErrCodeUnmodifiableEntityException, aerr.Error())
			case iam.ErrCodePolicyNotAttachableException:
				fmt.Println(iam.ErrCodePolicyNotAttachableException, aerr.Error())
			case iam.ErrCodeServiceFailureException:
				fmt.Println(iam.ErrCodeServiceFailureException, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return
	}

	fmt.Println(result)
}
