package main

import (
	"errors"

	"github.com/namsral/flag"
)

const (
	envDevelopment = "development"
	envTesting     = "testing"
	envStaging     = "staging"
	envProduction  = "production"
)

type config struct {
	env string
}

func (c *config) isDevelopment() bool {
	return c.env == envDevelopment
}

// parse parses configuration from flags or environment variables.
func (c *config) parse() error {
	flag.StringVar(&c.env, "env", "", "Default environment")
	flag.Parse()

	if c.env == "" {
		return errors.New("env is not set")
	}

	return nil
}
