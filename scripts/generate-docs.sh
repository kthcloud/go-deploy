echo "Installing swaggo/swag"
go install github.com/swaggo/swag/cmd/swag@latest

# You might need to "source ./scripts/path-cmd.sh"
swag init --dir ../  -o ../docs/api/v1 --exclude ../routers/api/v2 --instanceName "V1" && swag init --dir ../  -o ../docs/api/v2 --exclude ../routers/api/v1 --instanceName "V2"
