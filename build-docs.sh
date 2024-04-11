export PATH=$(go env GOPATH)/bin:$PATH
swag init --dir ./  -o ./docs/api/v1 --exclude ./routers/api/v2 --instanceName "V1"
swag init --dir ./  -o ./docs/api/v2 --exclude ./routers/api/v1 --instanceName "V2"
