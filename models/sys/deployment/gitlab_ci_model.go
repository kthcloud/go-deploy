package deployment

type GitLabCiConfig struct {
	Build Build `yaml:"build"`
}

type Build struct {
	Image     string            `yaml:"image"`
	Stage     string            `yaml:"stage"`
	Services  []string          `yaml:"services"`
	Variables map[string]string `yaml:"variables"`
	Script    []string          `yaml:"script"`
}
