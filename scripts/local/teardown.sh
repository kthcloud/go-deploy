#!/bin/bash

# Ensure this script is run from the script folder by checking if the parent folder contains mod.go
if [ ! -f "../../go.mod" ]; then
  echo "$RED_CROSS Please run this script from the scripts folder"
  exit 1
fi

source "./common.sh"

GREEN_CHECK="\033[32;1m✔\033[0m"
RED_CROSS="\033[31;1m✗\033[0m"
DIR=$(pwd)

function delete_k3d_cluster() {
  name="go-deploy-dev"
  current=$(k3d cluster list | grep -c $name)
  if [ $current -eq 1 ]; then
    k3d cluster delete $name
  fi
}

function delete_harbor() {
  # If the docker-compose file exists, bring down the services
  if [ -f "harbor/docker-compose.yml" ]; then
      sudo docker compose -f $DIR/harbor/docker-compose.yml down > /dev/null 2>&1
  fi

  sudo rm -rf $DIR/harbor
  sudo rm -f $DIR/harbor*.tgz.*
}

function delete_local_dns_record() {
  sudo rm -f /etc/dnsmasq.d/50-go-deploy-dev.conf
}

run_with_spinner "Delete Harbor" delete_harbor
run_with_spinner "Delete K3d Cluster" delete_k3d_cluster
run_with_spinner "Delete local DNS record" delete_local_dns_record