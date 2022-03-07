package main

import (
	"context"
	"os"


	"github.com/nickfiggins/elasticspot"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func handler(ctx context.Context, event *events.CloudWatchEvent) (*elasticspot.SuccessResponse, error) {
	return elasticspot.NewHandler(ec2_sess, elasticIp).Handle(ctx, event)
}

var ec2_sess *ec2.EC2
var elasticIp string

func initSessions() {
	sess, err := session.NewSession(&aws.Config{
        Region: aws.String(os.Getenv("AWS_REGION"))},
	); if err != nil {
		panic(err)
	}
	ec2_sess = ec2.New(sess)
}

func init() {
	initSessions()
	elasticIp = os.Getenv("ELASTIC_IP")
	if len(elasticIp) == 0 {
		panic("Elastic IP has not been set.")
	}
}

func main() {
	lambda.Start(handler)
}