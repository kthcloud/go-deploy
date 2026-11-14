<p align="center">
  <img width=50% src="index/static/logo.svg" />
</p>

<div align="center">
  <img src="https://github.com/kthcloud/go-deploy/actions/workflows/run-tests.yml/badge.svg"\>
  <img src="https://github.com/kthcloud/go-deploy/actions/workflows/build-and-push-image.yml/badge.svg"\>
</div>

go-deploy is an API that is used to create container applications and virtual machines built on top of kthcloud using Kubernetes with [KubeVirt](https://kubevirt.io/).

It is hosted at [https://api.cloud.cbh.kth.se/deploy/v2](https://api.cloud.cbh.kth.se/deploy/v2) using kthcloud authentication.

## 🛠️ Local environment

A reproducible local development environment (Kind cluster + Helmfile dependencies + `config.local.yml`) is configured via `mise.toml`, `kind.yaml`, and `helmfile.yaml.gotmpl`:

```bash
mise install
mise run up
mise run run
```

Tear down with `mise run down`.

## 🤝 Contributing

Contributions, issues, and feature requests are welcome!
Set up the local development environment with `mise run up`.

Remember to run the tests before creating a pull request.

```bash
make test
# or
# make acc
# make e2e
```

## 📚 Docs

### Codebase
The code base is documented using godoc. You can view the documentation by running the documentation server locally.

1. Install godoc `go install golang.org/x/tools/cmd/godoc@latest`
2. Run godoc `godoc -http=:6060`
3. Visit [http://localhost:6060/pkg/go-deploy/](http://localhost:6060/pkg/go-deploy/)


### API
The API is documented using OpenAPI 3.1.0 specification and is available at [https://api.cloud.cbh.kth.se/deploy/v2/docs/index.html](https://api.cloud.cbh.kth.se/deploy/v2/docs/index.html).

You can also run the API locally (`mise run up && mise run run`) and visit [http://localhost:8080/v2/docs/index.html](http://localhost:8080/v2/docs/index.html).


Contributions, issues, and feature requests are welcome!

If you have any questions or feedback, open an Issue or join [Discord channel](https://discord.gg/MuHQd6QEtM).

## 📝 License

go-deploy is open-source software licensed under the [MIT license](https://opensource.org/licenses/MIT).

## 📧 Contact

If you have any questions or feedback, open an Issue or join [Discord channel](https://discord.gg/MuHQd6QEtM).
