package auth

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"go-deploy/pkg/config"
	"io/ioutil"
	"math/big"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	"github.com/patrickmn/go-cache"
	"golang.org/x/oauth2"
	"gopkg.in/square/go-jose.v2/jwt"
)

// VarianceTimer controls the max runtime of New() and AuthChain() middleware
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
	body, _ := ioutil.ReadAll(resp.Body)

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
	keyCloakToken := KeycloakToken{}
	var err error
	parsedJWT, err := jwt.ParseSigned(token.AccessToken)
	if err != nil {
		glog.Errorf("[Gin-OAuth] jwt not decodable: %s", err)
		return nil, err
	}
	key, err := getPublicKey(parsedJWT.Headers[0].KeyID, config)
	if err != nil {
		glog.Errorf("Failed to get publickey %v", err)
		return nil, err
	}

	err = parsedJWT.Claims(key, &keyCloakToken)
	if err != nil {
		glog.Errorf("Failed to get claims JWT:%+v", err)
		return nil, err
	}
	return &keyCloakToken, nil
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

func getTestTokenContainer(userID string) *TokenContainer {

	//userID := "955f0f87-37fd-4792-90eb-9bf6989e698e"

	const (
		AdminUserID   = "955f0f87-37fd-4792-90eb-9bf6989e698a"
		PowerUserID   = "955f0f87-37fd-4792-90eb-9bf6989e698b"
		DefaultUserID = "955f0f87-37fd-4792-90eb-9bf6989e698c"
	)

	tc := &TokenContainer{}
	tc.KeyCloakToken = &KeycloakToken{
		Jti:            "",
		Exp:            time.Now().Add(time.Hour * 24).Unix(),
		Nbf:            0,
		Iat:            0,
		Iss:            "http://localhost",
		Sub:            userID,
		Typ:            "Bearer",
		Azp:            "deploy",
		Nonce:          "",
		AuthTime:       0,
		SessionState:   "",
		Acr:            "",
		ClientSession:  "",
		AllowedOrigins: nil,
		ResourceAccess: nil,
		//Name:              "tester",
		//PreferredUsername: "tester",
		//GivenName:         "tester",
		//FamilyName:        "tester",
		//Email:             "test@example.com",
		RealmAccess: ServiceRole{},
		//Groups: []string{
		//	"platinum",
		//	"admin",
		//},
	}

	switch userID {
	case AdminUserID:
		tc.KeyCloakToken.Name = "tester-admin"
		tc.KeyCloakToken.PreferredUsername = "tester-admin"
		tc.KeyCloakToken.GivenName = "tester-admin-first"
		tc.KeyCloakToken.FamilyName = "tester-admin-last"
		tc.KeyCloakToken.Email = "tester-admin@test.com"
		tc.KeyCloakToken.Groups = []string{"admin", "platinum"}
	case PowerUserID:
		tc.KeyCloakToken.Name = "tester-power"
		tc.KeyCloakToken.PreferredUsername = "tester-power"
		tc.KeyCloakToken.GivenName = "tester-power-first"
		tc.KeyCloakToken.FamilyName = "tester-power-last"
		tc.KeyCloakToken.Email = "tester-power@test.com"
	case DefaultUserID:
		tc.KeyCloakToken.Name = "tester-default"
		tc.KeyCloakToken.PreferredUsername = "tester-default"
		tc.KeyCloakToken.GivenName = "tester-default-first"
		tc.KeyCloakToken.FamilyName = "tester-default-last"
		tc.KeyCloakToken.Email = "tester-default@test.com"
	default:
		return nil
	}

	return tc
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

func New(accessCheckFunction AccessCheckFunction, endpoints KeycloakConfig) gin.HandlerFunc {
	return authChain(endpoints, accessCheckFunction)
}

func authChain(kcConfig KeycloakConfig, accessCheckFunctions ...AccessCheckFunction) gin.HandlerFunc {
	// middleware
	return func(ctx *gin.Context) {
		t := time.Now()
		varianceControl := make(chan bool, 1)

		go func() {
			tokenContainer, ok := getTokenContainer(ctx, kcConfig)
			if !ok {
				if config.Config.TestMode {
					testUserID := ctx.GetHeader("go-deploy-test-user")
					if testUserID != "" {
						if testTokenContainer := getTestTokenContainer(testUserID); testTokenContainer != nil {
							for _, fn := range accessCheckFunctions {
								if fn(testTokenContainer, ctx) {
									varianceControl <- true
									return
								}
							}
						}
					}

				}

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
			return
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

func RequestLogger(keys []string, contentKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		request := c.Request
		c.Next()
		err := c.Errors
		if request.Method != "GET" && err == nil {
			data, e := c.Get(contentKey)
			if e != false { //key is non-existent
				values := make([]string, 0)
				for _, key := range keys {
					val, keyPresent := c.Get(key)
					if keyPresent {
						values = append(values, val.(string))
					}
				}
				glog.Infof("[Gin-OAuth] Request: %+v for %s", data, strings.Join(values, "-"))
			}
		}
	}
}
