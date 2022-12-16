variable "cloudstack_api_url" {
  default = "noset"
}

variable "cloudstack_api_key" {
  default = "noset"
}

variable "cloudstack_secret_key" {
  default = "noset"
}

provider "cloudstack" {
  api_url    = var.cloudstack_api_url
  api_key    = var.cloudstack_api_key
  secret_key = var.cloudstack_secret_key
}