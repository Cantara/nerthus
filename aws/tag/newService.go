package tag

import (
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	elbv2 "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/cantara/nerthus/aws/util"
)

type New struct {
	Scope              string `json:"-"`
	Id                 string `json:"id"`
	Name               string `json:"name"`
	KeyId              string `json:"key_id"`
	SecurityGroupId    string `json:"security_group_id"`
	ServerId           string `json:"server_id"`
	VolumeId           string `json:"volume_id"`
	NetworkInterfaceId string `json:"network_interface_id"`
	ImageId            string `json:"image_id"`
	TargetGroupARN     string `json:"target_group_arn"`
	RuleARN            string `json:"rule_arn"`
	ListenerARN        string `json:"listener_arn"`
	LoadbalancerARN    string `json:"loadbalancer_arn"`
	tag                *tag
	created            bool
}

func NewNewTag(serviceName, scope, keyId, securityGroupId, serverId, volumeId, networkInterfaceId, imageId,
	targetGroupARN, ruleARN, listnerARN, loadbalancerARN string,
	e2 *ec2.Client, el *elbv2.Client) (t New, err error) {

	err = util.CheckEC2Session(e2)
	if err != nil {
		return
	}
	err = util.CheckELBV2Session(el)
	if err != nil {
		return
	}
	t = New{
		Scope:              scope,
		Name:               serviceName,
		KeyId:              keyId,
		SecurityGroupId:    securityGroupId,
		ServerId:           serverId,
		VolumeId:           volumeId,
		NetworkInterfaceId: networkInterfaceId,
		ImageId:            imageId,
		TargetGroupARN:     targetGroupARN,
		RuleARN:            ruleARN,
		ListenerARN:        listnerARN,
		LoadbalancerARN:    loadbalancerARN,
		tag: &tag{
			ec2Resources: []string{
				keyId,
				securityGroupId,
				serverId,
				networkInterfaceId,
				volumeId,
				imageId,
			},
			elbResources: []string{
				targetGroupARN,
				ruleARN,
				listnerARN,
				loadbalancerARN,
			},
			Key:   serviceName,
			Value: scope,
			ec2:   e2,
			elb:   el,
		},
	}
	return
}

func (t *New) Create() (id string, err error) {
	id, err = t.tag.Create()
	return
}

func (t *New) Delete() (err error) {
	err = t.tag.Delete()
	return
}
