package awsservices

import (
	"context"
	"encoding/json"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sns/snsiface"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func PublishEvent(snsClient snsiface.SNSAPI, ctx context.Context, message map[string]interface{}) error {
	if topicARN, ok := os.LookupEnv("SNS_TOPIC_ARN"); ok {
		jsonMsg, err := json.Marshal(message)
		if err != nil {
			return errors.Wrapf(err, "marshalling event")
		}

		r, err := snsClient.PublishWithContext(ctx, &sns.PublishInput{
			TopicArn: aws.String(topicARN),
			Subject:  aws.String("VPN Event"),
			Message:  aws.String(string(jsonMsg)),
		})
		if err != nil {
			return errors.Wrap(err, "sending event")
		}

		log.Debugf("Published event with ID: %s", aws.StringValue(r.MessageId))
	}

	return nil
}
