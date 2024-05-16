package utils

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"go-deploy/pkg/log"
	"math/rand"
	"strings"
	"time"
)

// HashString hashes a string using sha256 and returns the base64 encoded string
func HashString(token string) string {
	hasher := sha256.New()
	hasher.Write([]byte(token))
	return base64.URLEncoding.EncodeToString(hasher.Sum(nil))
}

// HashStringRfc1123 hashes a string using sha256 and returns the base64 encoded string
// It also removes the characters that are not allowed in a RFC1123 string
func HashStringRfc1123(token string) string {
	urlHash := HashString(token)
	urlHash = strings.ReplaceAll(urlHash, "_", "")
	urlHash = strings.ReplaceAll(urlHash, "=", "")
	urlHash = strings.ReplaceAll(urlHash, "-", "")
	return strings.ToLower(urlHash)
}

// HashStringAlphanumeric hashes a string using sha256 and returns the base64 encoded string
// It also removes the characters that are not allowed in a alphanumeric string
func HashStringAlphanumeric(token string) string {
	urlHash := HashString(token)
	urlHash = strings.ReplaceAll(urlHash, "_", "")
	urlHash = strings.ReplaceAll(urlHash, "=", "")
	urlHash = strings.ReplaceAll(urlHash, "-", "")
	return urlHash
}

// HashStringAlphanumericLower is a lowercase version of HashStringAlphanumeric
func HashStringAlphanumericLower(token string) string {
	return strings.ToLower(HashStringAlphanumeric(token))
}

// GetPage returns a page of a list
// If pageSize is 0, the whole list is returned
func GetPage[T any](list []T, pageSize, page int) []T {
	if pageSize == 0 {
		return list
	}

	start := page * pageSize
	end := start + pageSize

	if start > len(list) {
		return make([]T, 0)
	}

	if end > len(list) {
		end = len(list)
	}

	return list[start:end]
}

// PrettyPrintError prints an error in a pretty way
// It splits on `details: ` and adds `due to: \n`
func PrettyPrintError(err error) {
	all := make([]string, 0)

	currentErr := err
	for errors.Unwrap(currentErr) != nil {
		all = append(all, currentErr.Error())
		currentErr = errors.Unwrap(currentErr)
	}

	all = append(all, currentErr.Error())

	for i := 0; i < len(all); i++ {
		// remove the parts of the string that exists in all[i+1]
		if i+1 < len(all) {
			all[i] = strings.ReplaceAll(all[i], all[i+1], "")
		}
	}

	for i := 0; i < len(all); i++ {
		// remove details: from the string
		all[i] = strings.ReplaceAll(all[i], "details: ", "")

		// remove punctuation in the end of the string
		all[i] = strings.TrimRight(all[i], ".")
	}

	output := ""

	for i := 0; i < len(all); i++ {
		if i == 0 {
			output = all[i]
		} else {
			output = fmt.Sprintf("%s\n\tdue to: %s", output, all[i])
		}
	}

	log.Println(output)
}

// EmptyValue checks if a string pointer is nil or empty
func EmptyValue(s *string) bool {
	return s != nil && *s == ""
}

// NonZeroOrNil returns a pointer to a time.Time if the time is not zero, otherwise it returns nil
func NonZeroOrNil(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}

	return &t
}

const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// GenerateSalt generates the salt that can be used when hashing a password
func GenerateSalt() string {
	//goland:noinspection GoDeprecation
	rand.Seed(time.Now().UnixNano())
	salt := make([]byte, 16)
	for i := range salt {
		salt[i] = alphabet[rand.Intn(len(alphabet))]
	}
	return string(salt)
}

// StrPtr converts a string to a pointer
func StrPtr(s string) *string {
	return &s
}

// Int64Ptr converts an int64 to a pointer
func Int64Ptr(i int64) *int64 {
	return &i
}

// PtrOf converts a value to a pointer
func PtrOf[T any](v T) *T {
	return &v
}
