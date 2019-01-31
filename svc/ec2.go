package svc

import (
	"os"

	"github.com/atsushi-ishibashi/cosmosmonkey/model"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
)

type EC2Service interface {
	DestroyInstance(instance model.ClusterInstance) error
}

type ec2Service struct {
	svc ec2iface.EC2API
}

func NewEC2Service() EC2Service {
	return &ec2Service{
		svc: ec2.New(session.New(), aws.NewConfig().WithRegion(os.Getenv("_CM_AWS_REGION"))),
	}
}

func (s *ec2Service) DestroyInstance(instance model.ClusterInstance) error {
	input := &ec2.TerminateInstancesInput{
		InstanceIds: []*string{aws.String(instance.InstanceID)},
	}
	_, err := s.svc.TerminateInstances(input)
	return err
}
