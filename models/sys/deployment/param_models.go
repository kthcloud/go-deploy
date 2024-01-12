package deployment

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
	Replicas     *int

	GitHub *GitHubCreateParams

	Zone string
}

type GitHubCreateParams struct {
	Token        string
	RepositoryID int64
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
	Replicas     *int

	// ownership update
	TransferUserID *string
	TransferCode   *string
}

type BuildParams struct {
	Name      string
	Tag       string
	Branch    string
	ImportURL string
}
