<p align="center">
  <img width=50% src="index/static/logo.svg" />
</p>

<div align="center">
  <img src="https://github.com/kthcloud/go-deploy/actions/workflows/run-tests.yml/badge.svg"\>
  <img src="https://github.com/kthcloud/go-deploy/actions/workflows/build-and-push-image.yml/badge.svg"\>
</div>

go-deploy is an API that is used to create websites and virtual machines built on top of kthcloud using Kubernetes with [KubeVirt](https://kubevirt.io/).

It is hosted at [https://api.cloud.cbh.kth.se/deploy/v2](https://api.cloud.cbh.kth.se/deploy/v2) using kthcloud authentication.

## 🛠️ Local environment

go-deploy can be hosted locally using the `setup.sh` script found under the `scripts/local` directory. The script will set up an entire local environment with every service installed in a Kind Kubernetes cluster on your local machine. 

Refer to the [local environment documentation](scripts/local/README.md) for more information.

## 🤝 Contributing

Contributions, issues, and feature requests are welcome!\
You can install a local development environment by following the [local environment documentation](scripts/local/README.md). 

Remember to run the tests before creating a pull request.

#### Acceptance Tests
```bash
go test ./acc/...
```
#### End-to-End Tests
```bash
go build -o go-deploy .     # Build
./go-deploy --mode=test     # Start

# Wait for the API to return 200 on /healthz
until $(curl --output /dev/null --silent --head --fail http://localhost:8080/healthz); do
    echo "Waiting for API to start"
    sleep 1
done
go test ./e2e/...          # Run e2e tests
```

## 📚 Docs

### Codebase
The code base is documented using godoc. You can view the documentation by running the documentation server locally.

1. Install godoc `go install golang.org/x/tools/cmd/godoc`
2. Run godoc `godoc -http=:6060`
3. Visit [http://localhost:6060/pkg/go-deploy/](http://localhost:6060/pkg/go-deploy/)


### API
The API is documented using OpenAPI 3.0 specification and is available at [https://api.cloud.cbh.kth.se/deploy/v2/docs/index.html](https://api.cloud.cbh.kth.se/deploy/v2/docs/index.html).

You can also run the API locally by [setting up the local environment](scripts/local/README.md) and visiting [http://localhost:8080/v2/docs/index.html](http://localhost:8080/v2/docs/index.html).


Contributions, issues, and feature requests are welcome!

If you have any questions or feedback, open an Issue or join [Discord channel](https://discord.gg/MuHQd6QEtM).

## 📝 License

go-deploy is open-source software licensed under the [MIT license](https://opensource.org/licenses/MIT).

## 📧 Contact

If you have any questions or feedback, open an Issue or join [Discord channel](https://discord.gg/MuHQd6QEtM).
