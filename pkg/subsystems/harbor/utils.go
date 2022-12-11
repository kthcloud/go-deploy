package harbor

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/sethvargo/go-password/password"
	"go-deploy/pkg/subsystems/harbor/models"
	"go-deploy/utils/requestutils"
	"golang.org/x/crypto/bcrypt"
	"io"
	"math/big"
	"unicode"
)

func getRobotFullName(projectName, name string) string {
	return fmt.Sprintf("robot$%s", getRobotName(projectName, name))
}

func getRobotName(projectName, name string) string {
	return fmt.Sprintf("%s+%s", projectName, name)
}

func generateRandomString(n int) (string, error) {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-"
	ret := make([]byte, n)
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		ret[i] = letters[num.Int64()]
	}

	return string(ret), nil
}

func generateToken(secret string) (string, error) {
	salt, err := generateRandomString(10)
	if err != nil {
		return "", err
	}

	saltedSecret := fmt.Sprintf("%s%s", secret, salt)
	hash, err := bcrypt.GenerateFromPassword([]byte(saltedSecret), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	hasher := md5.New()
	hasher.Write(hash)
	encoded := hex.EncodeToString(hasher.Sum(nil))
	return encoded, nil
}

func createAuthHeader(secret string) string {
	fullSecret := fmt.Sprintf("cloud:%s", secret)
	encoded := base64.StdEncoding.EncodeToString([]byte(fullSecret))
	return fmt.Sprintf("Basic %s", encoded)
}

func createPassword() (string, error) {
	for {
		generatedPassword, err := password.Generate(15, 2, 0, false, false)
		if err != nil {
			return "", err
		}

		anyUpper := false
		anyLower := false
		for _, letter := range generatedPassword {
			if unicode.IsLower(letter) {
				anyLower = true
			} else if unicode.IsUpper(letter) {
				anyUpper = true
			}
		}
		if anyLower && anyUpper {
			return generatedPassword, nil
		}
	}
}

func makeApiError(readCloser io.ReadCloser, makeError func(error) error) error {
	body, err := requestutils.ReadBody(readCloser)
	if err != nil {
		return makeError(err)
	}
	defer requestutils.CloseBody(readCloser)

	apiError := models.ApiError{}
	err = requestutils.ParseJson(body, &apiError)
	if err != nil {
		return makeError(err)
	}

	if len(apiError.Errors) == 0 {
		requestError := fmt.Errorf("erroneous request. details: unknown")
		return makeError(requestError)
	}

	resCode := apiError.Errors[0].Code
	resMsg := apiError.Errors[0].Message
	requestError := fmt.Errorf("erroneous request (%s). details: %s", resCode, resMsg)
	return makeError(requestError)
}
