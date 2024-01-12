<p align="center">
  <img width=40% src="https://github.com/kthcloud/go-deploy/assets/26722370/f0e0729f-224a-4ac8-a88b-a6ed98760edd" />
</p>
 
<div align="center">
  <img src="https://github.com/kthcloud/go-deploy/actions/workflows/test.yml/badge.svg"\>
  <img src="https://github.com/kthcloud/go-deploy/actions/workflows/build-image.yml/badge.svg"\>
</div>

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

1. Clone the repository: `git clone https://github.com/kthcloud/go-deploy`
2. Clone the admin repository to get access to the configuration file.
3. Set up a local Redis instance and make sure the credentials in the configuration file match.
4. Build and run the project using your Go executable, and make sure the DEPLOY_CONFIG_FILE environment variables are set to the configuration file path. 
5. The API is by default available at [http://localhost:8080/v1](http://localhost:8080/v1)

## 📝 License

go-deploy is open-source software licensed under the [MIT license](https://opensource.org/licenses/MIT).

## 📧 Contact

If you have any questions or feedback, open an Issue or join [Discord channel](https://discord.gg/MuHQd6QEtM).
