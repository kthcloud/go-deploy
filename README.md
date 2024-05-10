<p align="center">
  <img width=50% src="index/static/logo.svg" />
</p>

<div align="center">
  <img src="https://github.com/kthcloud/go-deploy/actions/workflows/run-tests.yml/badge.svg"\>
  <img src="https://github.com/kthcloud/go-deploy/actions/workflows/build-and-push-image.yml/badge.svg"\>
</div>

go-deploy is a simple API to create deployments similar to Heroku and spin up virtual machines built on top of kthcloud.

It is publically hosted at [https://api.cloud.cbh.kth.se/deploy/v1](https://api.cloud.cbh.kth.se/deploy/v1) or view
its [status page](https://api.cloud.cbh.kth.se/deploy).

See our wiki for more information about [go-deploy](https://wiki.cloud.cbh.kth.se/index.php/Deploy).

## 📚 Docs

go-deploy is documented using godoc. You can view the documentation by running the documentation server locally.

1. Install godoc `go install golang.org/x/tools/cmd/godoc`
2. Run godoc `godoc -http=:6060`
3. Visit [http://localhost:6060/pkg/go-deploy/](http://localhost:6060/pkg/go-deploy/)

### API

The API is documented using Swagger.
You can view the API documentation by running go-deploy locally.

1. Install go.mod dependencies `go mod download`
2. Set the config environment variable `export DEPLOY_CONFIG_FILE=/path/to/config/file`
3. Run go-deploy `go run main.go`
4. Visit [http://localhost:8080/v1/docs](http://localhost:8080/deploy/v1/docs).

If you are unable to run the API locally, you can try it out using the deployed API.
Go to [https://api.cloud.cbh.kth.se/deploy/v1/docs/index.html](https://api.cloud.cbh.kth.se/deploy/v1/docs/index.html).

## 🤝 Contributing

Thank you for considering contributing to go-deploy!

Right now we don't support running go-deploy outside of kthcloud, so you would need to setup a Kubernetes cluster,
CloudStack enviroment, Harbor service and a GitLab instance.

If you are part of the development team at kthcloud you will find the current configuration file with the required
secrets in our private admin repository.

## 🚀 Getting Started

1. Clone the repository: `git clone https://github.com/kthcloud/go-deploy`
2. Clone the admin repository to get access to the configuration file.
3. Set up a local Redis instance and make sure the credentials in the configuration file match.
4. Build and run the project using your Go executable, and make sure the DEPLOY_CONFIG_FILE environment variables are
   set to the configuration file path.
5. The API is by default available at [http://localhost:8080/v1](http://localhost:8080/v1)

## 📝 License

go-deploy is open-source software licensed under the [MIT license](https://opensource.org/licenses/MIT).

## 📧 Contact

If you have any questions or feedback, open an Issue or join [Discord channel](https://discord.gg/MuHQd6QEtM).
