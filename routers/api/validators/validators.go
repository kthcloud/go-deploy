package validators

import (
	"encoding/json"
	"github.com/go-playground/validator/v10"
	bodyV1 "go-deploy/dto/v1/body"
	bodyV2 "go-deploy/dto/v2/body"
	"go-deploy/pkg/config"
	"golang.org/x/crypto/ssh"
	"golang.org/x/net/idna"
	"io"
	"net/http"
	"regexp"
	"strings"
)

// googleDnsResponse is the response from the Google DNS API
// This is used to check if a domain points to the correct IP
type googleDnsResponse struct {
	Status   int  `json:"Status,omitempty"`
	Tc       bool `json:"TC,omitempty"`
	Rd       bool `json:"RD,omitempty"`
	Ra       bool `json:"RA,omitempty"`
	Ad       bool `json:"AD,omitempty"`
	Cd       bool `json:"CD,omitempty"`
	Question []struct {
		Name string `json:"name,omitempty"`
		Type int    `json:"type,omitempty"`
	} `json:"Question,omitempty"`
	Answer []struct {
		Name string `json:"name,omitempty"`
		Type int    `json:"type,omitempty"`
		TTL  int    `json:"TTL,omitempty"`
		Data string `json:"data,omitempty"`
	} `json:"Answer,omitempty"`
	Comment string `json:"Comment,omitempty"`
}

// Rfc1035 is a validator for RFC 1035 hostnames
func Rfc1035(fl validator.FieldLevel) bool {
	name, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}

	rfc1035 := regexp.MustCompile(`^[a-zA-Z]([a-zA-Z0-9-]*[a-zA-Z0-9])?([a-zA-Z]([a-zA-Z0-9-]*[a-zA-Z0-9])?)*$`)
	return rfc1035.MatchString(name)
}

// SshPublicKey is a validator for SSH public keys.
// It attempts to parse the key using the golang.org/x/crypto/ssh package
// If the key is invalid, it will return false
func SshPublicKey(fl validator.FieldLevel) bool {
	publicKey, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}

	_, _, _, _, err := ssh.ParseAuthorizedKey([]byte(publicKey))
	if err != nil {
		return false
	}
	return true
}

// EnvName is a validator for environment variable names
// It ensure the name is valid for use in the environment
func EnvName(fl validator.FieldLevel) bool {
	name, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}

	regex := regexp.MustCompile(`^[a-zA-Z]([a-zA-Z0-9-_]*[a-zA-Z0-9])?([a-zA-Z]([a-zA-Z0-9-_]*[a-zA-Z0-9])?)*$`)
	match := regex.MatchString(name)
	return match
}

// EnvList is a validator for environment variable lists.
// It ensures that every environment variable name is unique
func EnvList(fl validator.FieldLevel) bool {
	envList, ok := fl.Field().Interface().([]bodyV1.Env)
	if !ok {
		return false
	}

	names := make(map[string]bool)
	for _, env := range envList {
		if _, ok := names[env.Name]; ok {
			return false
		}
		names[env.Name] = true
	}
	return true
}

// PortListNames is a validator for port lists.
// It ensures that every port name is unique
func PortListNames(fl validator.FieldLevel) bool {
	// We need to try with both PortCreate and PortUpdate

	portListCreateV1, ok := fl.Field().Interface().([]bodyV1.PortCreate)
	if ok {
		names := make(map[string]bool)
		for _, port := range portListCreateV1 {
			if _, exists := names[port.Name]; exists {
				return false
			}
			names[port.Name] = true
		}

		return true
	}

	portListUpdateV1, ok := fl.Field().Interface().([]bodyV1.PortUpdate)
	if ok {
		names := make(map[string]bool)
		for _, port := range portListUpdateV1 {
			if _, exists := names[port.Name]; exists {
				return false
			}
			names[port.Name] = true
		}

		return true
	}

	portListCreateV2, ok := fl.Field().Interface().([]bodyV2.PortCreate)
	if ok {
		names := make(map[string]bool)
		for _, port := range portListCreateV2 {
			if _, exists := names[port.Name]; exists {
				return false
			}
			names[port.Name] = true
		}

		return true
	}

	portListUpdateV2, ok := fl.Field().Interface().([]bodyV2.PortUpdate)
	if ok {
		names := make(map[string]bool)
		for _, port := range portListUpdateV2 {
			if _, exists := names[port.Name]; exists {
				return false
			}
			names[port.Name] = true
		}

		return true
	}

	return false
}

// PortListNumbers is a validator for port lists.
// It ensures that every port number is unique per protocol
func PortListNumbers(fl validator.FieldLevel) bool {
	// We need to try with both PortCreate and PortUpdate

	portListCreateV1, ok := fl.Field().Interface().([]bodyV1.PortCreate)
	if ok {
		numbers := make(map[int]bool)
		for _, port := range portListCreateV1 {
			if _, exists := numbers[port.Port]; exists {
				return false
			}
			numbers[port.Port] = true
		}

		return true
	}

	portListUpdateV1, ok := fl.Field().Interface().([]bodyV1.PortUpdate)
	if ok {
		numbers := make(map[int]bool)
		for _, port := range portListUpdateV1 {
			if _, exists := numbers[port.Port]; exists {
				return false
			}
			numbers[port.Port] = true
		}

		return true
	}

	portListCreateV2, ok := fl.Field().Interface().([]bodyV2.PortCreate)
	if ok {
		numbers := make(map[int]bool)
		for _, port := range portListCreateV2 {
			if _, exists := numbers[port.Port]; exists {
				return false
			}
			numbers[port.Port] = true
		}

		return true
	}

	portListUpdateV2, ok := fl.Field().Interface().([]bodyV2.PortUpdate)
	if ok {
		numbers := make(map[int]bool)
		for _, port := range portListUpdateV2 {
			if _, exists := numbers[port.Port]; exists {
				return false
			}
			numbers[port.Port] = true
		}

		return true
	}

	return false
}

// PortListHttpProxies is a validator for port lists.
// It ensures that every proxy name is unique
func PortListHttpProxies(fl validator.FieldLevel) bool {
	// We need to try with both PortCreate and PortUpdate

	portListCreateV1, ok := fl.Field().Interface().([]bodyV1.PortCreate)
	if ok {
		names := make(map[string]bool)
		for _, port := range portListCreateV1 {
			if port.HttpProxy != nil {
				if _, exists := names[port.HttpProxy.Name]; exists {
					return false
				}
				names[port.HttpProxy.Name] = true
			}
		}

		return true
	}

	portListUpdateV1, ok := fl.Field().Interface().([]bodyV1.PortUpdate)
	if ok {
		names := make(map[string]bool)
		for _, port := range portListUpdateV1 {
			if port.HttpProxy != nil {
				if _, exists := names[port.HttpProxy.Name]; exists {
					return false
				}
				names[port.HttpProxy.Name] = true
			}
		}

		return true
	}

	portListCreateV2, ok := fl.Field().Interface().([]bodyV2.PortCreate)
	if ok {
		names := make(map[string]bool)
		for _, port := range portListCreateV2 {
			if port.HttpProxy != nil {
				if _, exists := names[port.HttpProxy.Name]; exists {
					return false
				}
				names[port.HttpProxy.Name] = true
			}
		}

		return true
	}

	portListUpdateV2, ok := fl.Field().Interface().([]bodyV2.PortUpdate)
	if ok {
		names := make(map[string]bool)
		for _, port := range portListUpdateV2 {
			if port.HttpProxy != nil {
				if _, exists := names[port.HttpProxy.Name]; exists {
					return false
				}
				names[port.HttpProxy.Name] = true
			}
		}

		return true
	}

	return false
}

// DomainName is a validator for domain names.
// It ensures that the domain name is not a parent domain of any zone
// It also ensures that the domain name is valid and can be converted to punycode
func DomainName(fl validator.FieldLevel) bool {
	domain, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}

	// Deletion through empty string
	if domain == "" {
		return true
	}

	illegalSuffixes := make([]string, len(config.Config.Deployment.Zones))
	for idx, zone := range config.Config.Deployment.Zones {
		illegalSuffixes[idx] = zone.ParentDomain
	}

	for _, suffix := range illegalSuffixes {
		if strings.HasSuffix(domain, suffix) {
			return false
		}
	}

	_, err := idna.Lookup.ToASCII(domain)
	if err != nil {
		return false
	}

	return true
}

// CustomDomain is a validator for custom domain names.
// It ensures that the domain name is valid and that it points to the correct IP
func CustomDomain(fl validator.FieldLevel) bool {
	domain, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}

	// Deletion through empty string
	if domain == "" {
		return true
	}

	punyEncoded, err := idna.Lookup.ToASCII(domain)
	if err != nil {
		return false
	}

	if !domainPointsToDeploy(punyEncoded) {
		return false
	}

	return true
}

// HealthCheckPath is a validator for health check paths.
// It ensures that the path is valid and that it starts with a slash
func HealthCheckPath(fl validator.FieldLevel) bool {
	path, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}

	if len(path) > 0 && path[0] != '/' {
		return false
	}

	if !goodURL(path) {
		return false
	}

	return true
}

// TeamName is a validator for team names.
// It ensures that the name is valid and does not start or end with a space
func TeamName(fl validator.FieldLevel) bool {
	name, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}

	if name[0] == ' ' || name[len(name)-1] == ' ' {
		return false
	}

	regex := regexp.MustCompile(`^[a-zA-Z0-9-_]+$`)
	match := regex.MatchString(name)

	return match
}

// TeamMemberList is a validator for team member lists.
// It ensures that every team member is unique
func TeamMemberList(fl validator.FieldLevel) bool {
	if memberListUpdate, ok := fl.Field().Interface().([]bodyV1.TeamMemberUpdate); ok {
		id := make(map[string]bool)
		for _, member := range memberListUpdate {
			if _, ok := id[member.ID]; ok {
				return false
			}
			id[member.ID] = true
		}

		return true
	} else if memberListCreate, ok := fl.Field().Interface().([]bodyV1.TeamMemberCreate); ok {
		id := make(map[string]bool)
		for _, member := range memberListCreate {
			if _, ok := id[member.ID]; ok {
				return false
			}
			id[member.ID] = true
		}
	}

	return true
}

// TeamResourceList is a validator for team model lists.
// It ensures that every team model is unique
func TeamResourceList(fl validator.FieldLevel) bool {
	resourceList, ok := fl.Field().Interface().([]string)
	if !ok {
		return false
	}

	id := make(map[string]bool)
	for _, resourceID := range resourceList {
		if _, ok := id[resourceID]; ok {
			return false
		}
		id[resourceID] = true
	}
	return true
}

// VolumeName is a validator for volume names.
// It ensures that the name is valid and does not start or end with a space
func VolumeName(fl validator.FieldLevel) bool {
	name, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}

	if name[0] == ' ' || name[len(name)-1] == ' ' {
		return false
	}

	regex := regexp.MustCompile(`^[a-zA-Z0-9-_ ]+$`)
	match := regex.MatchString(name)

	return match
}

// domainPointsToDeploy is a helper function that checks if the domain points to the correct IP
func domainPointsToDeploy(domainName string) bool {
	for _, zone := range config.Config.Deployment.Zones {
		mustPointAt := zone.CustomDomainIP

		pointsTo := lookUpIP(domainName)
		if pointsTo == nil {
			return false
		}

		if *pointsTo == mustPointAt {
			return true
		}
	}
	return false
}

// lookUpIP is a helper function that looks up the IP of a domain using the Google DNS API
func lookUpIP(domainName string) *string {
	requestURL := "https://dns.google.com/resolve?name=" + domainName + "&type=A"
	resp, err := http.Get(requestURL)
	if err != nil {
		return nil
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}

	var response googleDnsResponse
	err = json.Unmarshal(respBody, &response)
	if err != nil {
		return nil
	}

	if len(response.Answer) == 0 {
		return nil
	}

	return &response.Answer[len(response.Answer)-1].Data
}

// goodURL is a helper function that checks if a URL is valid according to RFC 3986
func goodURL(url string) bool {
	rfc3986Characters := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~:/?#[]@!$&'()*+,;="
	for _, c := range url {
		if !strings.ContainsRune(rfc3986Characters, c) {
			return false
		}
	}
	return true
}
