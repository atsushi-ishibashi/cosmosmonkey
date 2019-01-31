package svc

import (
	"fmt"
	"os"

	"github.com/atsushi-ishibashi/cosmosmonkey/model"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/ecs/ecsiface"
)

type ECSService interface {
	ListClusterInstances(cluster string) ([]model.ClusterInstance, error)
	DrainContainerInstance(instance model.ClusterInstance) error
	GetContainerInstanceStatus(instance model.ClusterInstance) (string, error)
}

type ecsService struct {
	svc ecsiface.ECSAPI
}

func NewECSService() ECSService {
	return &ecsService{
		svc: ecs.New(session.New(), aws.NewConfig().WithRegion(os.Getenv("_CM_AWS_REGION"))),
	}
}

func (s *ecsService) ListClusterInstances(cluster string) ([]model.ClusterInstance, error) {
	result := make([]model.ClusterInstance, 0)
	linput := &ecs.ListContainerInstancesInput{
		Cluster: aws.String(cluster),
	}
	lresp, err := s.svc.ListContainerInstances(linput)
	if err != nil {
		return result, err
	}

	dinput := &ecs.DescribeContainerInstancesInput{
		Cluster:            aws.String(cluster),
		ContainerInstances: lresp.ContainerInstanceArns,
	}
	dresp, err := s.svc.DescribeContainerInstances(dinput)
	if err != nil {
		return result, err
	}

	for _, v := range dresp.ContainerInstances {
		ci := model.ClusterInstance{
			Cluster:            cluster,
			InstanceID:         aws.StringValue(v.Ec2InstanceId),
			Status:             aws.StringValue(v.Status),
			ClusterInstanceArn: aws.StringValue(v.ContainerInstanceArn),
		}
		result = append(result, ci)
	}

	return result, nil
}

func (s *ecsService) DrainContainerInstance(instance model.ClusterInstance) error {
	input := &ecs.DeregisterContainerInstanceInput{
		Cluster:           aws.String(instance.Cluster),
		ContainerInstance: aws.String(instance.ClusterInstanceArn),
		Force:             aws.Bool(false),
	}
	_, err := s.svc.DeregisterContainerInstance(input)
	return err
}

func (s *ecsService) GetContainerInstanceStatus(instance model.ClusterInstance) (string, error) {
	input := &ecs.DescribeContainerInstancesInput{
		Cluster:            aws.String(instance.Cluster),
		ContainerInstances: []*string{aws.String(instance.ClusterInstanceArn)},
	}
	resp, err := s.svc.DescribeContainerInstances(input)
	if err != nil {
		return "", err
	}
	if len(resp.ContainerInstances) != 1 {
		return "", fmt.Errorf("expect #ContainerInstances == 1, got %d", len(resp.ContainerInstances))
	}
	ci := resp.ContainerInstances[0]
	return aws.StringValue(ci.Status), nil
}
