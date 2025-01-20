echo "Installing swaggo/swag"
go install github.com/swaggo/swag/v2/cmd/swag@latest

swag init --dir ../  -o ../docs/api/v2 --exclude ../routers/api/v1 --instanceName "V2" -v3.1
