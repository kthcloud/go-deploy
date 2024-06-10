package v2

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"go-deploy/dto/v2/body"
	"go-deploy/pkg/sys"
	"go-deploy/service/core"
	"reflect"
)

// WithAuth returns the auth info for the current request.
// This includes all user info necessary to perform authorization checks
func WithAuth(context *sys.ClientContext) (*core.AuthInfo, error) {
	makeError := func(error) error {
		return errors.New("failed to get auth info")
	}

	user := context.GetAuthUser()
	if user == nil {
		return nil, makeError(fmt.Errorf("auth user not found in context"))
	}

	return core.CreateAuthInfo(user), nil
}

// msgForTag returns a human readable error message for a validator.FieldError
func msgForTag(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Must be a valid email address"
	case "eq":
		return "Must be " + fe.Param()
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
		return "Must be a valid environment name, ex. ENV, MY_ENV, my_ENV_123"
	case "env_list":
		return "Every env name must be unique"
	case "port_list_names":
		return "Every port name must be unique"
	case "port_list_numbers":
		return "Every port number must be unique per protocol"
	case "port_list_http_proxies":
		return "Every proxy name must be unique"
	case "valid_domain":
		return "Must be a valid domain name that is convertible to punycode, and less than 243 characters"
	case "health_check_path":
		return "Must be a valid path (RFC 3986), ex. /healthz or /ping-me"
	case "team_name":
		return "Must not start or end with a space and contain only alphanumeric characters, dashes, and underscores"
	case "team_member_list":
		return "Every team member must be unique"
	case "team_resource_list":
		return "Every team model must be unique"
	case "time_in_future":
		return "Must be a time in the future"
	case "volume_name":
		return "Must be a valid volume name, ex. my-volume, my-volume-123, my volume"
	case "domain_name":
		return "Must be a valid domain name that can be puny-encoded and is less than 243 characters"
	case "deployment_name":
		return "Must not end with -custom-domain or -auth-proxy"
	case "vm_name":
		return "Must not end with"
	case "vm_port_name":
		return "Must not end with -custom-domain or -proxy"
	}
	return fe.Error()
}

// CreateBindingError creates a binding error from a validator error
// This is useful for returning a binding error from a gin handler
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
