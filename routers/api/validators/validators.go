package validators

import (
	"github.com/go-playground/validator/v10"
	"go-deploy/models/dto/body"
	"go-deploy/pkg/conf"
	"golang.org/x/crypto/ssh"
	"golang.org/x/net/idna"
	"net"
	"regexp"
	"strconv"
	"strings"
)

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

func DomainName(fl validator.FieldLevel) bool {
	domain, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}

	illegalSuffixes := make([]string, len(conf.Env.Deployment.Zones))
	for idx, zone := range conf.Env.Deployment.Zones {
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

func domainPointsToDeploy(domainName string) bool {
	for _, zone := range conf.Env.Deployment.Zones {
		mustPointAt := zone.CustomDomainIP

		ips, _ := net.LookupIP(domainName)
		for _, ip := range ips {
			if ipv4 := ip.To4(); ipv4 != nil {
				if ipv4.String() == mustPointAt {
					return true
				}
			}
		}

	}
	return false
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
