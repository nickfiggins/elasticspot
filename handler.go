package elasticspot

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
)

// Handler is a Lambda handler that associates an Elastic IP with an EC2 instance.
// For non-lambda environments, the Client may be used directly instead.
type Handler struct {
	client    associator
	elasticIP string
}

type associator interface {
	AssociateIP(ctx context.Context, ip string, instanceID string) (*AssociateResponse, error)
}

func NewV1Handler(ec2 EC2API, elasticIP string) *Handler {
	return &Handler{client: &Client{EC2API: ec2}, elasticIP: elasticIP}
}

type HandleFunc func(ctx context.Context, event *events.CloudWatchEvent) (*HandleResponse, error)

// HandleV1 is a convenience function for creating an ElasticSpot lambda handler for v1 of the
// AWS SDK.
func HandleV1(ec2 EC2API, elasticIp string) HandleFunc {
	h := NewV1Handler(ec2, elasticIp)
	return h.Handle
}

// HandleV2 is a convenience function for creating an ElasticSpot lambda handler for v2 of the
// AWS SDK.
func HandleV2(ec2 EC2APIV2, elasticIp string) HandleFunc {
	h := NewV2Handler(ec2, elasticIp)
	return h.Handle
}

func NewV2Handler(ec2v2 EC2APIV2, elasticIP string) *Handler {
	return &Handler{client: &ClientV2{EC2API: ec2v2}, elasticIP: elasticIP}
}

type HandleResponse struct {
	InstanceId string `json:"instanceID,omitempty"`
	ElasticIP  string `json:"elasticIP,omitempty"`
	Message    string `json:"message,omitempty"`
}

// Handle retrieves the EC2 instance ID from the CloudWatch event and associates the configured Elastic IP
// with the instance. If the Elastic IP is already associated with the instance, the handler will return
// a successful response with a message indicating that the Elastic IP is already associated with the instance.
func (h *Handler) Handle(ctx context.Context, event *events.CloudWatchEvent) (*HandleResponse, error) {
	var eventDetails EventDetails
	if err := json.Unmarshal(event.Detail, &eventDetails); err != nil {
		return nil, fmt.Errorf("error unmarshaling cloudwatch event: %w", err)
	}

	instanceID := eventDetails.Ec2Instanceid

	res, err := h.client.AssociateIP(ctx, h.elasticIP, instanceID)
	if err != nil {
		return nil, err
	}

	if res.AlreadyAssociated {
		return &HandleResponse{
			ElasticIP:  h.elasticIP,
			InstanceId: instanceID,
			Message:    "elastic ip already associated with instance id",
		}, nil
	}

	successMsg := fmt.Sprintf(
		`Successfully allocated %s with instance %s. Allocation ID: %s Association ID: %s`,
		h.elasticIP, instanceID, res.AllocationID, res.AssociationID,
	)

	return &HandleResponse{
		ElasticIP:  h.elasticIP,
		InstanceId: eventDetails.Ec2Instanceid,
		Message:    successMsg,
	}, nil
}
