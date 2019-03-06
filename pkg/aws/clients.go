package awsservices

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/service/ses/sesiface"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/secretsmanager/secretsmanageriface"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sns/snsiface"
	"github.com/aws/aws-xray-sdk-go/xray"
)

var (
	sess = session.Must(session.NewSession())
)

func makeConfig(serviceName string) *aws.Config {
	config := aws.NewConfig()
	if role, ok := os.LookupEnv(fmt.Sprintf("AWS_%s_ROLE_ARN", serviceName)); ok {
		config.WithCredentials(stscreds.NewCredentials(sess, role))
	}

	if endpoint, ok := os.LookupEnv(fmt.Sprintf("AWS_%s_ENDPOINT", serviceName)); ok {
		config.WithDisableSSL(true)
		config.WithEndpoint(endpoint)
	}

	return config
}

func NewDynamoDBClient() dynamodbiface.DynamoDBAPI {
	svc := dynamodb.New(sess, makeConfig("DYNAMODB"))
	xray.AWS(svc.Client)
	return svc
}

func NewSecretsManagerClient() secretsmanageriface.SecretsManagerAPI {
	svc := secretsmanager.New(sess, makeConfig("SECRETSMANAGER"))
	xray.AWS(svc.Client)
	return svc
}

func NewSNSClient() snsiface.SNSAPI {
	svc := sns.New(sess, makeConfig("SNS"))
	xray.AWS(svc.Client)
	return svc
}

func NewEC2Client() ec2iface.EC2API {
	svc := ec2.New(sess, makeConfig("EC2"))
	xray.AWS(svc.Client)
	return svc
}

func NewSESClient() sesiface.SESAPI {
	svc := ses.New(sess, makeConfig("SES"))
	xray.AWS(svc.Client)
	return svc
}
