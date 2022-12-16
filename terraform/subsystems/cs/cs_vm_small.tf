resource "cloudstack_instance" "web" {
  name             = "deploy-demo"
  service_offering = "Small HA"
  network_id       = "4a065a52-f290-4d2e-aeb4-6f48d3bd9bfe"
  template         = "e1a0479c-76a2-44da-8b38-a3a3fa316287"
  zone             = "Flemingsberg"
  project          = "deploy"
}