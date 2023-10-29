package github_service

import (
	"encoding/json"
	"fmt"
	"go-deploy/pkg/subsystems/github"
	"go-deploy/utils/requestutils"
	"net/url"
)

func withGitHubClient(token string) (*github.Client, error) {
	return github.New(&github.ClientConf{
		Token: token,
	})
}

func fetchAccessToken(code, clientId string, clientSecret string) (string, error) {
	apiRoute := "https://github.com/login/oauth/access_token"

	body := map[string]string{
		"client_id":     clientId,
		"client_secret": clientSecret,
		"code":          code,
	}

	bodyData, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	res, err := requestutils.DoRequest("POST", apiRoute, bodyData, nil)
	if err != nil {
		return "", err
	}

	if res.StatusCode != 200 {
		return "", fmt.Errorf("failed to get github access token. status code: %d", res.StatusCode)
	}

	readBody, err := requestutils.ReadBody(res.Body)
	if err != nil {
		return "", err
	}

	paramsStrings := string(readBody)

	params, err := url.ParseQuery(paramsStrings)
	if err != nil {
		return "", err
	}

	accessToken := params.Get("access_token")
	if accessToken == "" {
		return "", fmt.Errorf("failed to get github access token. access token is empty")
	}

	return accessToken, nil
}
