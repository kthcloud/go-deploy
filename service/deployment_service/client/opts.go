package client

// Opts is used to specify which resources to get.
// For example, if you want to get only the deployment, you can use OptsOnlyDeployment.
// If you want to get only the client, you can use OptsOnlyClient.
// If you want to get both the deployment and the client, you can use OptsAll.
type Opts struct {
	Deployment bool
	Client     bool
	Generator  bool
}

var (
	OptsAll = &Opts{
		Deployment: true,
		Client:     true,
		Generator:  true,
	}
	OptsNoDeployment = &Opts{
		Deployment: false,
		Client:     true,
		Generator:  true,
	}
	OptsNoGenerator = &Opts{
		Deployment: true,
		Client:     true,
		Generator:  false,
	}
	OptsOnlyDeployment = &Opts{
		Deployment: true,
		Client:     false,
		Generator:  false,
	}
	OptsOnlyClient = &Opts{
		Deployment: false,
		Client:     true,
		Generator:  false,
	}
)
