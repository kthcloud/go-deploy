package auth

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/kthcloud/go-deploy/models/mode"
	"github.com/kthcloud/go-deploy/pkg/config"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/golang/glog"
	"github.com/patrickmn/go-cache"
	"golang.org/x/oauth2"
)

// VarianceTimer controls the max runtime of SetupKeycloakChain() and AuthChain() middleware
var VarianceTimer = 30000 * time.Millisecond
var publicKeyCache = cache.New(8*time.Hour, 8*time.Hour)

// TokenContainer stores all relevant token information
type TokenContainer struct {
	Token         *oauth2.Token
	KeyCloakToken *KeycloakToken
}

func extractToken(r *http.Request) (*oauth2.Token, error) {
	hdr := r.Header.Get("Authorization")
	if hdr == "" {
		return nil, errors.New("no authorization header")
	}

	th := strings.Split(hdr, " ")
	if len(th) != 2 {
		return nil, errors.New("incomplete authorization header")
	}

	return &oauth2.Token{AccessToken: th[1], TokenType: th[0]}, nil
}

func GetTokenContainer(token *oauth2.Token, config KeycloakConfig) (*TokenContainer, error) {

	keyCloakToken, err := decodeToken(token, config)
	if err != nil {
		return nil, err
	}

	return &TokenContainer{
		Token: &oauth2.Token{
			AccessToken: token.AccessToken,
			TokenType:   token.TokenType,
		},
		KeyCloakToken: keyCloakToken,
	}, nil
}

func getPublicKey(keyId string, config KeycloakConfig) (interface{}, error) {

	keyEntry, err := getPublicKeyFromCacheOrBackend(keyId, config)
	if err != nil {
		return nil, err
	}
	if strings.ToUpper(keyEntry.Kty) == "RSA" {
		n, _ := base64.RawURLEncoding.DecodeString(keyEntry.N)
		bigN := new(big.Int)
		bigN.SetBytes(n)
		e, _ := base64.RawURLEncoding.DecodeString(keyEntry.E)
		bigE := new(big.Int)
		bigE.SetBytes(e)
		return &rsa.PublicKey{N: bigN, E: int(bigE.Int64())}, nil
	} else if strings.ToUpper(keyEntry.Kty) == "EC" {
		x, _ := base64.RawURLEncoding.DecodeString(keyEntry.X)
		bigX := new(big.Int)
		bigX.SetBytes(x)
		y, _ := base64.RawURLEncoding.DecodeString(keyEntry.Y)
		bigY := new(big.Int)
		bigY.SetBytes(y)

		var curve elliptic.Curve
		crv := strings.ToUpper(keyEntry.Crv)
		switch crv {
		case "P-224":
			curve = elliptic.P224()
		case "P-256":
			curve = elliptic.P256()
		case "P-384":
			curve = elliptic.P384()
		case "P-521":
			curve = elliptic.P521()
		default:
			return nil, errors.New("EC curve algorithm not supported " + keyEntry.Kty)
		}

		return &ecdsa.PublicKey{
			Curve: curve,
			X:     bigX,
			Y:     bigY,
		}, nil
	}

	return nil, errors.New("no support for keys of type " + keyEntry.Kty)
}

func getPublicKeyFromCacheOrBackend(keyId string, config KeycloakConfig) (KeyEntry, error) {
	entry, exists := publicKeyCache.Get(keyId)
	if exists {
		return entry.(KeyEntry), nil
	}

	u, err := url.Parse(config.Url)
	if err != nil {
		return KeyEntry{}, err
	}

	if config.FullCertsPath != nil {
		u.Path = *config.FullCertsPath
	} else {
		u.Path = path.Join(u.Path, "realms", config.Realm, "protocol/openid-connect/certs")
	}

	resp, err := http.Get(u.String())
	if err != nil {
		return KeyEntry{}, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var certs Certs
	err = json.Unmarshal(body, &certs)
	if err != nil {
		return KeyEntry{}, err
	}

	for _, keyIdFromServer := range certs.Keys {
		if keyIdFromServer.Kid == keyId {
			publicKeyCache.Set(keyId, keyIdFromServer, cache.DefaultExpiration)
			return keyIdFromServer, nil
		}
	}

	return KeyEntry{}, errors.New("No public key found with kid " + keyId + " found")
}

func decodeToken(token *oauth2.Token, config KeycloakConfig) (*KeycloakToken, error) {
	var keycloakToken KeycloakToken

	// Parse the token and extract the kid
	parsed, err := jwt.Parse(token.AccessToken, func(t *jwt.Token) (any, error) {
		// Ensure the signing method is as expected (RSA)
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}

		// Extract Key ID (kid) and fetch the appropriate public key
		kid, _ := t.Header["kid"].(string)
		key, err := getPublicKey(kid, config)
		if err != nil {
			return nil, fmt.Errorf("failed to get public key: %w", err)
		}
		return key, nil
	})
	if err != nil {
		glog.Errorf("[Gin-OAuth] jwt not decodable: %v", err)
		return nil, err
	}

	// Validate token and extract claims
	if claims, ok := parsed.Claims.(jwt.MapClaims); ok && parsed.Valid {
		if err := mapToStruct(claims, &keycloakToken); err != nil {
			glog.Errorf("Failed to parse claims into KeycloakToken: %v", err)
			return nil, err
		}
		return &keycloakToken, nil
	}

	glog.Errorf("Invalid JWT or unexpected claims structure")
	return nil, fmt.Errorf("invalid JWT or claims")
}

func mapToStruct(m jwt.MapClaims, v interface{}) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

func isExpired(token *KeycloakToken) bool {
	if token.Exp == 0 {
		return false
	}
	now := time.Now()
	fromUnixTimestamp := time.Unix(token.Exp, 0)
	return now.After(fromUnixTimestamp)
}

func getTokenContainer(ctx *gin.Context, config KeycloakConfig) (*TokenContainer, bool) {
	var oauthToken *oauth2.Token
	var tc *TokenContainer
	var err error

	if oauthToken, err = extractToken(ctx.Request); err != nil {
		return nil, false
	}

	if !oauthToken.Valid() {
		return nil, false
	}

	if tc, err = GetTokenContainer(oauthToken, config); err != nil {
		return nil, false
	}

	if isExpired(tc.KeyCloakToken) {
		return nil, false
	}

	return tc, true
}

func (t *TokenContainer) Valid() bool {
	if t.Token == nil {
		return false
	}
	return t.Token.Valid()
}

type KeycloakConfig struct {
	Url           string
	Realm         string
	FullCertsPath *string
}

func SetupKeycloakChain(accessCheckFunction AccessCheckFunction, endpoints KeycloakConfig) gin.HandlerFunc {
	return authChain(endpoints, accessCheckFunction)
}

func authChain(kcConfig KeycloakConfig, accessCheckFunctions ...AccessCheckFunction) gin.HandlerFunc {
	// middleware
	return func(ctx *gin.Context) {
		// We only do this if an API key is not supplied and a Bearer token is supplied
		// Authentication is chosen later by the supplied methods
		if ctx.Request.Header.Get("X-API-Key") != "" {
			ctx.Next()
			return
		}

		if ctx.Request.Header.Get("Authorization") == "" && config.Config.Mode != mode.Test {
			ctx.Next()
			return
		}

		t := time.Now()
		varianceControl := make(chan bool, 1)

		go func() {
			tokenContainer, ok := getTokenContainer(ctx, kcConfig)
			if !ok {
				_ = ctx.AbortWithError(http.StatusUnauthorized, errors.New("no token in context"))
				varianceControl <- false
				return
			}

			if !tokenContainer.Valid() {
				_ = ctx.AbortWithError(http.StatusUnauthorized, errors.New("invalid Token"))
				varianceControl <- false
				return
			}
			for _, fn := range accessCheckFunctions {
				if fn(tokenContainer, ctx) {
					varianceControl <- true
					return
				}
			}
			_ = ctx.AbortWithError(http.StatusForbidden, errors.New("access to the Resource is forbidden"))
			varianceControl <- false
		}()

		select {
		case ok := <-varianceControl:
			if !ok {
				glog.V(2).Infof("[Gin-OAuth] %12v %s access not allowed", time.Since(t), ctx.Request.URL.Path)
				return
			}
		case <-time.After(VarianceTimer):
			_ = ctx.AbortWithError(http.StatusGatewayTimeout, errors.New("authorization check overtime"))
			glog.V(2).Infof("[Gin-OAuth] %12v %s overtime", time.Since(t), ctx.Request.URL.Path)
			return
		}

		glog.V(2).Infof("[Gin-OAuth] %12v %s access allowed", time.Since(t), ctx.Request.URL.Path)
	}
}
