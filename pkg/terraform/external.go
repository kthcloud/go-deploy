package terraform

type External struct {
	envs        map[string]string
	scriptPaths []string
}

func (modifier *External) SetEnv(key, val string) {
	modifier.envs[key] = val
}

func (modifier *External) SetScriptPaths(scriptFilepaths []string) {
	modifier.scriptPaths = scriptFilepaths
}

func (client *Client) AddModifier(fn func(*External) error) {
	client.externalGenerators = append(client.externalGenerators, fn)
}
