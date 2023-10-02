# 🚀 kthcloud/go-deploy
[![ci](https://github.com/kthcloud/go-deploy/actions/workflows/docker-image.yml/badge.svg)](https://github.com/kthcloud/go-deploy/actions/workflows/docker-image.yml)

go-deploy is a simple API to create deployments similar to Heroku and spin up virtual machines built on top of kthcloud.

See our wiki for more information about [go-deploy](https://wiki.cloud.cbh.kth.se/index.php/Deploy).

## 📚 Docs
go-deploy's API is documented using Swagger. You can view the API documentation by running the server and visiting [http://localhost:8080/v1/docs/index.html](http://localhost:8080/deploy/v1/docs/index.html). 

Or view the publically hosted documentation [here](https://api.cloud.cbh.kth.se/deploy/v1/docs/index.html).

## 🤝 Contributing

Thank you for considering contributing to go-deploy!

Right now we don't support running go-deploy outside of kthcloud, so you would need to setup a Kubernetes cluster, CloudStack enviroment, Harbor service and a GitLab instance. 

If you are part of the development team at kthcloud you will find the current configuration file with the required secrets in our private admin repository.

## 🚀 Getting Started

1. Clone the repository: `git clone https://github.com/kthcloud/go-deploy/edit/main/README.md.git`
2. Build and run the project using your Go exectuatable, and make sure the DEPLOY_CONFIG_FILE environment variables are set to the configuration file path.
3. The API is by default available at [http://localhost:8080/v1](http://localhost:8080/v1)

## 📝 License

go-deploy is open-source software licensed under the [MIT license](https://opensource.org/licenses/MIT).

## 📧 Contact

If you have any questions or feedback, please feel free to contact us at [email@example.com](mailto:email@example.com).
