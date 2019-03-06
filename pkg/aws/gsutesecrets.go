package awsservices

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/secretsmanager/secretsmanageriface"
)

type awsServiceAccountProvider struct {
	secretsManager secretsmanageriface.SecretsManagerAPI
	secretId       string
}

func NewAWSServiceAccountProvider(secretsManager secretsmanageriface.SecretsManagerAPI, secretId string) *awsServiceAccountProvider {
	return &awsServiceAccountProvider{
		secretsManager: secretsManager,
		secretId:       secretId,
	}
}

func (p awsServiceAccountProvider) GetKey(ctx context.Context) ([]byte, error) {
	res, err := p.secretsManager.GetSecretValueWithContext(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(p.secretId),
	})
	if err != nil {
		return nil, err
	}

	return res.SecretBinary, nil
}
