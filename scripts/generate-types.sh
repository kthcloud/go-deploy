#!/bin/sh

go install github.com/gzuidhof/tygo@latest

export PATH=$(go env GOPATH)/bin:$PATH

cd ../export && tygo generate --config tygo.yml