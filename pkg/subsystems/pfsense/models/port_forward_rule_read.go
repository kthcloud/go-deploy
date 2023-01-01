package models

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
)

type PortForwardRuleRead struct {
	CurrentID int `json:"id"`
	Source    struct {
		Any string `json:"any"`
	} `json:"source"`
	Destination struct {
		Address string `json:"address"`
		Port    string `json:"port"`
	} `json:"destination"`
	IpProtocol       string `json:"ipprotocol"`
	Protocol         string `json:"protocol"`
	Target           string `json:"target"`
	LocalPort        string `json:"local-port"`
	Interface        string `json:"interface"`
	Description      string `json:"descr"`
	AssociatedRuleID string `json:"associated-rule-id"`
	Created          struct {
		Username string `json:"username"`
	} `json:"created"`
	Updated struct {
		Username string `json:"username"`
	} `json:"updated"`
}

func (rule *PortForwardRuleRead) GetIdAndName() (string, string, error) {
	uuidLen := len(uuid.New().String())
	if len(rule.Description) < uuidLen {
		return "", "", errors.New("could not parse id. details: invalid uuid length")
	}

	id := rule.Description[len(rule.Description)-uuidLen:]
	if _, err := uuid.Parse(id); err != nil {
		return "", "", fmt.Errorf("could not parse id. details: %s", err)
	}

	name := rule.Description[:len(rule.Description)-uuidLen-1]

	return id, name, nil
}
