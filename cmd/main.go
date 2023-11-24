package main

import (
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/nickfiggins/elasticspot"
)

var handler *elasticspot.V1Handler

func init() {
	cfg := &aws.Config{Region: aws.String(os.Getenv("AWS_REGION"))}
	sess, err := session.NewSession(cfg)
	if err != nil {
		panic(fmt.Sprintf("error creating aws session: %v", err))
	}
	ec2Session := ec2.New(sess)
	elasticIP := os.Getenv("ELASTIC_IP")
	if len(elasticIP) == 0 {
		panic("Elastic IP has not been set.")
	}
	handler = elasticspot.NewV1Handler(ec2Session, elasticIP)
}

func main() {
	lambda.Start(handler.Handle)
}
