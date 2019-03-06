package main

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/empathybroker/aws-vpn/pkg/pki"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var (
	sess    = session.Must(session.NewSession())
	secrets = secretsmanager.New(sess)
)

const (
	kStepCreate = "createSecret"
	kStepSet    = "setSecret"
	kStepTest   = "testSecret"
	kStepFinish = "finishSecret"

	kStageCurrent = "AWSCURRENT"
	kStagePending = "AWSPENDING"

	kCAValidity = 90 * 24 * time.Hour
)

type SecretRotationEvent struct {
	ClientRequestToken string `json:"ClientRequestToken"`

	SecretId string `json:"SecretId"`
	Step     string `json:"Step"`
}

func init() {
	if os.Getenv("DEBUG") == "true" {
		log.SetLevel(log.DebugLevel)
	}
	log.SetFormatter(&log.JSONFormatter{
		TimestampFormat: time.RFC3339Nano,
		FieldMap: log.FieldMap{
			log.FieldKeyTime: "@timestamp",
		},
	})
}

func renewCAKey(ctx context.Context, secretId string, caName string, serialNumber string) (pki.CAData, error) {
	res, err := secrets.GetSecretValueWithContext(ctx, &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(secretId),
		VersionStage: aws.String(kStageCurrent),
	})
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == secretsmanager.ErrCodeResourceNotFoundException {
				log.Warnf("Secret is empty. Generating new CA Key")

				return pki.NewCAKey(caName, serialNumber, kCAValidity)
			}
		} else {
			return pki.CAData{}, errors.Wrap(err, "obtaining current key")
		}
	}

	var oldCA *pki.CAData
	if err := json.Unmarshal(res.SecretBinary, &oldCA); err != nil {
		log.WithError(err).Errorf("Error unmarshalling current key. New key will not be cross-signed")

		return pki.NewCAKey(caName, serialNumber, kCAValidity)
	}

	return oldCA.Renew(caName, serialNumber, kCAValidity)
}

func Handler(ctx context.Context, event SecretRotationEvent) error {
	log.WithFields(log.Fields{
		"token":  event.ClientRequestToken,
		"secret": event.SecretId,
		"step":   event.Step,
	}).Info("Doing rotation step")

	switch event.Step {
	case kStepCreate:
		caName := os.Getenv("PKI_CA_NAME")
		if caName == "" {
			return errors.New("Missing environment variable PKI_CA_NAME")
		}

		newCA, err := renewCAKey(ctx, event.SecretId, caName, event.ClientRequestToken)
		if err != nil {
			return errors.Wrap(err, "renewing CA key")
		}

		keyData, err := newCA.MarshalJSON()
		if err != nil {
			return errors.Wrap(err, "encoding new key")
		}

		res, err := secrets.PutSecretValueWithContext(ctx, &secretsmanager.PutSecretValueInput{
			ClientRequestToken: aws.String(event.ClientRequestToken),
			SecretId:           aws.String(event.SecretId),
			SecretBinary:       keyData,
			VersionStages:      aws.StringSlice([]string{kStagePending}),
		})
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				if awsErr.Code() == secretsmanager.ErrCodeResourceExistsException {
					log.Infof("Key already exists. Could be a retry. Skipping error")
					return nil
				}
			}
			return errors.Wrap(err, "writing new key")
		}

		log.Infof("Created new CA key with version ID %s", aws.StringValue(res.VersionId))
	case kStepSet:
		// Do nothing
	case kStepTest:
		// TODO
	case kStepFinish:
		secretInfo, err := secrets.DescribeSecretWithContext(ctx, &secretsmanager.DescribeSecretInput{
			SecretId: aws.String(event.SecretId),
		})
		if err != nil {
			return errors.Wrap(err, "obtaining secret details")
		}

		var currentVersion string
		for versionId, stages := range secretInfo.VersionIdsToStages {
			for _, stage := range aws.StringValueSlice(stages) {
				if stage == kStageCurrent {
					currentVersion = versionId
				}
			}
		}

		_, err = secrets.UpdateSecretVersionStageWithContext(ctx, &secretsmanager.UpdateSecretVersionStageInput{
			SecretId:            aws.String(event.SecretId),
			RemoveFromVersionId: aws.String(currentVersion),
			MoveToVersionId:     aws.String(event.ClientRequestToken),
			VersionStage:        aws.String(kStageCurrent),
		})
		if err != nil {
			return errors.Wrap(err, "error updating secret stage")
		}

		log.Info("CA rekeying finished")
	default:
		log.Errorf("Unknown step: %s", event.Step)
	}

	return nil
}

func main() {
	lambda.Start(Handler)
}
