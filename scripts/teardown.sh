GREEN_CHECK="\033[32;1m✔\033[0m"
RED_CROSS="\033[31;1m✗\033[0m"

# Ensure this script is run from the script folder by checking if the parent folder contains mod.go
if [ ! -f "../go.mod" ]; then
  echo "$RED_CROSS Please run this script from the scripts folder"
  exit 1
fi

echo "$GREEN_CHECK Removing k3d cluster"
k3d cluster delete go-deploy-dev

if [ "$(docker ps -q -f name=go-deploy-mongodb)" ]; then
    echo "$GREEN_CHECK Removing MongoDB"
    docker stop go-deploy-mongodb
    docker rm go-deploy-mongodb
else
    echo "$GREEN_CHECK MongoDB not running"
fi

if [ "$(docker ps -q -f name=go-deploy-redis)" ]; then
    echo "$GREEN_CHECK Removing Redis"
    docker stop go-deploy-redis
    docker rm go-deploy-redis
else
    echo "$GREEN_CHECK Redis not running"
fi

if [ -d "harbor" ]; then
    echo "$GREEN_CHECK Removing Harbor"
    # If the docker-compose file exists, bring down the services
    if [ -f "harbor/docker-compose.yml" ]; then
        sudo docker compose -f harbor/docker-compose.yml down
    fi
    sudo rm -rf harbor
else
    echo "$GREEN_CHECK Harbor not running"
fi