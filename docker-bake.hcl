# docker-bake.hcl
group "default" {
  targets = ["go-deploy"]
}

variable "GIT_TAG" {
  default = "latest"
}

variable "REGISTRY" {
  default = "ghcr.io"
}

variable "REPO" {
  default = "kthcloud/go-deploy"
}

variable "RUN_USER" {
  default = "go-deploy"
}

target "go-deploy" {
  dockerfile = "Dockerfile"
  args = {
    RUN_USER = "${RUN_USER}"
  }
  tags = concat(
    ["${REGISTRY}/${REPO}:${GIT_TAG}"],
    GIT_TAG == "latest" ? [] : ["${REGISTRY}/${REPO}:latest"]
  )
}
