package config

import (
	"fmt"

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

	/* Validate Join nodes param */
	if !config.Bootstrap {
		if len(config.Join) == 0 {
			return fmt.Errorf("You need to specify join nodes")
		}

		for _, v := range config.Join {
			if err := validate.Field(v, "ip"); err != nil {
				return fmt.Errorf("join parameter needs to be a valid IP v4")
			}
		}
	}

	return nil
}

func (bgp Bgp) Validate() error {
	return validate.Struct(bgp)
}

func (ipam Ipam) Validate() error {
	return validate.Struct(ipam)
}
