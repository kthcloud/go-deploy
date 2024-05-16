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

function delete_kind_cluster() {
  name="go-deploy-dev"
  current=$(kind get clusters 2> /dev/stdout | grep -c $name)
  if [ $current -eq 1 ]; then
    kind delete cluster --name $name > /dev/null 2>&1
  fi

  rm -f ./manifests/kind-config.yml
}

function delete_local_dns_record() {
  sudo rm -f /etc/dnsmasq.d/50-go-deploy-dev.conf
}

run_with_spinner "Delete Kind cluster" delete_kind_cluster
run_with_spinner "Delete local DNS record" delete_local_dns_record