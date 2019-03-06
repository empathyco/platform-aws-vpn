package main

import (
	"bytes"
	"context"
	"io"
	"os"
	"sort"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"
	awsservices "github.com/empathybroker/aws-vpn/pkg/aws"
	"github.com/empathybroker/aws-vpn/pkg/pki"
	awspki "github.com/empathybroker/aws-vpn/pkg/pki/aws"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
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

var (
	sesClient  = awsservices.NewSESClient()
	pkiStorage = awspki.NewAWSStorage(awsservices.NewSecretsManagerClient(), awsservices.NewDynamoDBClient())
	awsPKI     = pki.NewPKI(pkiStorage)
)

func getEmailBody(w io.Writer, certs []*pki.CertificateInfo) error {
	sort.Slice(certs, func(i, j int) bool {
		return certs[i].NotAfter.Before(certs[j].NotAfter)
	})

	return tplEmail.Execute(w, bodyData{
		Certificates: certs,

		AdminURL:  configNotifier.AdminURL,
		HelpURL:   configNotifier.HelpURL,
		Signature: configNotifier.EmailSignature,
	})
}

func sendNotificationEmail(ctx context.Context, to string, body string) error {
	sesInput := &ses.SendEmailInput{
		Source: aws.String(configNotifier.EmailFrom),
		Destination: &ses.Destination{
			ToAddresses: aws.StringSlice([]string{to}),
		},
		Message: &ses.Message{
			Subject: &ses.Content{
				Data:    aws.String(configNotifier.EmailSubject),
				Charset: aws.String("UTF-8"),
			},
			Body: &ses.Body{
				Text: &ses.Content{
					Data:    aws.String(body),
					Charset: aws.String("UTF-8"),
				},
			},
		},
	}

	if configNotifier.EmailSourceArn != "" {
		sesInput.SourceArn = aws.String(configNotifier.EmailSourceArn)
	}

	r, err := sesClient.SendEmailWithContext(ctx, sesInput)
	if err != nil {
		return errors.Wrap(err, "sending message with SES")
	}

	log.Infof("Sent message to %s with ID %s", to, aws.StringValue(r.MessageId))
	return nil
}

func handler(ctx context.Context) error {
	certs, err := awsPKI.ListCerts(ctx, "")
	if err != nil {
		log.WithError(err).Error("Error listing certificates")
		return err
	}

	cutoffTime := time.Now().Add(time.Hour * 24 * time.Duration(configNotifier.DaysBefore)).UTC()
	log.Debugf("Sending notifications to users with certs expiring before %s", cutoffTime.Format(time.RFC3339))

	targets := make(map[string][]*pki.CertificateInfo)
	for _, cert := range certs {
		if cert.Revoked != nil {
			continue
		}

		if cert.NotAfter.Before(cutoffTime) {
			targets[cert.Subject] = append(targets[cert.Subject], cert)
		}
	}

	for to, certs := range targets {
		var body bytes.Buffer
		if err := getEmailBody(&body, certs); err != nil {
			log.WithError(err).Error("Error building email body from template")
			continue
		}

		if err := sendNotificationEmail(ctx, to, body.String()); err != nil {
			log.WithError(err).Error("Error sending email")
			continue
		}
	}

	log.Infof("Sent %d notification emails", len(targets))
	return nil
}

func main() {
	lambda.Start(handler)
}
