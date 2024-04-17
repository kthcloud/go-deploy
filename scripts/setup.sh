#!/bin/bash

# Ensure this script is run from the script folder by checking if the parent folder contains mod.go
if [ ! -f "../go.mod" ]; then
  echo "$RED_CROSS Please run this script from the scripts folder"
  exit 1
fi

source ./common.sh

GREEN_CHECK="\033[32;1m✔\033[0m"
RED_CROSS="\033[31;1m✗\033[0m"

PLACEHOLDER_GIT_REPO="https://github.com/kthcloud/go-deploy-placeholder.git"

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
    sed -i 's/reg.mydomain.com/localhost/g' ./harbor/harbor.yml
    # Edit http port to be 8000
    sed -i 's/port: 80/port: 8000/g' ./harbor/harbor.yml

    sudo ./harbor/install.sh > /dev/null 2>&1

    # Wait for Harbor to be up
    sleep 5
  fi

  # If robot_token file does exists, skip
  if ! [ -f "./harbor/robot_token" ]; then
    # Create Robot account for Harbor
    payload='{"name":"go-deploy","duration":-1,"disable":false,"level":"system","permissions":[{"kind":"project","namespace":"*","access":[{"resource":"repository","action":"pull"},{"resource":"repository","action":"push"}]}]}'
    res=$(curl -s -u admin:Harbor12345 -X POST -H "Content-Type: application/json" -d "$payload" http://localhost:8000/api/v2.0/robots)

    # If contains "already exists", then delete it and create again
    if [[ $res == *"already exists"* ]]; then
      fetched_id=$(curl -s -u admin:Harbor12345 -X GET http://localhost:8000/api/v2.0/robots | jq -r '.[] | select(.name=="robot$go-deploy") | .id')
      curl -u admin:Harbor12345 -X DELETE http://localhost:8000/api/v2.0/robots/$fetched_id
      res=$(curl -s -u admin:Harbor12345 -X POST -H "Content-Type: application/json" -d "$payload" http://localhost:8000/api/v2.0/robots)
    fi

    # res: "{"creation_time":"2024-04-17T13:41:10.609Z","expires_at":-1,"id":36,"name":"robot$go-deploy","secret":"d6LV52nMjrk11G7ufVE0ssI2gJesd4dm"}"
    # Extract the secret from the response
    secret=$(echo $res | jq -r '.secret')

    echo $secret > ./harbor/robot_token
  fi  
}

function seed_harbor_with_placeholder_images() {
  # If repository "go-deploy-placeholder" in project "library" already exists, skip
  res=$(curl -s -u admin:Harbor12345 -X GET http://localhost:8000/api/v2.0/projects/library/repositories | jq -r '.[] | select(.name=="library/go-deploy-placeholder") | .name')
  if [ "$res" == "library/go-deploy-placeholder" ]; then
    return
  fi

  # Download repo and build the image
  if [ ! -d "go-deploy-placeholder" ]; then
    git clone $PLACEHOLDER_GIT_REPO > /dev/null 2>&1
  fi

  # Use 'library' so we don't need to create our own (library is the default namespace in Harbor)
  docker build go-deploy-placeholder/ -t localhost:8000/library/go-deploy-placeholder:latest > /dev/null 2>&1
  docker login localhost:8000 -u admin -p Harbor12345 > /dev/null 2>&1
  docker push localhost:8000/library/go-deploy-placeholder:latest > /dev/null 2>&1

  # Remove the placeholder repo
  rm -rf go-deploy-placeholder
}

function generate_config() {
  # if [ ! -f "../config.local.yml" ]; then
    cp config.yml.tmpl ../config.local.yml

    export port=8080
    envsubst < config.yml.tmpl > ../config.local.yml

    # sed -i 's/port:/port: 8000/' ../config.local.yml
    # sed -i 's/placeholderImage:/placeholderImage: localhost:8000\/library\/go-deploy-placeholder/g' ../config.local.yml
  
  # fi
}


run_with_spinner "Install k3d" install_k3d
run_with_spinner "Set up k3d cluster" create_k3d_cluster
run_with_spinner "Set up MongoDB" setup_mongodb
run_with_spinner "Set up Redis" setup_redis
run_with_spinner "Set up Harbor" setup_harbor
run_with_spinner "Seed Harbor with placeholder images" seed_harbor_with_placeholder_images
run_with_spinner "Generate config.local.yml" generate_config