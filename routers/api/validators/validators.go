package validators

import (
	"encoding/json"
	"github.com/go-playground/validator/v10"
	"go-deploy/models/dto/body"
	"go-deploy/pkg/config"
	"golang.org/x/crypto/ssh"
	"golang.org/x/net/idna"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

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

func Rfc1035(fl validator.FieldLevel) bool {
	name, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}

	rfc1035 := regexp.MustCompile(`^[a-zA-Z]([a-zA-Z0-9-]*[a-zA-Z0-9])?([a-zA-Z]([a-zA-Z0-9-]*[a-zA-Z0-9])?)*$`)
	return rfc1035.MatchString(name)
}
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

func EnvName(fl validator.FieldLevel) bool {
	name, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}

	regex := regexp.MustCompile(`^[a-zA-Z]([a-zA-Z0-9-_]*[a-zA-Z0-9])?([a-zA-Z]([a-zA-Z0-9-_]*[a-zA-Z0-9])?)*$`)
	match := regex.MatchString(name)
	return match
}

func EnvList(fl validator.FieldLevel) bool {
	envList, ok := fl.Field().Interface().([]body.Env)
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

func PortListNames(fl validator.FieldLevel) bool {
	portList, ok := fl.Field().Interface().([]body.Port)
	if !ok {
		return false
	}

	names := make(map[string]bool)
	for _, port := range portList {
		if _, ok := names[port.Name]; ok {
			return false
		}
		names[port.Name] = true
	}
	return true
}

func PortListNumbers(fl validator.FieldLevel) bool {
	portList, ok := fl.Field().Interface().([]body.Port)
	if !ok {
		return false
	}

	ports := make(map[string]bool)
	for _, port := range portList {
		identifier := strconv.Itoa(port.Port) + "/" + port.Protocol
		if _, ok := ports[identifier]; ok {
			return false
		}
		ports[identifier] = true
	}
	return true
}

func PortListHttpProxies(fl validator.FieldLevel) bool {
	portList, ok := fl.Field().Interface().([]body.Port)
	if !ok {
		return false
	}

	names := make(map[string]bool)
	for _, port := range portList {
		if port.HttpProxy != nil {
			if _, ok := names[port.HttpProxy.Name]; ok {
				return false
			}
			names[port.HttpProxy.Name] = true
		}
	}

	return true
}

func DomainName(fl validator.FieldLevel) bool {
	domain, ok := fl.Field().Interface().(string)
	if !ok {
		return false
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

func CustomDomain(fl validator.FieldLevel) bool {
	domain, ok := fl.Field().Interface().(string)
	if !ok {
		return false
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

func TeamMemberList(fl validator.FieldLevel) bool {
	memberList, ok := fl.Field().Interface().([]body.TeamMemberUpdate)
	if !ok {
		return false
	}

	id := make(map[string]bool)
	for _, member := range memberList {
		if _, ok := id[member.ID]; ok {
			return false
		}
		id[member.ID] = true
	}
	return true
}

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

func goodURL(url string) bool {
	rfc3986Characters := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~:/?#[]@!$&'()*+,;="
	for _, c := range url {
		if !strings.ContainsRune(rfc3986Characters, c) {
			return false
		}
	}
	return true
}
