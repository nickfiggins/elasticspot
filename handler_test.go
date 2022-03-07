package elasticspot_test

import (
	"context"
	"errors"
	"testing"

	"github.com/nickfiggins/elasticspot"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/stretchr/testify/assert"
)

var elasticIp = "192.0.0.1"
var testAllocationId = "<test allocation id>"
var testAssociationId = "<test association id>"
var testInstanceId = "instance-id"

func TestHandler(t *testing.T) {
	cases := []struct {
		scenario string
		request  *events.CloudWatchEvent
		ec2 *mockEc2
		message string
		err error
	}{
		{
			scenario: "happy path",
			request: &events.CloudWatchEvent{Detail: []byte(`{"ec2InstanceId": "instance-id"}`)},
			ec2: &mockEc2{
				describeAddrOutput: &ec2.DescribeAddressesOutput{
					Addresses: []*ec2.Address{{
						PublicIp: aws.String(elasticIp),
						AllocationId: aws.String(testAllocationId),
					}},
				},
				associateAddrOutput: &ec2.AssociateAddressOutput{
					AssociationId: aws.String(testAssociationId),
				},
				describeInstancesOutput: &ec2.DescribeInstancesOutput{
					Reservations: []*ec2.Reservation{{
						Instances: []*ec2.Instance{{
							InstanceId: aws.String(testInstanceId),
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
			request: &events.CloudWatchEvent{Detail: []byte(`{"ec2InstanceId": "instance-id"}`)},
			ec2: &mockEc2{
				describeAddrOutput: &ec2.DescribeAddressesOutput{
					Addresses: []*ec2.Address{{
						PublicIp: aws.String(elasticIp),
						AllocationId: aws.String(testAllocationId),
					}},
				},
				associateAddrOutput: &ec2.AssociateAddressOutput{
					AssociationId: aws.String(testAssociationId),
				},
				describeInstancesOutput: &ec2.DescribeInstancesOutput{
					Reservations: []*ec2.Reservation{{
						Instances: []*ec2.Instance{{
							InstanceId: aws.String(testInstanceId),
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
			request: &events.CloudWatchEvent{Detail: []byte(`{"ec2InstanceId": "instance-id"}`)},
			ec2: &mockEc2{},
			err: errors.New("no instance found for the given id"),
		},
	}

	for _, c := range cases {
		t.Run(c.scenario, func(t *testing.T) {
			h := elasticspot.NewHandler(c.ec2, elasticIp)
			r, err := h.Handle(context.Background(), c.request)
			if c.err != nil {
				assert.EqualError(t, err, c.err.Error())
			}
			if err == nil {
				assert.NotNil(t, r)
				assert.Equal(t, c.message, r.Message)
			}
		})
	}
}



type mockEc2 struct {
	describeAddrOutput *ec2.DescribeAddressesOutput
	associateAddrOutput *ec2.AssociateAddressOutput
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