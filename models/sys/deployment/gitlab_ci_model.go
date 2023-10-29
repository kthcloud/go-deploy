package deployment

type GitLabCiConfig struct {
	Build Build `yaml:"build"`
}

type Build struct {
	Image        string            `yaml:"image,omitempty"`
	Stage        string            `yaml:"stage,omitempty"`
	Services     []string          `yaml:"services,omitempty"`
	BeforeScript []string          `yaml:"before_script,omitempty"`
	Variables    map[string]string `yaml:"variables,omitempty"`
	Script       []string          `yaml:"script,omitempty"`
}
