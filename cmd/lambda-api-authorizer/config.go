package main

import "github.com/kelseyhightower/envconfig"

const kConfigPrefix = "AUTH"

var configApiAuthorizer struct {
	Audience      string   `split_words:"true" required:"true"`
	HostedDomains []string `split_words:"true" required:"true"`
}

func init() {
	envconfig.MustProcess(kConfigPrefix, &configApiAuthorizer)
}
