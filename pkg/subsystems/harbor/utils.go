package harbor

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"go-deploy/utils/subsystemutils"
	"golang.org/x/crypto/bcrypt"
	"math/big"
)

func getRobotFullName(name string) string {
	return fmt.Sprintf("robot$%s", getRobotName(name))
}

func getRobotName(name string) string {
	return fmt.Sprintf("%s+%s", subsystemutils.GetPrefixedName(name), name)
}

func int64Ptr(i int64) *int64 { return &i }

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
