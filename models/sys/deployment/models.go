package deployment

import "go-deploy/models/sys/deployment/subsystems"

const (
	TypeCustom   = "custom"
	TypePrebuilt = "prebuilt"
)

type App struct {
	Name         string   `bson:"name"`
	Image        string   `bson:"image"`
	InternalPort int      `bson:"internalPort"`
	Private      bool     `bson:"private"`
	Envs         []Env    `bson:"envs"`
	Volumes      []Volume `bson:"volumes"`
	InitCommands []string `bson:"initCommands"`
	CustomDomain *string  `bson:"customDomain,omitempty"`
	PingPath     string   `bson:"pingPath"`
	PingResult   int      `bson:"pingResult"`
}

type Subsystems struct {
	K8s    subsystems.K8s    `bson:"k8s"`
	Harbor subsystems.Harbor `bson:"harbor"`
	GitHub subsystems.GitHub `bson:"github"`
	GitLab subsystems.GitLab `bson:"gitlab"`
}

type Env struct {
	Name  string `json:"name" bson:"name"`
	Value string `json:"value" bson:"value"`
}

type Volume struct {
	Name       string `bson:"name"`
	Init       bool   `bson:"init"`
	AppPath    string `bson:"appPath"`
	ServerPath string `bson:"serverPath"`
}

type Usage struct {
	Count int
}

type UpdateParams struct {
	// update
	Name         *string
	OwnerID      *string
	Private      *bool
	Envs         *[]Env
	InternalPort *int
	Volumes      *[]Volume
	InitCommands *[]string
	CustomDomain *string
	Image        *string
	PingPath     *string

	// ownership update
	TransferUserID *string
	TransferCode   *string
}

type GitHubCreateParams struct {
	Token        string
	RepositoryID int64
}

type CreateParams struct {
	Name string
	Type string

	Image        string
	InternalPort int
	Private      bool
	Envs         []Env
	Volumes      []Volume
	InitCommands []string
	PingPath     string
	CustomDomain *string

	GitHub *GitHubCreateParams

	Zone string
}

type BuildParams struct {
	Name      string
	Tag       string
	Branch    string
	ImportURL string
}

type GitHubRepository struct {
	ID            int64
	Name          string
	Owner         string
	CloneURL      string
	DefaultBranch string
}

type GitHubWebhook struct {
	ID     int64
	Events []string
}

type Transfer struct {
	Code   string `bson:"code"`
	UserID string `bson:"userId"`
}
