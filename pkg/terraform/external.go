package terraform

import "github.com/hashicorp/hcl/v2/hclwrite"

type Provider struct {
	Name    string
	Source  string
	Version string
}

type External struct {
	Envs        map[string]string
	ScriptPaths []string
	Scripts     map[string]*hclwrite.File
	Provider    Provider
}

func (client *Client) Add(external *External) {
	client.externals = append(client.externals, *external)
}
