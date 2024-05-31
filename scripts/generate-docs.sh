echo "Installing swaggo/swag"
go install github.com/swaggo/swag/cmd/swag@latest

swag init --dir ../  -o ../docs/api/v2 --exclude ../routers/api/v1 --instanceName "V2"
