package model

type GithubActionConfig struct {
	Name string `yaml:"name"`
	On   On     `yaml:"on"`
	Jobs Jobs   `yaml:"jobs"`
}

type Push struct {
	Branches []string `yaml:"branches"`
}

type On struct {
	Push Push `yaml:"push"`
}

type With struct {
	Registry string `yaml:"registry,omitempty"`
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
	Push     bool   `yaml:"push,omitempty"`
	Tags     string `yaml:"tags,omitempty"`
}

type Steps struct {
	Name string `yaml:"name"`
	Uses string `yaml:"uses"`
	With With   `yaml:"with,omitempty"`
}

type Docker struct {
	RunsOn string  `yaml:"runs-on"`
	Steps  []Steps `yaml:"steps"`
}

type Jobs struct {
	Docker Docker `yaml:"docker"`
}
