package v1

import (
	"encoding/json"
	"errors"
	"github.com/go-playground/validator/v10"
	"go-deploy/models/dto/body"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/sys"
	"go-deploy/service"
	"net"
	"reflect"
)

func WithAuth(context *sys.ClientContext) (*service.AuthInfo, error) {
	token, err := context.GetKeycloakToken()
	if err != nil {
		return nil, err
	}

	return service.CreateAuthInfo(token.Sub, token, token.Groups), nil
}

func msgForTag(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Must be a valid email address"
	case "min":
		switch fe.Kind() {
		case reflect.String:
			return "Must be at least " + fe.Param() + " characters long"
		case reflect.Int:
			return "Must be at least " + fe.Param()
		default:
			return "Must be at least " + fe.Param()
		}
	case "max":
		switch fe.Kind() {
		case reflect.String:
			return "Must be at most " + fe.Param() + " characters long"
		case reflect.Int:
			return "Must be at most " + fe.Param()
		default:
			return "Must be at most " + fe.Param()
		}
	case "boolean":
		return "Must be true or false"
	case "alphanum":
		return "Must be alphanumeric"
	case "uuid4":
		return "Must be a valid UUIDv4"
	case "rfc1035":
		return "Must be a valid hostname (RFC 1035)"
	case "ssh_public_key":
		return "Must be a valid SSH public key"
	case "oneof":
		return "Must be one of: " + fe.Param()
	case "base64":
		return "Must be a valid base64 encoded string"
	case "env_name":
		return "Must be a valid environment name. Ex. ENV, MY_ENV, my_ENV_123"
	case "env_list":
		return "Every env name must be unique"
	case "port_list_names":
		return "Every port name must be unique"
	case "port_list_numbers":
		return "Every port number must be unique per protocol"
	case "extra_domain_list":
		return "Every domain name must be unique and be a valid hostname (RFC 1035). And must point to " + conf.Env.Deployment.ExtraDomainIP
	}
	return fe.Error()
}

func CreateBindingError(err error) *body.BindingError {
	out := &body.BindingError{
		ValidationErrors: make(map[string][]string),
	}

	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		for _, fe := range ve {
			fieldName := fe.Field()
			out.ValidationErrors[fieldName] = append(out.ValidationErrors[fieldName], msgForTag(fe))
		}
	}

	var je *json.UnmarshalTypeError
	if errors.As(err, &je) {
		fieldName := je.Field
		out.ValidationErrors[fieldName] = append(out.ValidationErrors[fieldName], "Must be of type "+je.Type.String())
	}

	return out
}

func IsValidDomain(domainName string) bool {
	mustPointAt := conf.Env.Deployment.ExtraDomainIP

	ips, _ := net.LookupIP(domainName)
	for _, ip := range ips {
		if ipv4 := ip.To4(); ipv4 != nil {
			if ipv4.String() == mustPointAt {
				return true
			}
		}
	}
	return false
}
