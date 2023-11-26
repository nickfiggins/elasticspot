package elasticspot

import (
	"context"
	"errors"
	"fmt"
	"os"

	ec2v2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go/aws"
)

type AssociateResponse struct {
	AlreadyAssociated bool
	AllocationID      string
	AssociationID     string
}

// ClientV2 will create a new client for version 2 of the AWS SDK.
type ClientV2 struct {
	EC2API EC2APIV2
}

type EC2APIV2 interface {
	DescribeAddresses(ctx context.Context, params *ec2v2.DescribeAddressesInput, optFns ...func(*ec2v2.Options)) (*ec2v2.DescribeAddressesOutput, error)
	AssociateAddress(ctx context.Context, params *ec2v2.AssociateAddressInput, optFns ...func(*ec2v2.Options)) (*ec2v2.AssociateAddressOutput, error)
	DescribeInstances(ctx context.Context, params *ec2v2.DescribeInstancesInput, optFns ...func(*ec2v2.Options)) (*ec2v2.DescribeInstancesOutput, error)
}

func (a *ClientV2) AssociateIP(ctx context.Context, ip string, instanceID string) (*AssociateResponse, error) {
	instance, err := a.getInstanceById(ctx, instanceID)
	if err != nil {
		return nil, fmt.Errorf("error fetching instance: %w", err)
	}

	if *instance.PublicIpAddress == ip {
		return &AssociateResponse{AlreadyAssociated: true}, nil
	}

	address, err := a.getAddressForIp(ctx, ip)
	if err != nil {
		return nil, fmt.Errorf("error fetching elastic ip address: %w", err)
	}

	assocRes, err := a.EC2API.AssociateAddress(ctx, &ec2v2.AssociateAddressInput{
		AllocationId: address.AllocationId,
		InstanceId:   aws.String(instanceID),
	})
	if err != nil {
		return nil, fmt.Errorf("error associating elastic ip address with %q %w", instanceID, err)
	}

	return &AssociateResponse{
		AllocationID:  *address.AllocationId,
		AssociationID: *assocRes.AssociationId,
	}, nil
}

func (a *ClientV2) getInstanceById(ctx context.Context, id string) (*types.Instance, error) {
	instances, err := a.EC2API.DescribeInstances(ctx, &ec2v2.DescribeInstancesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("instance-id"),
				Values: []string{id},
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("DescribeInstances failed: %w", err)
	}

	if instances == nil || len(instances.Reservations) == 0 || len(instances.Reservations[0].Instances) == 0 {
		return nil, errors.New("no instance found for the given id")
	}

	return &instances.Reservations[0].Instances[0], nil
}

func (c *ClientV2) getAddressForIp(ctx context.Context, ip string) (*types.Address, error) {
	result, err := c.EC2API.DescribeAddresses(ctx, &ec2v2.DescribeAddressesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("domain"),
				Values: []string{"vpc"},
			},
		},
		PublicIps: []string{ip},
	})
	if err != nil {
		return nil, err
	}

	if len(result.Addresses) == 0 {
		return nil, fmt.Errorf("elastic ip address not found in region %q", os.Getenv("AWS_REGION"))
	}

	return &result.Addresses[0], nil
}
