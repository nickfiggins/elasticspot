package elasticspot_test

import (
	"context"
	"testing"

	"github.com/nickfiggins/elasticspot"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

var elasticIp = "192.0.0.1"
var testAllocationId = "<test allocation id>"
var testAssociationId = "<test association id>"
var testInstanceId = "instance-id"

func TestHandler(t *testing.T) {
	tests := []struct {
		scenario string
		request  *events.CloudWatchEvent
		ec2      *mockEc2
		message  string
		wantErr  bool
	}{
		{
			scenario: "happy path",
			request:  &events.CloudWatchEvent{Detail: []byte(`{"ec2InstanceId": "instance-id"}`)},
			ec2: &mockEc2{
				describeAddrOutput: &ec2.DescribeAddressesOutput{
					Addresses: []*ec2.Address{{
						PublicIp:     aws.String(elasticIp),
						AllocationId: aws.String(testAllocationId),
					}},
				},
				associateAddrOutput: &ec2.AssociateAddressOutput{
					AssociationId: aws.String(testAssociationId),
				},
				describeInstancesOutput: &ec2.DescribeInstancesOutput{
					Reservations: []*ec2.Reservation{{
						Instances: []*ec2.Instance{{
							InstanceId:      aws.String(testInstanceId),
							PublicIpAddress: aws.String("192.0.0.0"),
						}},
					},
					},
				},
			},
			message: "Successfully allocated " + elasticIp + " with instance instance-id.\n\t" +
				"allocation id: <test allocation id>, association id: <test association id>\n",
		},
		{
			scenario: "elastic ip already assigned to instance",
			request:  &events.CloudWatchEvent{Detail: []byte(`{"ec2InstanceId": "instance-id"}`)},
			ec2: &mockEc2{
				describeAddrOutput: &ec2.DescribeAddressesOutput{
					Addresses: []*ec2.Address{{
						PublicIp:     aws.String(elasticIp),
						AllocationId: aws.String(testAllocationId),
					}},
				},
				associateAddrOutput: &ec2.AssociateAddressOutput{
					AssociationId: aws.String(testAssociationId),
				},
				describeInstancesOutput: &ec2.DescribeInstancesOutput{
					Reservations: []*ec2.Reservation{{
						Instances: []*ec2.Instance{{
							InstanceId:      aws.String(testInstanceId),
							PublicIpAddress: aws.String(elasticIp),
						}},
					},
					},
				},
			},
			message: "elastic ip already associated with instance id",
		},
		{
			scenario: "no instance found",
			request:  &events.CloudWatchEvent{Detail: []byte(`{"ec2InstanceId": "instance-id"}`)},
			ec2:      &mockEc2{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.scenario, func(t *testing.T) {
			h := elasticspot.NewV1Handler(tt.ec2, elasticIp)
			got, err := h.Handle(context.Background(), tt.request)
			if tt.wantErr != (err != nil) {
				t.Errorf("wantErr: %v, gotErr: %v", tt.wantErr, err)
				return
			}
			if err == nil && got.Message != tt.message {
				t.Errorf("want message: %s, got: %s", tt.message, got.Message)
			}
		})
	}
}

type mockEc2 struct {
	describeAddrOutput      *ec2.DescribeAddressesOutput
	associateAddrOutput     *ec2.AssociateAddressOutput
	describeInstancesOutput *ec2.DescribeInstancesOutput
}

func (m *mockEc2) DescribeAddresses(input *ec2.DescribeAddressesInput) (*ec2.DescribeAddressesOutput, error) {
	return m.describeAddrOutput, nil
}

func (m *mockEc2) AssociateAddress(input *ec2.AssociateAddressInput) (*ec2.AssociateAddressOutput, error) {
	return m.associateAddrOutput, nil
}

func (m *mockEc2) DescribeInstances(input *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error) {
	return m.describeInstancesOutput, nil
}
