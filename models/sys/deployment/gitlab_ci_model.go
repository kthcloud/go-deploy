package deployment

type GitLabCiConfig struct {
	Build Build `yaml:"build"`
}

type Build struct {
	Image        string            `yaml:"image"`
	Stage        string            `yaml:"stage"`
	Services     []string          `yaml:"services"`
	BeforeScript []string          `yaml:"before_script"`
	Variables    map[string]string `yaml:"variables"`
	Script       []string          `yaml:"script"`
}
