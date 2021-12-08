package key

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/cantara/nerthus/aws/util"
)

type Key struct {
	Name        string
	Fingerprint string
	Material    string
	Type        string
	ec2         *ec2.EC2
}

func NewKey(name string, e2 *ec2.EC2) (k Key, err error) {
	err = util.CheckEC2Session(e2)
	if err != nil {
		return
	}
	k = Key{
		Name: name + "-key",
		Type: "ed25519",
		ec2:  e2,
	}
	return
}

func (k *Key) Create() (id string, err error) {
	err = util.CheckEC2Session(k.ec2)
	if err != nil {
		return
	}
	keyResult, err := k.ec2.CreateKeyPair(&ec2.CreateKeyPairInput{
		KeyName: aws.String(k.Name),
		KeyType: aws.String(k.Type),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == "InvalidKeyPair.Duplicate" {
			err = util.CreateError{
				Text: "Duplicate key pair",
				Err:  aerr,
			}
			return
		}
		err = util.CreateError{
			Text: "Unable to create key pair: " + k.Name,
			Err:  err,
		}
		return
	}
	k.Fingerprint = aws.StringValue(keyResult.KeyFingerprint)
	k.Material = aws.StringValue(keyResult.KeyMaterial)
	id = k.Name
	return
}

func (k *Key) Delete() (err error) {
	err = util.CheckEC2Session(k.ec2)
	if err != nil {
		return
	}
	input := &ec2.DeleteKeyPairInput{
		KeyName: aws.String(k.Name),
	}

	_, err = k.ec2.DeleteKeyPair(input)
	return
}
