variable "nginx_proxy_manager_api_url" {
  default = "noset"
}

variable "nginx_proxy_manager_username" {
  default = "noset"
}

variable "nginx_proxy_manager_password" {
  default = "noset"
}

provider "nginx-proxy-manager" {
  url      = var.nginx_proxy_manager_api_url
  username = var.nginx_proxy_manager_username
  password = var.nginx_proxy_manager_password
}
