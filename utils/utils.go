package utils

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"strings"
)

func HashString(token string) string {
	hasher := sha256.New()
	hasher.Write([]byte(token))
	return base64.URLEncoding.EncodeToString(hasher.Sum(nil))
}

func HashStringRfc1123(token string) string {
	urlHash := HashString(token)
	urlHash = strings.ReplaceAll(urlHash, "_", "")
	urlHash = strings.ReplaceAll(urlHash, "=", "")
	urlHash = strings.ReplaceAll(urlHash, "-", "")
	return strings.ToLower(urlHash)
}

func GetPage[T any](list []T, pageSize, page int) []T {
	if pageSize == 0 {
		return list
	}

	start := (page - 1) * pageSize
	end := page * pageSize

	if start > len(list) {
		return make([]T, 0)
	}

	if end > len(list) {
		end = len(list)
	}

	return list[start:end]
}

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
