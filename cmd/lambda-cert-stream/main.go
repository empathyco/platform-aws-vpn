package main

import (
	"context"
	"crypto/x509"
	"encoding/hex"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	awsservices "github.com/empathybroker/aws-vpn/pkg/aws"
	"github.com/empathybroker/aws-vpn/pkg/pki"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	kAttrSerialNumber   = "SerialNumber"
	kAttrAuthorityKeyId = "AuthorityKeyId"
	kAttrSubjectKeyId   = "SubjectKeyId"
	kAttrSubjectName    = "SubjectName"
	kAttrCertType       = "CertType"
	kAttrIssuedAt       = "IssuedAt"
	kAttrValidUntil     = "ValidUntil"
	kAttrRevocationTime = "RevocationTime"
	kAttrData           = "Data"
)

var (
	snsClient = awsservices.NewSNSClient()
)

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

func parseCertificate(data map[string]events.DynamoDBAttributeValue) (info pki.CertificateInfo, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.Errorf("Panic parsing certificate: %s", r)
		}
	}()

	for k, v := range data {
		switch k {
		case kAttrData:
			cert, err := x509.ParseCertificate(v.Binary())
			if err != nil {
				return info, err
			}
			info.Certificate = cert
		case kAttrCertType:
			info.CertType = pki.CertType(v.String())
		case kAttrSerialNumber:
			info.SerialBytes = v.Binary()
			info.Serial = hex.EncodeToString(info.SerialBytes)
		case kAttrSubjectKeyId:
			info.KeyId = hex.EncodeToString(v.Binary())
		case kAttrSubjectName:
			info.Subject = v.String()
		case kAttrIssuedAt:
			intv, err := v.Integer()
			if err != nil {
				return info, err
			}
			info.NotBefore = time.Unix(intv, 0).UTC()
		case kAttrValidUntil:
			intv, err := v.Integer()
			if err != nil {
				return info, err
			}
			info.NotAfter = time.Unix(intv, 0).UTC()
		case kAttrRevocationTime:
			intv, err := v.Integer()
			if err != nil {
				return info, err
			}
			if intv > 0 {
				rt := time.Unix(intv, 0).UTC()
				info.Revoked = &rt
			}
		}
	}

	return info, nil
}

func Handler(ctx context.Context, event events.DynamoDBEvent) error {
	for _, record := range event.Records {
		event := make(map[string]interface{})

		switch record.EventName {
		case string(events.DynamoDBOperationTypeInsert):
			event["event"] = "cert_store_new"
		case string(events.DynamoDBOperationTypeModify):
			event["event"] = "cert_store_updated"
		case string(events.DynamoDBOperationTypeRemove):
			event["event"] = "cert_store_deleted"
		default:
			log.Errorf("Unknown operation type: %s", record.EventName)
			continue
		}

		keySerial, ok := record.Change.Keys[kAttrSerialNumber]
		if !ok || keySerial.DataType() != events.DataTypeBinary {
			log.Errorf("Missing %s key", kAttrSerialNumber)
			continue
		}
		event["serial"] = hex.EncodeToString(keySerial.Binary())

		if len(record.Change.NewImage) > 0 {
			info, err := parseCertificate(record.Change.NewImage)
			if err != nil {
				log.WithError(err).Errorf("Error parsing new DynamoDB image")
				continue
			}
			event["cert"] = info
		}

		if len(record.Change.OldImage) > 0 {
			info, err := parseCertificate(record.Change.OldImage)
			if err != nil {
				log.WithError(err).Errorf("Error parsing old DynamoDB image")
				continue
			}
			event["cert_prev"] = info
		}

		if err := awsservices.PublishEvent(snsClient, ctx, event); err != nil {
			log.WithError(err).Errorf("Error publishing event")
			continue
		}
	}

	return nil
}

func main() {
	lambda.Start(Handler)
}
