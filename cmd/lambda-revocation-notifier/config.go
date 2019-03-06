package main

import (
	"github.com/kelseyhightower/envconfig"
)

const kConfigPrefix = "NOTIFIER"

var configNotifier struct {
	EmailFrom      string `split_words:"true" required:"true"`
	EmailSubject   string `split_words:"true" default:"VPN Certificate Expiration"`
	EmailSignature string `split_words:"true" default:"Your Friendly Ops Team"`
	EmailSourceArn string `split_words:"true"`

	AdminURL string `split_words:"true" required:"true"`
	HelpURL  string `split_words:"true" required:"true"`

	DaysBefore int `split_words:"true" default:"3"`
}

func init() {
	envconfig.MustProcess(kConfigPrefix, &configNotifier)
}
