package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	log "github.com/cantara/bragi"
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
	log.Println("vim-go")
	loadEnv()
	serverName := "TestServer9"
	region := "us-west-2"

	// Initialize a session in us-west-2 that the SDK will use to load
	// credentials from the shared credentials file ~/.aws/credentials.
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region)},
	)

	// Create an EC2 service client.
	svc := ec2.New(sess)

	pairName := serverName + "-key"
	// Creates a new  key pair with the given name
	keyResult, err := svc.CreateKeyPair(&ec2.CreateKeyPairInput{
		KeyName: aws.String(pairName),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == "InvalidKeyPair.Duplicate" {
			log.Fatalf("Keypair %q already exists.", pairName)
		}
		log.Fatalf("Unable to create key pair: %s, %v.", pairName, err)
	}

	fmt.Printf("Created key pair %q %s\n%s\n",
		*keyResult.KeyName, *keyResult.KeyFingerprint,
		*keyResult.KeyMaterial)
	err = sendSlackMessage(fmt.Sprintf("%s.pem\n```%s```", pairName, *keyResult.KeyMaterial))
	if err != nil {
		log.Fatal(err)
	}

	// Get a list of VPCs so we can associate the group with the first VPC.
	result, err := svc.DescribeVpcs(nil)
	if err != nil {
		log.Fatalf("Unable to describe VPCs, %v", err)
	}
	if len(result.Vpcs) == 0 {
		log.Fatalf("No VPCs found to associate security group with.")
	}

	secName := serverName + "-sg"
	vpcID := aws.StringValue(result.Vpcs[0].VpcId)
	secGroupRes, err := svc.CreateSecurityGroup(&ec2.CreateSecurityGroupInput{
		GroupName:   aws.String(secName),
		Description: aws.String("Security group for server: " + serverName),
		VpcId:       aws.String(vpcID),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case "InvalidVpcID.NotFound":
				log.Fatalf("Unable to find VPC with ID %q.", vpcID)
			case "InvalidGroup.Duplicate":
				log.Fatalf("Security group %q already exists.", secName)
			}
		}
		log.Fatalf("Unable to create security group %q, %v", secName, err)
	}

	fmt.Printf("Created security group %s with VPC %s.\n",
		aws.StringValue(secGroupRes.GroupId), vpcID)

	// Add tags to the created instance
	_, errtag := svc.CreateTags(&ec2.CreateTagsInput{
		//Resources: []*string{runResult.Instances[0].InstanceId},
		Resources: []*string{secGroupRes.GroupId},
		Tags: []*ec2.Tag{
			{
				Key:   aws.String("Name"),
				Value: aws.String(serverName),
			},
		},
	})
	if errtag != nil {
		log.Println("Could not create tags for sg", secGroupRes.GroupId, errtag)
		return
	}

	fmt.Println("Successfully tagged sg")

	/*
		input := &ec2.CreateVolumeInput{
			AvailabilityZone: aws.String(region + "a"),
			Size:             aws.Int64(20),
			VolumeType:       aws.String("gp3"),
		}

		result, err := svc.CreateVolume(input)
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				switch aerr.Code() {
				default:
					log.Fatal(aerr.Error())
				}
			} else {
				// Print the error, cast err to awserr.Error to get the Code and
				// Message from an error.
				log.Fatal(err.Error())
			}
		}
	*/

	// Specify the details of the instance that you want to create
	runResult, err := svc.RunInstances(&ec2.RunInstancesInput{
		ImageId:          aws.String("ami-0142f6ace1c558c7d"),
		InstanceType:     aws.String("t3.micro"),
		MinCount:         aws.Int64(1),
		MaxCount:         aws.Int64(1),
		SecurityGroupIds: aws.StringSlice([]string{*secGroupRes.GroupId}),
		KeyName:          aws.String(pairName),
		BlockDeviceMappings: []*ec2.BlockDeviceMapping{
			{
				DeviceName: aws.String("/dev/xvda"),
				Ebs: &ec2.EbsBlockDevice{
					VolumeSize: aws.Int64(20),
					VolumeType: aws.String("gp3"),
				},
			},
		},
		TagSpecifications: []*ec2.TagSpecification{
			{
				ResourceType: aws.String("instance"),
				Tags: []*ec2.Tag{
					{
						Key:   aws.String("Name"),
						Value: aws.String(serverName),
					},
				},
			},
			{
				ResourceType: aws.String("volume"),
				Tags: []*ec2.Tag{
					{
						Key:   aws.String("Name"),
						Value: aws.String(serverName), // + "-vol"),
					},
				},
			},
		},
	})

	if err != nil {
		fmt.Println("Could not create instance", err)
		return
	}

	fmt.Println("Created instance", *runResult.Instances[0].InstanceId)

	input := &ec2.AuthorizeSecurityGroupIngressInput{
		GroupId: secGroupRes.GroupId,
		IpPermissions: []*ec2.IpPermission{
			{
				FromPort:   aws.Int64(80),
				IpProtocol: aws.String("tcp"),
				ToPort:     aws.Int64(80),
				UserIdGroupPairs: []*ec2.UserIdGroupPair{
					{
						Description: aws.String("HTTP access from other instances"),
						GroupId:     aws.String("sg-1325864d"), //TODO: dynamically get loadbalancer sg
					},
				},
			},
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

	_, err = svc.AuthorizeSecurityGroupIngress(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				log.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			log.Println(err.Error())
		}
	}

	log.Debug(runResult)
	describeInstancesInput := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{
			runResult.Instances[0].InstanceId,
		},
	}

	describeInstances, err := svc.DescribeInstances(describeInstancesInput)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				log.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			log.Println(err.Error())
		}
		return
	}

	sendSlackMessage(fmt.Sprintf("`ssh -i ec2-user@%s %s.pem`", *describeInstances.Reservations[0].Instances[0].PublicDnsName, pairName))
}

type slackMessage struct {
	SlackId string `json:"recepientId"`
	Message string `json:"message"`
	//	Username    string   `json:"username"`
	Pinned bool `json:"pinned"`
	//	Attachments []string `json:"attachments"`
}

func sendSlackMessage(message string) (err error) {
	if token == "" {
		token, err = getWhydahAuthToken()
		for count := 0; err != nil && count < 10; count++ {
			token, err = getWhydahAuthToken()
		}
	}
	return postAuth(os.Getenv("entraos_api_uri")+"/slack/api/message", slackMessage{
		SlackId: os.Getenv("slack_channel"),
		Message: message,
		Pinned:  false,
	}, nil, token)
}

type applicationcredential struct {
	Params applicationCredentialParams `xml:"params"`
}

type applicationCredentialParams struct {
	AppId     string `xml:"applicationID"`
	AppName   string `xml:"applicationName"`
	AppSecret string `xml:"applicationSecret"`
}

type applicationtoken struct {
	Params applicationTokenParams `xml:"params"`
}

type applicationTokenParams struct {
	AppTokenId string `xml:"applicationtokenID"`
	AppId      string `xml:"applicationid"`
	AppName    string `xml:"applicationName"`
	expires    int    `xml:"expires"`
}

func getWhydahAuthToken() (token string, err error) {
	appCred := applicationcredential{
		Params: applicationCredentialParams{
			AppId:     os.Getenv("whydah_application_id"),
			AppName:   os.Getenv("whydah_application_name"),
			AppSecret: os.Getenv("whydah_application_secret"),
		},
	}
	appCredXML, err := xml.Marshal(appCred)
	if err != nil {
		return
	}
	data := url.Values{
		"applicationcredential": {string(appCredXML)},
	}
	log.Println(data)
	resp, err := http.PostForm(os.Getenv("whydah_uri")+"/tokenservice/logon", data)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	var tokenData applicationtoken
	err = xml.Unmarshal(body, &tokenData)
	if err != nil {
		return
	}
	token = tokenData.Params.AppTokenId
	return
}

func postAuth(uri string, data interface{}, out interface{}, token string) (err error) {
	log.Info(token)
	jsonValue, _ := json.Marshal(data)
	client := &http.Client{}
	req, err := http.NewRequest("POST", uri, bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := client.Do(req)
	log.Info(resp.StatusCode)
	if err != nil || out == nil {
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, out)
	return
}

/*
func AddUser() models.Resp {
	res := new(models.User)
	var jsonData = []byte(`{"first_name":"` + res.Fname + `", "last_name":"` + res.Lname + `"}`)
	client := &http.Client{}
	req, err := http.NewRequest("POST", os.Env(""), bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	resp, er := client.Do(req)
	fmt.Println(resp.Body)

	if er != nil {
		log.Info("Error in reqeust send")
	}

	if err != nil {
		log.Info("Error in reqeust create")
	}
	defer resp.Body.Close()

	fmt.Println(resp.StatusCode)
	if resp.StatusCode == 200 {
		log.Info("Successfully! Added User")

	}
	var data models.Resp
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		fmt.Println(err)
	}

	return data
}
*/
