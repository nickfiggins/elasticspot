package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/nickfiggins/elasticspot"
)

var handler elasticspot.HandleFunc

func init() {
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		panic(fmt.Sprintf("error loading default config: %v", err))
	}

	client := ec2.NewFromConfig(cfg)
	elasticIP := os.Getenv("ELASTIC_IP")
	if len(elasticIP) == 0 {
		panic("Elastic IP has not been set.")
	}
	handler = elasticspot.HandleV2(client, elasticIP)
}

func main() {
	lambda.Start(handler)
}
