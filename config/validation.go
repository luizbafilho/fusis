package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/logutils"

	"gopkg.in/go-playground/validator.v8"
)

var validate *validator.Validate

func init() {
	config := &validator.Config{TagName: "validate"}
	validate = validator.New(config)
}

//Validater validates a config
type Validater interface {
	Validate() error
}

func (config BalancerConfig) Validate() error {
	/* Validate LogLevel config */
	if err := validateLogLevel(config.LogLevel); err != nil {
		return err
	}

	/* Validate BGP config */
	if config.ClusterMode == "anycast" {
		if err := config.Bgp.Validate(); err != nil {
			return err
		}
	}

	/* Validate IPAM config */
	if len(config.Ipam.Ranges) > 0 {
		if err := config.Ipam.Validate(); err != nil {
			return err
		}
	}

	/* Validate Interfaces config */
	if err := config.Interfaces.Validate(); err != nil {
		return err
	}

	/* Validate Join nodes param */
	if !config.Bootstrap {
		if len(config.Join) == 0 {
			return fmt.Errorf("You need to specify join nodes or start in Bootstrap mode.")
		}

		for _, v := range config.Join {
			if err := validate.Field(v, "ip"); err != nil {
				return fmt.Errorf("Join parameter needs to be a valid IP v4")
			}
		}
	}

	return nil
}

func validateLogLevel(level string) error {
	err := errors.New("invalid log level")
	for _, l := range LOG_LEVELS {
		if l == logutils.LogLevel(strings.ToUpper(level)) {
			return nil
		}
	}
	return err
}

func (bgp Bgp) Validate() error {
	return validate.Struct(bgp)
}

func (ipam Ipam) Validate() error {
	return validate.Struct(ipam)
}

func (interfaces Interfaces) Validate() error {
	return validate.Struct(interfaces)
}
