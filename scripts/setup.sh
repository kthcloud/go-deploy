GREEN_CHECK="\033[32;1m✔\033[0m"
RED_CROSS="\033[31;1m✗\033[0m"

# Ensure this script is run from the script folder by checking if the parent folder contains mod.go
if [ ! -f "../go.mod" ]; then
  echo "$RED_CROSS Please run this script from the scripts folder"
  exit 1
fi

k3s_install_path="curl -s https://raw.githubusercontent.com/rancher/k3d/main/install.sh | bash"

# If not exists, install k3d
if ! [ -x "$(command -v k3d)" ]; then
  echo "$RED_CROSS k3d not found, installing..."
  eval $k3s_install_path
else
  echo "$GREEN_CHECK k3d already installed"
fi

# If already exists, skip
name="go-deploy-dev"
current=$(k3d cluster list | grep -c $name)
if [ $current -eq 0 ]; then
  echo "$GREEN_CHECK Creating k3d cluster"
  k3d cluster create $name
else
  echo "$GREEN_CHECK k3d cluster already exists"
fi

# MongoDB setup in Docker
# if exists skip
if [ "$(docker ps -q -f name=go-deploy-mongodb)" ]; then
    echo "$GREEN_CHECK MongoDB already running"
else
    echo "$GREEN_CHECK Setting up MongoDB"
    docker run -d -p 27017:27017 --name go-deploy-mongodb mongo:6.0
fi


# Redis setup in Docker
if [ "$(docker ps -q -f name=go-deploy-redis)" ]; then
    echo "$GREEN_CHECK Redis already running"
else
    echo "$GREEN_CHECK Setting up Redis"
    docker run -d -p 8379:6379 --name go-deploy-redis redis:6.2
fi

# Harbor setup
curl_result=$(curl -s localhost:8000 | grep -c "Harbor")
if [ $curl_result -eq 0 ]; then
    echo "$GREEN_CHECK Setting up Harbor"
    
    # If Harbor folder does not exist, download and extract
    if [ ! -d "harbor" ]; then
        echo "$GREEN_CHECK Downloading Harbor"
        download_url="https://github.com/goharbor/harbor/releases/download/v2.9.4-rc1/harbor-offline-installer-v2.9.4-rc1.tgz"
        wget -O harbor.tgz $download_url
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

    sudo ./harbor/install.sh 
else
    echo "$GREEN_CHECK Harbor already running"
fi
