package elasticspot

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// Client will create a new client for version 1 of the AWS SDK.
type Client struct {
	EC2API EC2API
}

type EC2API interface {
	DescribeAddresses(input *ec2.DescribeAddressesInput) (*ec2.DescribeAddressesOutput, error)
	AssociateAddress(input *ec2.AssociateAddressInput) (*ec2.AssociateAddressOutput, error)
	DescribeInstances(input *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error)
}

func (c *Client) AssociateIP(ctx context.Context, ip, instanceID string) (*AssociateResponse, error) {
	instance, err := c.getInstanceById(instanceID)
	if err != nil {
		return nil, fmt.Errorf("error fetching instance: %w", err)
	}

	if *instance.PublicIpAddress == ip {
		return &AssociateResponse{AlreadyAssociated: true}, nil
	}

	address, err := c.getAddressForIp(ip)
	if err != nil {
		return nil, fmt.Errorf("error fetching elastic ip address: %w", err)
	}

	assocRes, err := c.EC2API.AssociateAddress(&ec2.AssociateAddressInput{
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

func (a *Client) getInstanceById(id string) (*ec2.Instance, error) {
	instances, err := a.EC2API.DescribeInstances(&ec2.DescribeInstancesInput{
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

func (a *Client) getAddressForIp(ip string) (*ec2.Address, error) {
	result, err := a.EC2API.DescribeAddresses(&ec2.DescribeAddressesInput{
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
