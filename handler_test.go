package elasticspot_test

import (
	"context"
	"testing"

	"github.com/nickfiggins/elasticspot"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/stretchr/testify/assert"
)

func TestHandler(t *testing.T) {
	cases := []struct {
		scenario string
		request  *events.CloudWatchEvent
		ec2 *mockEc2
		message string
	}{
		{
			scenario: "happy path",
			request: &events.CloudWatchEvent{Detail: []byte(`{"ec2InstanceId": "instance-id"}`)},
			ec2: &mockEc2{
				describeAddrOutput: &ec2.DescribeAddressesOutput{
					Addresses: []*ec2.Address{{
						PublicIp: aws.String("192.168.0.0"),
						AllocationId: aws.String("<test allocation id>"),
					}},
				},
				associateAddrOutput: &ec2.AssociateAddressOutput{
					AssociationId: aws.String("<test association id>"),
				},
				describeInstancesOutput: &ec2.DescribeInstancesOutput{
					Reservations: []*ec2.Reservation{{
						Instances: []*ec2.Instance{{
							InstanceId: aws.String("test instance id"),
							PublicIpAddress: aws.String("192.0.0.0"),
						}},
					},
					},
					},
			},
			message: "Successfully allocated 192.168.0.0 with instance instance-id. " + 
			"allocation id: <test allocation id>, association id: <test association id>",
		},
	}

	for _, c := range cases {
		t.Run(c.scenario, func(t *testing.T) {
			h := elasticspot.NewHandler(c.ec2, "192.0.0.1")
			r, err := h.Handle(context.Background(), c.request)
			assert.NoError(t, err)
			if c.ec2.err == nil {
				assert.NotNil(t, r)
			} else if r != nil {
				assert.Equal(t, c.message, r.Message)
			}
		})
	}
}



type mockEc2 struct {
	err error
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