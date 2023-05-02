package v1

import (
	"encoding/json"
	"errors"
	"github.com/go-playground/validator/v10"
	"go-deploy/models/dto/body"
	"go-deploy/pkg/app"
	"go-deploy/pkg/auth"
	"go-deploy/pkg/conf"
	"golang.org/x/crypto/ssh"
	"reflect"
)

type AuthInfo struct {
	UserID      string              `json:"userId"`
	JwtToken    *auth.KeycloakToken `json:"jwtToken"`
	IsAdmin     bool                `json:"isAdmin"`
	IsPowerUser bool                `json:"isPowerUser"`
}

func IsAdmin(context *app.ClientContext) bool {
	return InGroup(context, conf.Env.Keycloak.AdminGroup)
}

func IsPowerUser(context *app.ClientContext) bool {
	return InGroup(context, conf.Env.Keycloak.PowerUserGroup)
}

func InGroup(context *app.ClientContext, group string) bool {
	token, err := context.GetKeycloakToken()
	if err != nil {
		return false
	}

	for _, userGroup := range token.Groups {
		if userGroup == group {
			return true
		}
	}

	return false
}

func NilOrFalse(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

func WithAuth(context *app.ClientContext) (*AuthInfo, error) {
	token, err := context.GetKeycloakToken()
	if err != nil {
		return nil, err
	}

	return &AuthInfo{
		UserID:      token.Sub,
		JwtToken:    token,
		IsAdmin:     IsAdmin(context),
		IsPowerUser: IsPowerUser(context),
	}, nil

}

func IsValidSshPublicKey(key string) bool {
	_, _, _, _, err := ssh.ParseAuthorizedKey([]byte(key))
	if err != nil {
		return false
	}
	return true
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
	}
	return fe.Error()
}

func CreateBindingError(data interface{}, err error) *body.BindingError {
	out := &body.BindingError{
		ValidationErrors: make(map[string][]string),
	}

	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		for _, fe := range ve {
			fieldName := fe.Field()
			field, _ := reflect.TypeOf(data).Elem().FieldByName(fieldName)
			fieldJSONName, _ := field.Tag.Lookup("json")

			out.ValidationErrors[fieldJSONName] = append(out.ValidationErrors[fieldJSONName], msgForTag(fe))
		}
	}

	var je *json.UnmarshalTypeError
	if errors.As(err, &je) {
		fieldName := je.Field
		out.ValidationErrors[fieldName] = append(out.ValidationErrors[fieldName], "Must be of type "+je.Type.String())
	}

	return out
}

func CreateBindingErrorFromString(fieldName, message string) *body.BindingError {
	out := &body.BindingError{
		ValidationErrors: make(map[string][]string),
	}

	out.ValidationErrors[fieldName] = append(out.ValidationErrors[fieldName], message)

	return out
}
