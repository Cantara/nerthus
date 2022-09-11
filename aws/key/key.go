package key

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/cantara/nerthus/aws/util"
	"time"
)

type Key struct {
	Scope       string           `json:"-"`
	Id          string           `json:"id"`
	Name        string           `json:"name"`
	PemName     string           `json:"pem_name"`
	Fingerprint string           `json:"fingerprint"`
	Material    string           `json:"material"`
	Type        ec2types.KeyType `json:"type"`
	ec2         *ec2.Client
	created     bool
}

func NewKey(scope string, e2 *ec2.Client) (k Key, err error) {
	err = util.CheckEC2Session(e2)
	if err != nil {
		return
	}
	k = Key{
		Scope: scope,
		Name:  scope + "-key",
		Type:  ec2types.KeyTypeEd25519,
		ec2:   e2,
	}
	return
}

func (k *Key) Create() (id string, err error) {
	err = util.CheckEC2Session(k.ec2)
	if err != nil {
		return
	}
	keyResult, err := k.ec2.CreateKeyPair(context.Background(), &ec2.CreateKeyPairInput{
		KeyName: aws.String(k.Name),
		KeyType: k.Type,
	})
	if err != nil {
		return
	}
	k.Id = aws.ToString(keyResult.KeyPairId)
	k.Fingerprint = aws.ToString(keyResult.KeyFingerprint)
	k.Material = aws.ToString(keyResult.KeyMaterial)
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
	err = ec2.NewKeyPairExistsWaiter(k.ec2).Wait(context.Background(), &ec2.DescribeKeyPairsInput{
		KeyPairIds: []string{
			k.Id,
		},
	}, 5*time.Minute)
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

	_, err = k.ec2.DeleteKeyPair(context.Background(), input)
	return
}

func (k Key) WithEC2(e *ec2.Client) Key {
	k.ec2 = e
	return k
}
