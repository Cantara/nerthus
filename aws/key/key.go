package key

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/cantara/nerthus/aws/util"
)

type Key struct {
	Scope       string `json:"-"`
	Id          string `json:"id"`
	Name        string `json:"name"`
	PemName     string `json:"pem_name"`
	Fingerprint string `json:"fingerprint"`
	Material    string `json:"material"`
	Type        string `json:"type"`
	ec2         *ec2.EC2
	created     bool
}

func NewKey(scope string, e2 *ec2.EC2) (k Key, err error) {
	err = util.CheckEC2Session(e2)
	if err != nil {
		return
	}
	k = Key{
		Scope: scope,
		Name:  scope + "-key",
		Type:  "ed25519",
		ec2:   e2,
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
	k.Id = aws.StringValue(keyResult.KeyPairId)
	k.Fingerprint = aws.StringValue(keyResult.KeyFingerprint)
	k.Material = aws.StringValue(keyResult.KeyMaterial)
	k.PemName = k.Name + ".pem"
	id = k.Id
	k.created = true
	return
}

func (k Key) Wait() (err error) {
	if !k.created {
		return
	}
	err = util.CheckEC2Session(k.ec2)
	if err != nil {
		return
	}
	err = k.ec2.WaitUntilKeyPairExists(&ec2.DescribeKeyPairsInput{
		KeyPairIds: []*string{
			aws.String(k.Id),
		},
	})
	return
}

func (k *Key) Delete() (err error) {
	if !k.created {
		return
	}
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

func (k Key) WithEC2(e *ec2.EC2) Key {
	k.ec2 = e
	return k
}
