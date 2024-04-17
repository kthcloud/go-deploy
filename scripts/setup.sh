#!/bin/bash

# Ensure this script is run from the script folder by checking if the parent folder contains mod.go
if [ ! -f "../go.mod" ]; then
  echo "$RED_CROSS Please run this script from the scripts folder"
  exit 1
fi

source ./common.sh

GREEN_CHECK="\033[32;1m✔\033[0m"
RED_CROSS="\033[31;1m✗\033[0m"

# Check if Docker is installed, if not exit
if ! [ -x "$(command -v docker)" ]; then
  echo -e "$RED_CROSS Docker is not installed. Please install Docker"
  exit 1
fi

# If not exists, install k3d
function install_k3d() {
  k3s_install_path="curl -s https://raw.githubusercontent.com/rancher/k3d/main/install.sh | bash"
  if ! [ -x "$(command -v k3d)" ]; then
    eval $k3s_install_path
  fi
}


function create_k3d_cluster() {
  name="go-deploy-dev"
  current=$(k3d cluster list | grep -c $name)
  if [ $current -eq 0 ]; then
    k3d cluster create $name
  fi
}


function setup_mongodb() {
  if ! [ "$(docker ps -q -f name=go-deploy-mongodb)" ]; then
    sudo docker run -d -p 27017:27017 --name go-deploy-mongodb mongo:6.0
  fi
}

function setup_redis() {
  if ! [ "$(docker ps -q -f name=go-deploy-redis)" ]; then
    sudo docker run -d -p 8379:6379 --name go-deploy-redis redis:6.2
  fi
}

function setup_harbor() {
  curl_result=$(curl -s localhost:8000 | grep -c "Harbor")
  if [ $curl_result -eq 0 ]; then     
      # If Harbor folder does not exist, download and extract
      if [ ! -d "harbor" ]; then
          download_url="https://github.com/goharbor/harbor/releases/download/v2.9.4-rc1/harbor-offline-installer-v2.9.4-rc1.tgz"
          wget -O harbor.tgz $download_url -q
          tar xvf harbor.tgz
          rm -rf harbor.tgz
      fi

      cp ./harbor/harbor.yml.tmpl ./harbor/harbor.yml
      # Disable https
      sed -i '/# https related config/,/private_key: \/your\/private\/key\/path/d' ./harbor/harbor.yml
      # Set hostname to harbor.local
      sed -i 's/reg.mydomain.com/harbor.local/g' ./harbor/harbor.yml
      # Edit http port to be 8000
      sed -i 's/port: 80/port: 8000/g' ./harbor/harbor.yml

      sudo ./harbor/install.sh > /dev/null 2>&1
  fi
}

run_with_spinner "Install k3d" install_k3d
run_with_spinner "Set up k3d cluster" create_k3d_cluster
run_with_spinner "Set up MongoDB" setup_mongodb
run_with_spinner "Set up Redis" setup_redis
run_with_spinner "Set up Harbor" setup_harbor