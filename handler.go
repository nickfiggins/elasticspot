package elasticspot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type V1Handler struct {
	Ec2       EC2API
	ElasticIP string
}

func NewV1Handler(ec2 EC2API, elasticIp string) *V1Handler {
	return &V1Handler{Ec2: ec2, ElasticIP: elasticIp}
}

type EC2API interface {
	DescribeAddresses(input *ec2.DescribeAddressesInput) (*ec2.DescribeAddressesOutput, error)
	AssociateAddress(input *ec2.AssociateAddressInput) (*ec2.AssociateAddressOutput, error)
	DescribeInstances(input *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error)
}

type SuccessResponse struct {
	InstanceId string `json:"instanceID,omitempty"`
	ElasticIP  string `json:"elasticIP,omitempty"`
	Message    string `json:"message,omitempty"`
}

func (h *V1Handler) Handle(ctx context.Context, event *events.CloudWatchEvent) (*SuccessResponse, error) {
	var eventDetails EventDetails
	if err := json.Unmarshal(event.Detail, &eventDetails); err != nil {
		return nil, fmt.Errorf("error unmarshaling cloudwatch event: %w", err)
	}

	instanceID := eventDetails.Ec2Instanceid

	instance, err := h.getInstanceById(instanceID)
	if err != nil {
		return nil, fmt.Errorf("error fetching instance: %w", err)
	}

	if *instance.PublicIpAddress == h.ElasticIP {
		return &SuccessResponse{
			ElasticIP:  h.ElasticIP,
			InstanceId: instanceID,
			Message:    "elastic ip already associated with instance id",
		}, nil
	}

	address, err := h.getAddressForIp(h.ElasticIP)
	if err != nil {
		return nil, fmt.Errorf("error fetching elastic ip address: %w", err)
	}

	assocRes, err := h.Ec2.AssociateAddress(&ec2.AssociateAddressInput{
		AllocationId: address.AllocationId,
		InstanceId:   aws.String(instanceID),
	})
	if err != nil {
		return nil, fmt.Errorf("error associating elastic ip address with %q %w", instanceID, err)
	}

	successMsg := fmt.Sprintf("Successfully allocated %s with instance %s.\n\tallocation id: %s, association id: %s\n",
		*address.PublicIp, instanceID, *address.AllocationId, *assocRes.AssociationId)

	return &SuccessResponse{
		ElasticIP:  h.ElasticIP,
		InstanceId: eventDetails.Ec2Instanceid,
		Message:    successMsg,
	}, nil
}

func (h *V1Handler) getInstanceById(id string) (*ec2.Instance, error) {
	instances, err := h.Ec2.DescribeInstances(&ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("instance-id"),
				Values: aws.StringSlice([]string{id}),
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("DescribeInstances failed: %w", err)
	}

	if instances == nil || len(instances.Reservations) == 0 || len(instances.Reservations[0].Instances) == 0 {
		return nil, errors.New("no instance found for the given id")
	}

	return instances.Reservations[0].Instances[0], nil
}

func (h *V1Handler) getAddressForIp(ip string) (*ec2.Address, error) {
	result, err := h.Ec2.DescribeAddresses(&ec2.DescribeAddressesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("domain"),
				Values: aws.StringSlice([]string{"vpc"}),
			},
		},
		PublicIps: []*string{aws.String(ip)},
	})
	if err != nil {
		return nil, err
	}

	if len(result.Addresses) == 0 {
		return nil, fmt.Errorf("elastic ip address not found in region %q", os.Getenv("AWS_REGION"))
	}

	return result.Addresses[0], nil
}
