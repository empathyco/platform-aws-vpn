package awspki

import (
	"github.com/kelseyhightower/envconfig"
)

const kConfigPrefix = "PKI_AWS"

var configAWSPKI struct {
	SecretName   string `split_words:"true" default:"VPN/CAPrivateKey"`
	TableName    string `split_words:"true" default:"vpn_certificates"`
	DurationDays int    `split_words:"true" default:"30"`
}

func init() {
	envconfig.MustProcess(kConfigPrefix, &configAWSPKI)
}
