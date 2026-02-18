// This will generate api types so we dont have to have them as runtime deps.
// As you can see below, nvidias dra package requires a flag to be set to a version
// we dont want to have to do this when we compile go-deploy.
// Thus the struct that we depend on from that package is generated as a copy, by
// parsing the json tags of said struct. This is controlled in the
// ./cmd/structclone/main.go file, it is extendable and more structs can be added if wanted.

//go:generate go run -ldflags=-X=github.com/NVIDIA/k8s-dra-driver-gpu/internal/info.version=v1.34.2 ./cmd/structclone/main.go

package api
