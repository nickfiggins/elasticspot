package elasticspot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type Handler struct {
	Ec2 EC2API
	ElasticIP string
}

func NewHandler(ec2 EC2API, elasticIp string) *Handler {
	return &Handler{Ec2: ec2, ElasticIP: elasticIp}
}

func (h *Handler) Handle(ctx context.Context, event *events.CloudWatchEvent) (*SuccessResponse, error) {
	var eventDetails EventDetails
	response := &SuccessResponse{ElasticIP: h.ElasticIP}
	json.Unmarshal(event.Detail, &eventDetails)

	instance, err := h.getInstanceById(eventDetails.Ec2Instanceid); if err != nil {
		log.Println("Error fetching instance: ", err)
		return nil, err
	}

	response.InstanceId = eventDetails.Ec2Instanceid

	if *instance.PublicIpAddress == h.ElasticIP {
		log.Println("elastic ip already associated with instance id", eventDetails.Ec2Instanceid)
		response.Message = "elastic ip already associated with instance id"
		return response, nil
	}
	
	address, err := h.getAddressForIp(h.ElasticIP); if err != nil {
		log.Println("error fetching elastic ip address", err)
		return nil, err
	}
	

	instanceID := eventDetails.Ec2Instanceid
	ip := address.PublicIp
	allocationID := address.AllocationId

	assocRes, err := h.Ec2.AssociateAddress(&ec2.AssociateAddressInput{
        AllocationId: allocationID,
        InstanceId:   aws.String(instanceID),
    }); if err != nil {
		log.Printf("Unable to associate IP address with %s, %v", instanceID, err)
		return nil, err
    }

	successMsg := fmt.Sprintf("Successfully allocated %s with instance %s.\n\tallocation id: %s, association id: %s\n",
	*ip, instanceID, *allocationID, *assocRes.AssociationId)
    log.Println(successMsg)

	response.Message = successMsg

	return response, nil
}

func (h *Handler) getInstanceById(id string) (*ec2.Instance, error) {
	instances, err := h.Ec2.DescribeInstances(&ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("instance-id"),
				Values: aws.StringSlice([]string{id}),
			},
		},
	}); if err != nil {
		return nil, err
	}

	if instances == nil || len(instances.Reservations) == 0 || len(instances.Reservations[0].Instances) == 0{
		return nil, errors.New("no instance found for the given id")
	}

	return instances.Reservations[0].Instances[0], nil
}

func (h *Handler) getAddressForIp(ip string) (*ec2.Address, error) {
	result, err := h.Ec2.DescribeAddresses(&ec2.DescribeAddressesInput{
        Filters: []*ec2.Filter{
            {
                Name:   aws.String("domain"),
                Values: aws.StringSlice([]string{"vpc"}),
            },
        },
		PublicIps: []*string{aws.String(ip)},
    }); if err != nil {
		return nil, err
	}

	if len(result.Addresses) == 0 {
		log.Printf("No elastic IPs for %s region\n", os.Getenv("AWS_REGION")) // TODO: fix
		return nil, errors.New("elastic ip address not found")
	}

	return result.Addresses[0], nil
}

type EC2API interface {
	DescribeAddresses(input *ec2.DescribeAddressesInput) (*ec2.DescribeAddressesOutput, error)
	AssociateAddress(input *ec2.AssociateAddressInput) (*ec2.AssociateAddressOutput, error)
	DescribeInstances(input *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error)
}