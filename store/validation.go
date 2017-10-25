package store

import (
	"fmt"
	"strings"

	"github.com/luizbafilho/fusis/types"
	validator "gopkg.in/go-playground/validator.v9"
)

// import (
// 	"fmt"
// 	"strings"
//
// 	"github.com/luizbafilho/fusis/types"
// 	"github.com/pkg/errors"
// 	validator "gopkg.in/go-playground/validator.v9"
// )
//
// // Validating service name uniqueness
// func (s *FusisStore) validateServiceNameUniqueness(svcKey string) error {
// 	exists, err := s.kv.Exists(svcKey)
// 	if err != nil {
// 		return errors.Wrap(err, "verifying service uniqueness failed")
// 	}
// 	if exists {
// 		return types.ErrValidation{Type: "service", Errors: map[string]string{"Name": "field must be unique"}}
// 	}
//
// 	return nil
// }
//
// // Validating IPVS uniqueness.
// func (s *FusisStore) validateServiceIpvsUniqueness(ipvsKey string) error {
// 	exists, err := s.kv.Exists(ipvsKey)
// 	if err != nil {
// 		return errors.Wrap(err, "verifying ipvs service existence failed")
// 	}
// 	if exists {
// 		return types.ErrValidation{Type: "service", Errors: map[string]string{"ipvs": "address, port and protocol belongs to another service. It must be unique."}}
// 	}
//
// 	return nil
// }
//
// // Validating destination name uniqueness
// func (s *FusisStore) validateDestinationNameUniqueness(dstKey string) error {
// 	exists, err := s.kv.Exists(dstKey)
// 	if err != nil {
// 		return errors.Wrap(err, "verifying destination uniqueness failed")
// 	}
// 	if exists {
// 		return types.ErrValidation{Type: "destination", Errors: map[string]string{"Name": "field must be unique"}}
// 	}
//
// 	return nil
// }
//
// // Validating destination IPVS uniqueness.
// func (s *FusisStore) validateDestinationIpvsUniqueness(ipvsKey string) error {
// 	exists, err := s.kv.Exists(ipvsKey)
// 	if err != nil {
// 		return errors.Wrap(err, "verifying ipvs destination existence failed")
// 	}
// 	if exists {
// 		return types.ErrValidation{Type: "destination", Errors: map[string]string{"ipvs": "address and port belongs to another destination. It must be unique."}}
// 	}
//
// 	return nil
// }
//
// func (s *FusisStore) validateDestination(dst *types.Destination) error {
// 	if err := s.validate.Struct(dst); err != nil {
// 		errValidation := types.ErrValidation{Type: "destination", Errors: make(map[string]string)}
// 		for _, err := range err.(validator.ValidationErrors) {
// 			errValidation.Errors[err.Field()] = getValidationMessage(err)
// 		}
// 		return errValidation
// 	}
//
// 	return nil
// }
//
func (s *FusisStore) validateService(svc *types.Service) error {
	if err := s.validate.Struct(svc); err != nil {
		errValidation := types.ErrValidation{Type: "service", Errors: make(map[string]string)}
		for _, err := range err.(validator.ValidationErrors) {
			errValidation.Errors[err.Field()] = getValidationMessage(err)
		}
		return errValidation
	}

	return nil
}

func validateValues(values []string) validator.Func {
	return func(fl validator.FieldLevel) bool {
		str := fl.Field().String()

		for _, v := range values {
			if v == str {
				return true
			}
		}

		return false
	}
}

func getValidationMessage(fieldError validator.FieldError) string {
	switch fieldError.Tag() {
	case "required":
		return "field is required"
	case "lte":
		return fmt.Sprintf("field field must be less than %s", fieldError.Param())
	case "gte":
		return fmt.Sprintf("field field must be greater than %s", fieldError.Param())
	case "protocols":
		return fmt.Sprintf("field must be one of the following: %s", strings.Join(types.Protocols, " | "))
	case "schedulers":
		return fmt.Sprintf("field must be one of the following: %s", strings.Join(types.Schedulers, " | "))
	}

	return "unknown validation error"
}
