package awspki

import (
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/aws/aws-sdk-go/service/secretsmanager/secretsmanageriface"
	"github.com/empathybroker/aws-vpn/pkg/pki"
)

type awsStorage struct {
	sm  secretsmanageriface.SecretsManagerAPI
	ddb dynamodbiface.DynamoDBAPI

	data pki.CAData
	mut  sync.Mutex
	exp  time.Time
}

func NewAWSStorage(sm secretsmanageriface.SecretsManagerAPI, ddb dynamodbiface.DynamoDBAPI) *awsStorage {
	return &awsStorage{
		sm:  sm,
		ddb: ddb,
	}
}
