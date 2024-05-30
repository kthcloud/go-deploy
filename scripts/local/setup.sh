#!/bin/bash

WAIT_SLEEP=10

source ./common.sh

# Ensure this script is run from the script folder by checking if the parent folder contains mod.go
if [ ! -f "../../go.mod" ]; then
  echo "$RED_CROSS Please run this script from the scripts folder"
  exit 1
fi

function print_usage() {
  echo -e "Usage: $0 [options]"
  echo -e "Options:"
  echo -e "  -h, --help\t\t\tPrint this help message"
  echo -e "  -y, --yes\t\t\tSkip confirmations. Default: false"
  echo -e "  --name [name]\t\t\tName of the cluster to create. Default: go-deploy-dev"
  echo -e "  --kubeconfig [path]\t\tPath to kubeconfig file that a new context will be added to. Default: ~/.kube/config"
  echo -e "  --non-interactive\t\tSkip all user input and fancy output. Default: false"
  echo -e "  --configure-dns\t\tConfigure dnsmasq. This will override your local DNS settings and fallback to systemd-resolved. Default: false"
  echo -e ""
  echo -e "dnsmasq is used to allow the names to resolve. See the following guides for help configuring it:"
  echo -e " - WSL2 (Windows): https://github.com/absolunet/pleaz/blob/production/documentation/installation/wsl2/dnsmasq.md"
  echo -e " - systemd-resolved (Linux): https://gist.github.com/frank-dspeed/6b6f1f720dd5e1c57eec8f1fdb2276df"
}

function parse_flags() {
  local args=("$@")
  local index=0

  SKIP_CONFIRMATIONS=false
  CLUSTER_NAME="go-deploy-dev"
  KUBECONFIG_PATH="${HOME}/.kube/config"  
  CONFIGURE_DNS=false

  while [[ $index -lt ${#args[@]} ]]; do
    case "${args[$index]}" in
      -h|--help)
        print_usage
        exit 0
        ;;
      -y|--yes)
        SKIP_CONFIRMATIONS=true
        ((index++))
        ;;
      --name)
        ((index++))
        CLUSTER_NAME="${args[$index]}"
        ((index++))
        ;;
      --kubeconfig)
        ((index++))
        KUBECONFIG_PATH="${args[$index]}"
        ((index++))
        ;;
      --configure-dns)
        ((index++))
        CONFIGURE_DNS=true
        ;;
      *)
        echo "Error: Unrecognized argument: ${args[$index]}"
        print_usage
        exit 1
        ;;
    esac
  done

  # If NON_INTERACTIVE is set, skip all confirmations
  if [ "$NON_INTERACTIVE" = true ]; then
    SKIP_CONFIRMATIONS=true
  fi
}


function check_dependencies() {
  # Check if Docker is installed, if not exit
  if ! [ -x "$(command -v docker)" ]; then
    echo -e "$RED_CROSS Docker is not installed. Please install Docker (https://docs.docker.com/engine/install)"
    exit 1
  fi

  # Check if Docker daemon is running, if not exit
  local res=$(docker info > /dev/stdout 2>&1 | grep -c "Cannot connect to the Docker daemon")
  if [ $res -ne 0 ]; then
    echo -e "$RED_CROSS Docker daemon is not running. Please start the Docker daemon"
    exit 1
  fi

  # Check if kubectl is installed, if not exit
  if ! [ -x "$(command -v kubectl)" ]; then
    echo -e "$RED_CROSS kubectl is not installed. Please install kubectl (https://kubernetes.io/docs/tasks/tools)"
    exit 1
  fi

  # Check if "permission denied" is in the output, if so, the user does not have permission to run docker
  local res=$(docker info > /dev/stdout 2>&1 | grep -c "permission denied")
  if [ $res -ne 0 ]; then
    echo -e "$RED_CROSS User does not have permission to run Docker. Please add the user to the docker group"
    exit 1
  fi

  # Check if Helm is installed, if not exit
  if ! [ -x "$(command -v helm)" ]; then
    echo -e "$RED_CROSS Helm is not installed. Please install Helm (https://helm.sh/docs/intro/install)"
    exit 1
  fi

  # Check if kind is installed, if not exit
  if ! [ -x "$(command -v kind)" ]; then
    echo -e "$RED_CROSS kind is not installed. Please install kind (https://kind.sigs.k8s.io/docs/user/quick-start/#installing-with-a-package-manager)"
    exit 1
  fi

  # Check if dnsmasq is installed, if not exit
  if ! [ -x "$(command -v dnsmasq)" ]; then
    echo -e "$RED_CROSS dnsmasq is not installed. Please install dnsmasq"
    echo -e ""
    echo -e "dnsmasq is used to allow the names to resolve. See the following guides for help configuring it:"
    echo -e " - WSL2 (Windows): https://github.com/absolunet/pleaz/blob/production/documentation/installation/wsl2/dnsmasq.md"
    echo -e " - systemd-resolved (Linux): https://gist.github.com/frank-dspeed/6b6f1f720dd5e1c57eec8f1fdb2276df"
    
    exit 1
  fi

  # Check if /etc/dnsqmasq.d exists, if not exit
  if ! [ -d "/etc/dnsmasq.d" ]; then
    echo -e "$RED_CROSS /etc/dnsmasq.d does not exist. This is usually caused by dnsmasq not being installed correctly"
    exit 1
  fi
}

# Function to update /etc/systemd/resolved.conf
function update_resolved_conf() {
    echo -e "${BLUE_RIGHT_ARROW} Updating /etc/systemd/resolved.conf..."

    RESOLVED_CONF="/etc/systemd/resolved.conf"
    DNS_LINE="DNS=127.0.0.2"

    # Check if RESOLVED_CONF exists
    if [ ! -f $RESOLVED_CONF ]; then
        echo -e "WARNING: $RESOLVED_CONF does not exist. Is systemd-resolved installed?"
        return
    fi

    if ! sudo grep -q "^$DNS_LINE" $RESOLVED_CONF; then
        sudo sed -i "/^\[Resolve\]/a $DNS_LINE" $RESOLVED_CONF

        # Print a warning if the DNS line was not added
        if ! sudo grep -q "^$DNS_LINE" $RESOLVED_CONF; then
            echo -e "WARNING: Could not add $DNS_LINE to $RESOLVED_CONF"
        fi
    fi

    echo -e "${GREEN_CHECK} Updated /etc/systemd/resolved.conf"
}

# Function to update /etc/dnsmasq.conf
function update_dnsmasq_conf() {
    echo -e "${BLUE_RIGHT_ARROW} Updating /etc/dnsmasq.conf..."

    DNSMASQ_CONF="/etc/dnsmasq.conf"
    LISTEN_ADDRESS="listen-address=127.0.0.2"
    BIND_INTERFACES="bind-interfaces"

    if [ ! -f $DNSMASQ_CONF ]; then
        echo -e "WARNING: $DNSMASQ_CONF does not exist. Is dnsmasq installed?"
        return
    fi

    if ! sudo grep -q "^$LISTEN_ADDRESS" $DNSMASQ_CONF; then
        echo "$LISTEN_ADDRESS" | sudo tee -a $DNSMASQ_CONF > /dev/null

        # Print a warning if the listen-address line was not added
        if ! sudo grep -q "^$LISTEN_ADDRESS" $DNSMASQ_CONF; then
            echo -e "WARNING: Could not add $LISTEN_ADDRESS to $DNSMASQ_CONF"
        fi
    fi

    if ! sudo grep -q "^$BIND_INTERFACES" $DNSMASQ_CONF; then
        echo "$BIND_INTERFACES" | sudo tee -a $DNSMASQ_CONF > /dev/null

        # Print a warning if the bind-interfaces line was not added
        if ! sudo grep -q "^$BIND_INTERFACES" $DNSMASQ_CONF; then
            echo -e "WARNING: Could not add $BIND_INTERFACES to $DNSMASQ_CONF"
        fi
    fi

    echo -e "${GREEN_CHECK} Updated /etc/dnsmasq.conf"
}

# Function to update /etc/default/dnsmasq
function update_default_dnsmasq() {
    echo -e "${BLUE_RIGHT_ARROW} Updating /etc/default/dnsmasq..."

    DEFAULT_DNSMASQ="/etc/default/dnsmasq"
    IGNORE_RESOLVCONF="IGNORE_RESOLVCONF=yes"
    ENABLED="ENABLED=1"

    if [ ! -f $DEFAULT_DNSMASQ ]; then
        echo -e "WARNING: $DEFAULT_DNSMASQ does not exist. Is dnsmasq installed?"
        return
    fi

    if ! sudo grep -q "^$IGNORE_RESOLVCONF" $DEFAULT_DNSMASQ; then
        echo "$IGNORE_RESOLVCONF" | sudo tee -a $DEFAULT_DNSMASQ > /dev/null

        # Print a warning if the IGNORE_RESOLVCONF line was not added
        if ! sudo grep -q "^$IGNORE_RESOLVCONF" $DEFAULT_DNSMASQ; then
            echo -e "WARNING: Could not add $IGNORE_RESOLVCONF to $DEFAULT_DNSMASQ"
        fi
    fi

    if ! sudo grep -q "^$ENABLED" $DEFAULT_DNSMASQ; then
        echo "$ENABLED" | sudo tee -a $DEFAULT_DNSMASQ > /dev/null

        # Print a warning if the ENABLED line was not added
        if ! sudo grep -q "^$ENABLED" $DEFAULT_DNSMASQ; then
            echo -e "WARNING: Could not add $ENABLED to $DEFAULT_DNSMASQ"
        fi
    fi

    echo -e "${GREEN_CHECK} Updated /etc/default/dnsmasq"
}

function increase_max_file_watchers() {
  # System defaults for comparison
  default_max_user_instances=128
  default_max_queued_events=16384
  default_max_user_watches=8192

  # Total available memory in KB for the inotify settings
  available_memory_kb=$((2 * 1024 * 1024))  # 2 GB in KB

  # Calculate the total "weight" based on default values to keep the same ratio
  total_weight=$((default_max_user_watches + default_max_user_watches + default_max_user_watches))

  # Calculate how much memory each "unit" represents
  memory_per_unit=$((available_memory_kb / total_weight))

  sudo sysctl -w fs.inotify.max_user_watches=$((memory_per_unit * default_max_user_watches))
  sudo sysctl -w fs.inotify.max_user_instances=$((memory_per_unit * default_max_user_instances))
  sudo sysctl -w fs.inotify.max_queued_events=$((memory_per_unit * default_max_queued_events))
}

function install_kubectl() {
  if [ -x "$(command -v kubectl)" ]; then
    return
  fi

  curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
  chmod +x kubectl
  sudo mv kubectl /usr/local/bin/kubectl
}

function install_helm() {
  if [ -x "$(command -v helm)" ]; then
    return
  fi

  local tmp_file=$(mktemp)

  curl -fsSL -o $tmp_file https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3
  chmod 700 $tmp_file
  $tmp_file
  rm $tmp_file
}

function install_kind() {
  # If Kind is not installed, install it
  if [ -x "$(command -v kind)" ]; then
    return
  fi

  if [ $(uname -m) = x86_64 ]; then
    url="https://kind.sigs.k8s.io/dl/v0.23.0/kind-linux-amd64"
  elif [ $(uname -m) = arm64 ]; then
    url="https://kind.sigs.k8s.io/dl/v0.23.0/kind-linux-arm64"
  fi

  [ $(uname -m) = x86_64 ] && curl -Lo ./kind $url
  chmod +x ./kind
  sudo mv ./kind /usr/local/bin/kind
}

function install_jq() {
  if [ -x "$(command -v jq)" ]; then
    return
  fi

  if [ -x "$(command -v apt-get)" ]; then
    sudo apt-get install jq -y
  elif [ -x "$(command -v dnf)" ]; then
    sudo dnf install jq -y
  fi
}

function install_and_configure_dnsmasq() {
  # Make systemd-resolved no longer listen on 127.0.0.1:53
  update_resolved_conf
  sudo systemctl restart systemd-resolved

  # Install dnsmasq, we ignore error here since it doesn't matter (it will fail because port 53 is already in use)
  if [ -x "$(command -v apt-get)" ]; then
    sudo apt-get install dnsmasq -y > /dev/null 2>&1
  elif [ -x "$(command -v dnf)" ]; then
    sudo dnf install dnsmasq -y > /dev/null 2>&1
  fi  
  
  # Make dnsmasq fallback to 127.0.0.2:53
  update_dnsmasq_conf
  update_default_dnsmasq

  sudo systemctl restart dnsmasq
}

function add_dns_masq_entry() {
  # If file /etc/dnsmasq.d/50-go-deploy-dev.conf does not exist, create it
  if ! [ -f "/etc/dnsmasq.d/50-go-deploy-dev.conf" ]; then
    echo "address=/deploy.localhost/127.0.0.1" | sudo tee -a /etc/dnsmasq.d/50-go-deploy-dev.conf
  fi

  sudo systemctl restart dnsmasq
}

function wait_for_dns() {
  while [ "$(dig +short deploy.localhost)" == "" ]; do
    sleep 5
  done
}

function generate_cluster_config() {
  ingress_http_port=$((RANDOM % 2768 + 30000))
  ingress_https_port=$((RANDOM % 2768 + 30000))
  mongo_db_port=$((RANDOM % 2768 + 30000))
  redis_port=$((RANDOM % 2768 + 30000))
  nfs_port=$((RANDOM % 2768 + 30000))
  harbor_port=$((RANDOM % 2768 + 30000))
  keycloak_port=$((RANDOM % 2768 + 30000))

  # Use 25 ports for the range starting at a random port in range 30000-32767
  port_range_start=$((RANDOM % 2768 + 30000))
  port_range_end=$((port_range_start + 25))
 
  # Write to cluster-config.rc
  echo -e "#!/bin/bash
# Cluster configuration
export cluster_name=$CLUSTER_NAME
export kubeconfig_output_path=../../kube

# Domain configuration
export domain=deploy.localhost

# Placeholder git repo
export placeholder_git_repo="https://github.com/kthcloud/go-deploy-placeholder.git"
export vm_image="https://cloud-images.ubuntu.com/noble/current/noble-server-cloudimg-amd64.img"

# NFS configuration
export nfs_base_path="/nfs"
export nfs_cluster_ip="10.96.200.2"

# IAM configuration
export keycloak_deploy_storage_secret=

# Ports configuration
export ingress_http_port=$ingress_http_port
export ingress_https_port=$ingress_https_port
export mongo_db_port=$mongo_db_port
export redis_port=$redis_port
export nfs_port=$nfs_port
export harbor_port=$harbor_port
export keycloak_port=$keycloak_port
export port_range_start=$port_range_start
export port_range_end=$port_range_end" > ./cluster-config.rc  
}

function read_cluster_config() {
  source ./cluster-config.rc
}

function generate_kind_cluster_config() {
  read_cluster_config

  # Generate kind config with correct ports
  config="kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  extraPortMappings:"

  # Add port mappings for all services that need to be exposed
  config="$config
  - containerPort: $ingress_http_port
    hostPort: $ingress_http_port
    listenAddress: 0.0.0.0
    protocol: TCP
  - containerPort: $ingress_https_port
    hostPort: $ingress_https_port
    listenAddress: 0.0.0.0
    protocol: TCP
  - containerPort: $mongo_db_port
    hostPort: $mongo_db_port
    listenAddress: 0.0.0.0
    protocol: TCP
  - containerPort: $redis_port
    hostPort: $redis_port
    listenAddress: 0.0.0.0
    protocol: TCP
  - containerPort: $nfs_port
    hostPort: $nfs_port
    listenAddress: 0.0.0.0
    protocol: TCP
  - containerPort: $harbor_port
    hostPort: $harbor_port
    listenAddress: 0.0.0.0
    protocol: TCP
  - containerPort: $keycloak_port
    hostPort: $keycloak_port
    listenAddress: 0.0.0.0
    protocol: TCP"

  for port in $(seq $port_range_start $port_range_end); do
    config="$config
  - containerPort: $port
    hostPort: $port
    listenAddress: 0.0.0.0
    protocol: TCP
  - containerPort: $port
    hostPort: $port
    listenAddress: 0.0.0.0
    protocol: UDP"
  done

  data_dir="${HOME}/go-deploy-data/${cluster_name}"
  mkdir -p $data_dir

  config="$config
  extraMounts:
  - hostPath: "$data_dir"
    containerPath: /mnt/nfs"
  
  echo "$config" > ./manifests/kind-config.yml
}

function create_kind_cluster() {
  read_cluster_config

  local current=$(kind get clusters 2> /dev/stdout | grep -c $cluster_name)
  if [ "$current" -eq 0 ]; then
    generate_kind_cluster_config

    export KUBECONFIG=$KUBECONFIG_PATH
    kind create cluster --name $cluster_name --config ./manifests/kind-config.yml --quiet
  
    # Wait for cluster to be up
    while [ "$(kubectl get nodes 2> /dev/stdout | grep -c Ready)" -lt 1 ]; do
      echo -e "Waiting for cluster to be up"
      echo -e ""
      kubectl get nodes
      sleep $WAIT_SLEEP
    done
  fi

  # Ensure that context is set to the correct cluster
  kubectl config use-context kind-$cluster_name

  # Wait for kubeconfig to change
  while [ "$(kubectl config current-context)" != "kind-$cluster_name" ]; do
    sleep 5
  done

  version=$(kubectl version --output=json 2> /dev/stdout | jq -r '.serverVersion.gitVersion' 2> /dev/null)
  echo -e "Cluster $cluster_name created (version: $version)"

  # Copy kubeconfig to local folder
  if [ ! -d $kubeconfig_output_path ]; then
    mkdir -p $kubeconfig_output_path
  fi

  # If already exists, remove it
  if [ -f "$kubeconfig_output_path/$cluster_name.yml" ]; then
    rm -f "$kubeconfig_output_path/$cluster_name.yml"
  fi

  kind get kubeconfig --name $cluster_name > "$kubeconfig_output_path/$cluster_name.yml"
}

function install_nfs_server() {
  read_cluster_config

  res=$(kubectl get ns 2> /dev/stdout | grep -c nfs-server)
  if [ $res -eq 0 ]; then
    nfs_server_values_subst=$(mktemp)
    export nfs_cluster_ip=$nfs_cluster_ip
    envsubst < ./manifests/nfs-server.yml > $nfs_server_values_subst
    kubectl apply -f $nfs_server_values_subst

    sleep 3
  fi

  # Wait for NFS server to be up
  while [ "$(kubectl get pod -n nfs-server -l app=nfs-server -o jsonpath="{.items[0].status.phase}" 2> /dev/stdout)" != "Running" ]; do
    echo -e "Waiting for NFS server to be up"
    echo -e ""
    kubectl get pod -n nfs-server
    echo -e ""
    sleep $WAIT_SLEEP
  done

  # Create subfolders deployments, vms, scratch and snapshots
  pod=$(kubectl get pod -n nfs-server -l app=nfs-server -o jsonpath="{.items[0].metadata.name}")
  kubectl exec -n nfs-server $pod -- mkdir -p  /exports/$nfs_base_path/deployments /exports/$nfs_base_path/vms /exports/$nfs_base_path/scratch /exports/$nfs_base_path/snapshots /exports/$nfs_base_path/misc
}

function install_nfs_csi() {
  read_cluster_config

  # If deployment 'csi-nfs-controller' in namespace 'kube-system' already exists, skip
  res=$(kubectl get deploy -n kube-system | grep -c csi-nfs-controller)
  if [ $res -eq 0 ]; then
    helm install csi-driver-nfs csi-driver-nfs \
      --repo https://raw.githubusercontent.com/kubernetes-csi/csi-driver-nfs/master/charts \
      --namespace kube-system \
      --version v4.6.0 \
      --set controller.dnsPolicy=ClusterFirstWithHostNet \
      --set node.dnsPolicy=ClusterFirstWithHostNet
  fi

  # If deploy-misc already exists, skip
  res=$(kubectl get deploy > /dev/stdout 2>&1 | grep -c deploy-misc)
  if [ $res -eq 0 ]; then
    sc_subst=$(mktemp)
    export nfs_server="nfs-server.nfs-server.svc.cluster.local"
    export nfs_share="$nfs_base_path/misc"
    envsubst < ./manifests/sc-misc.yml > $sc_subst
    kubectl apply -f $sc_subst
  fi

  # Ensure that the storage class 'deploy-misc' is set as the default storage class
  # First check if there is a default storage class, and unset it
  default_sc=$(kubectl get sc | grep -c "(default)")
  if [ $default_sc -ne 0 ]; then
    kubectl patch storageclass $(kubectl get sc | grep "(default)" | awk '{print $1}') -p '{"metadata": {"annotations": {"storageclass.kubernetes.io/is-default-class": "false"}}}'
  fi

  # Then set the new default storage class
  kubectl patch storageclass deploy-misc -p '{"metadata": {"annotations": {"storageclass.kubernetes.io/is-default-class": "true"}}}'

  # Wait for NFS CSI to be up
  while [ "$(kubectl get pod -n kube-system -l app=csi-nfs-controller -o jsonpath="{.items[0].status.phase}" 2> /dev/stdout)" != "Running" ]; do
    echo -e "Waiting for NFS CSI to be up"
    echo -e ""
    kubectl get pod -n kube-system
    echo -e ""
    sleep $WAIT_SLEEP
  done
}

function install_ingress_nginx() {
  read_cluster_config

  # If release 'ingress-nginx' in namespace 'ingress-nginx' already exists, skip
  res=$(helm list -n ingress-nginx | grep -c ingress-nginx)
  if [ $res -eq 0 ]; then
    helm upgrade --install ingress-nginx ingress-nginx \
      --repo https://kubernetes.github.io/ingress-nginx \
      --namespace ingress-nginx --create-namespace \
      --set controller.service.nodePorts.http=$ingress_http_port \
      --set controller.service.nodePorts.https=$ingress_https_port \
      --values - <<EOF
controller:
  ingressClassResource:
    default: "true"
  config:
    allow-snippet-annotations: "true"
    proxy-buffering: "on"
    proxy-buffers: 4 "512k"
    proxy-buffer-size: "256k"
EOF
  fi

  # Wait for ingress-nginx to be up
  while [ "$(kubectl get pod -n ingress-nginx -l app.kubernetes.io/name=ingress-nginx -o jsonpath="{.items[0].status.phase}" 2> /dev/stdout)" != "Running" ]; do
    echo -e "Waiting for ingress-nginx to be up"
    echo -e ""
    kubectl get pod -n ingress-nginx
    echo -e ""
    sleep $WAIT_SLEEP
  done
}

function install_harbor() {
  read_cluster_config

  # If helm release 'harbor' in namespace 'harbor' already exists, skip
  res=$(helm list -n harbor | grep -c harbor)
  if [ $res -eq 0 ]; then
    harbor_values_subst=$(mktemp)
    envsubst < ./helmvalues/harbor.values.yml > $harbor_values_subst

    helm repo add harbor https://helm.goharbor.io
    helm install harbor harbor/harbor \
      --namespace harbor \
      --create-namespace \
      --version v1.14.2 \
      --values $harbor_values_subst

    # Allow namespace to be created
    sleep 5
  fi

  # Wait for Harbor to be up
  while [ "$(curl -s -o /dev/null -w "%{http_code}" http://$domain:$harbor_port)" != "200" ]; do
    echo -e "Waiting for Harbor to be up"
    echo -e ""
    kubectl get pod -n harbor
    echo -e ""
    sleep $WAIT_SLEEP
  done

  # Setup an ingress for Harbor
  res=$(kubectl get ingress -n harbor -o yaml | grep -c harbor)
  if [ $res -eq 0 ]; then
    export domain=$domain
    export harbor_port=$harbor_port
    harbor_subst=$(mktemp)
    envsubst < ./manifests/harbor.yml > $harbor_subst
    kubectl apply -f $harbor_subst
  fi
}

function seed_harbor_with_images() {
  read_cluster_config

  kubectl get pods -n harbor

  # local url="http://harbor.$domain:$ingress_http_port"
  local url="http://localhost:$harbor_port"
  local domain="localhost:$harbor_port"
  local user="admin"
  local password="Harbor12345"

  local robot_user="$user"
  local robot_password="$password"

  # If repository "go-deploy-placeholder" in project "library" already exists, skip
  res=$(curl -s -u $user:$password -X GET $url/api/v2.0/projects/library/repositories | jq -r '.[] | select(.name=="library/go-deploy-placeholder") | .name')
  if [ "$res" == "library/go-deploy-placeholder" ]; then
    return
  fi


  kubectl get pods -n harbor

  # Download repo and build the image
  if [ ! -d "go-deploy-placeholder" ]; then
    git clone $placeholder_git_repo --quiet
  fi


  kubectl get pods -n harbor

  # Use 'library' so we don't need to create our own (library is the default namespace in Harbor)
  docker build go-deploy-placeholder/ -t $domain/library/go-deploy-placeholder:latest 2> /dev/null
  docker login $domain -u $robot_user -p $robot_password 2> /dev/null
  docker push $domain/library/go-deploy-placeholder:latest 2> /dev/null


  kubectl get pods -n harbor

  # Remove the placeholder repo
  rm -rf go-deploy-placeholder
}

function install_mongodb() {
  read_cluster_config

  # If namespace 'mongodb' already exists, skip
  res=$(kubectl get ns | grep -c mongodb)
  if [ $res -eq 0 ]; then
    mongodb_values_subst=$(mktemp)
    envsubst < ./manifests/mongodb.yml > $mongodb_values_subst
    kubectl apply -f $mongodb_values_subst
  fi

  # Wait for MongoDB to be up
  while [ "$(kubectl get pod -n mongodb -l app=mongodb -o jsonpath="{.items[0].status.phase}" 2> /dev/stdout)" != "Running" ]; do
    echo -e "Waiting for MongoDB to be up"
    echo -e ""
    kubectl get pod -n mongodb
    echo -e ""
    sleep $WAIT_SLEEP
  done
}

function install_redis() {
  read_cluster_config
  
  # If namespace 'redis' already exists, skip
  res=$(kubectl get ns | grep -c redis)
  if [ $res -eq 0 ]; then
    redis_values_subst=$(mktemp)
    envsubst < ./manifests/redis.yml > $redis_values_subst
    kubectl apply -f $redis_values_subst
  fi

  # Wait for Redis to be up
  while [ "$(kubectl get pod -n redis -l app=redis -o jsonpath="{.items[0].status.phase}" 2> /dev/stdout)" != "Running" ]; do
    echo -e "Waiting for Redis to be up"
    echo -e ""
    kubectl get pod -n redis
    echo -e ""

    # If ErrImagePull or ImagePullBackOff in the status, do kubectl describe deployment redis
    res=$(kubectl get pod -n redis)
    if [ "$(echo $res | grep -c ErrImagePull)" -ne 0 ] || [ "$(echo $res | grep -c ImagePullBackOff)" -ne 0 ]; then
      echo -e ""
      kubectl describe deployment redis -n redis
      echo -e ""
    fi

    sleep $WAIT_SLEEP
  done
}

function install_keycloak() {
  read_cluster_config

  # If namespace 'keycloak' already exists, skip
  res=$(kubectl get ns | grep -c keycloak)
  if [ $res -eq 0 ]; then
    keycloak_values_subst=$(mktemp)
    envsubst < ./manifests/keycloak.yml > $keycloak_values_subst
    kubectl apply -f $keycloak_values_subst
  fi

  rm -f keycloak.values.yml

  # Wait for Keycloak to be up
  while [ "$(curl -s -o /dev/null -w "%{http_code}" http://keycloak.$domain:$keycloak_port/health/ready)" != "200" ]; do
    echo -e "Waiting for Keycloak to be up"
    echo -e ""
    kubectl get pod -n keycloak
    echo -e ""
    sleep $WAIT_SLEEP
  done

  # Fetch admin token
  local token=$(curl -s \
    -X POST \
    -H "Content-Type: application/x-www-form-urlencoded" \
    -d "client_id=admin-cli&username=admin&password=admin&grant_type=password" \
    http://keycloak.$domain:$keycloak_port/realms/master/protocol/openid-connect/token \
    | jq -r '.access_token')
  
  # Check if go-deploy client exists, if not create it
  local check_exists=$(curl -s \
    -H "Content-Type: application/json" \
    -H \"Accept: application/json\" \
    -H "Authorization: Bearer $token" \
    -X GET http://keycloak.$domain:$keycloak_port/admin/realms/master/clients?clientId=go-deploy)
  local exists=$(echo $check_exists | jq -r '.[] | select(.clientId=="go-deploy") | .clientId')
  if [ "$exists" != "go-deploy" ]; then
    local payload='{
      "protocol":"openid-connect",
      "clientId":"go-deploy",
      "name":"go-deploy",
      "description":"go-deploy",
      "publicClient":true,
      "authorizationServicesEnabled":false,
      "serviceAccountsEnabled":true,
      "implicitFlowEnabled":false,
      "directAccessGrantsEnabled":true,
      "standardFlowEnabled":true,
      "frontchannelLogout":true,
      "attributes":{"saml_idp_initiated_sso_url_name":"","oauth2.device.authorization.grant.enabled":false,"oidc.ciba.grant.enabled":false},
      "alwaysDisplayInConsole":false,
      "rootUrl":"",
      "baseUrl":"",
      "redirectUris":["http://*", "https://*"]
      }'
    curl -s \
      -H "Content-Type: application/json" \
      -H "Accept: application/json" \
      -H "Authorization: Bearer $token" \
      -X POST http://keycloak.$domain:$keycloak_port/admin/realms/master/clients -d "$payload"
  fi

  # Check if go-deploy-storage client exists, if not create it
  local check_exists=$(curl -s \
    -H "Content-Type: application/json" \
    -H \"Accept: application/json\" \
    -H "Authorization: Bearer $token" \
    -X GET http://keycloak.$domain:$keycloak_port/admin/realms/master/clients?clientId=go-deploy-storage)
  local exists=$(echo $check_exists | jq -r '.[] | select(.clientId=="go-deploy-storage") | .clientId')
  if [ "$exists" != "go-deploy-storage" ]; then
    local payload='{
      "protocol":"openid-connect",
      "clientId":"go-deploy-storage",
      "name":"go-deploy-storage",
      "description":"go-deploy-storage",
      "publicClient":false,
      "authorizationServicesEnabled":false,
      "serviceAccountsEnabled":true,
      "implicitFlowEnabled":false,
      "directAccessGrantsEnabled":true,
      "standardFlowEnabled":true,
      "frontchannelLogout":true,
      "attributes":{"saml_idp_initiated_sso_url_name":"","oauth2.device.authorization.grant.enabled":false,"oidc.ciba.grant.enabled":false},
      "alwaysDisplayInConsole":false,
      "rootUrl":"",
      "baseUrl":"",
      "redirectUris":["http://*", "https://*"]
      }'
    curl -s \
      -H "Content-Type: application/json" \
      -H "Accept: application/json" \
      -H "Authorization: Bearer $token" \
      -X POST http://keycloak.$domain:$keycloak_port/admin/realms/master/clients -d "$payload"
  fi

  # Fetch created client's secret
  keycloak_deploy_storage_secret=$(curl -s \
    -H "Content-Type: application/json" \
    -H "Accept: application/json" \
    -H "Authorization: Bearer $token" \
    -X GET http://keycloak.$domain:$keycloak_port/admin/realms/master/clients?clientId=go-deploy-storage \
    | jq -r '.[0].secret')

  # Create necessary groups
  groups=("admin" "base" "power")
  for group in "${groups[@]}"; do
    local check_exists=$(curl -s \
      -H "Content-Type: application/json" \
      -H "Accept: application/json" \
      -H "Authorization: Bearer $token" \
      -X GET http://keycloak.$domain:$keycloak_port/admin/realms/master/groups?search=$group)
    local exists=$(echo $check_exists | jq -r '.[] | select(.name=="'$group'") | .name')
    if [ "$exists" != "$group" ]; then
      local payload='{"name":"'$group'"}'
      curl -s \
        -H "Content-Type: application/json" \
        -H "Accept: application/json" \
        -H "Authorization: Bearer $token" \
        -X POST http://keycloak.$domain:$keycloak_port/admin/realms/master/groups -d "$payload"
    fi
  done

  # Create an admin user, base user and power user
  users=("admin" "base" "power")
  for user in "${users[@]}"; do
    local check_exists=$(curl -s \
      -H "Content-Type: application/json" \
      -H "Accept: application/json" \
      -H "Authorization: Bearer $token" \
      -X GET http://keycloak.$domain:$keycloak_port/admin/realms/master/users?search=$user)
    local exists=$(echo $check_exists | jq -r '.[] | select(.username=="'$user'") | .username')
    if [ "$exists" != "$user" ]; then
      local payload='{"username":"'$user'","enabled":true,"emailVerified":true,"firstName":"'$user'","lastName":"'$user'","email":"'$user'@'$domain'","credentials":[{"type":"password","value":"'$user'","temporary":false}]}'
      curl -s \
        -H "Content-Type: application/json" \
        -H "Accept: application/json" \
        -H "Authorization: Bearer $token" \
        -X POST http://keycloak.$domain:$keycloak_port/admin/realms/master/users -d "$payload"
    fi
    
    # Add user to group
    local user_id=$(curl -s \
      -H "Content-Type: application/json" \
      -H "Accept: application/json" \
      -H "Authorization: Bearer $token" \
      -X GET http://keycloak.$domain:$keycloak_port/admin/realms/master/users?search=$user \
      | jq -r '.[0].id')

    local group_id=$(curl -s \
      -H "Content-Type: application/json" \
      -H "Accept: application/json" \
      -H "Authorization: Bearer $token" \
      -X GET http://keycloak.$domain:$keycloak_port/admin/realms/master/groups?search=$user \
      | jq -r '.[0].id')

    curl -s \
      -H "Content-Type: application/json" \
      -H "Accept: application/json" \
      -H "Authorization: Bearer $token" \
      -X PUT http://keycloak.$domain:$keycloak_port/admin/realms/master/users/$user_id/groups/$group_id
  done

  # Add "groups" protocol mapper to clients

  local deploy_client_id=$(curl -s \
    -H "Content-Type: application/json" \
    -H "Accept: application/json" \
    -H "Authorization: Bearer $token" \
    -X GET http://keycloak.$domain:$keycloak_port/admin/realms/master/clients?clientId=go-deploy \
    | jq -r '.[0].id')

  local deploy_storage_client_id=$(curl -s \
    -H "Content-Type: application/json" \
    -H "Accept: application/json" \
    -H "Authorization: Bearer $token" \
    -X GET http://keycloak.$domain:$keycloak_port/admin/realms/master/clients?clientId=go-deploy-storage \
    | jq -r '.[0].id')

  # Create groups mapping using: http://keycloak.deploy.localhost:31125/admin/realms/master/clients/b829f2ad-eb13-45f4-bf03-320fdd14ffe9/protocol-mappers/models
  local groups_mapping='{
    "name": "groups",
    "protocol": "openid-connect",
    "protocolMapper": "oidc-group-membership-mapper",
    "consentRequired": false,
    "config": {
      "access.token.claim": "true",
      "claim.name": "groups",
      "full.path": "false",
      "id.token.claim": "true",
      "introspection.token.claim": "true",
      "lightweight.claims": "false",
      "userinfo.token.claim": "true"
    }
  }'

  # Fetch protocol mappers for deploy client from http://keycloak.deploy.localhost:31125/admin/realms/master/clients/e7adf649-8bd2-473b-a19b-a74ce3a7abca

  local res=$(curl -s \
    -H "Content-Type: application/json" \
    -H "Accept: application/json" \
    -H "Authorization: Bearer $token" \
    -X GET http://keycloak.$domain:$keycloak_port/admin/realms/master/clients/$deploy_client_id \
    | jq -r '.protocolMappers[] | select(.name=="groups") | .name' | grep -c groups)
  if [ $res -eq 0 ]; then
    curl -s \
      -H "Content-Type: application/json" \
      -H "Accept: application/json" \
      -H "Authorization: Bearer $token" \
      -X POST http://keycloak.$domain:$keycloak_port/admin/realms/master/clients/$deploy_client_id/protocol-mappers/models -d "$groups_mapping"
  fi

  local res=$(curl -s \
    -H "Content-Type: application/json" \
    -H "Accept: application/json" \
    -H "Authorization: Bearer $token" \
    -X GET http://keycloak.$domain:$keycloak_port/admin/realms/master/clients/$deploy_storage_client_id \
    | jq -r '.protocolMappers[] | select(.name=="groups") | .name' | grep -c groups)
  if [ $res -eq 0 ]; then
    curl -s \
      -H "Content-Type: application/json" \
      -H "Accept: application/json" \
      -H "Authorization: Bearer $token" \
      -X POST http://keycloak.$domain:$keycloak_port/admin/realms/master/clients/$deploy_storage_client_id/protocol-mappers/models -d "$groups_mapping"
  fi
  
  # Write keycloak_deploy_storage_secret to cluster-config.rc
  # Overwrite if the row already exists
  sed -i "/export keycloak_deploy_storage_secret=/c\export keycloak_deploy_storage_secret=$keycloak_deploy_storage_secret" ./cluster-config.rc


  # Finally, we need to add a DNS record that points the keycloak name to the node's IP
  # This is required since the name can't be resolved properly inside the cluster (and we use a NodePort)
  dns_record="rewrite name keycloak.$domain $cluster_name-control-plane"
  configmap=$(kubectl get configmap coredns -n kube-system -o json)

  echo -e $configmap

  if ! echo "${configmap}" | grep -q "${dns_record}"; then
    echo -e "Adding DNS record for keycloak.$domain -> $cluster_name-control-plane"
    corefile=$(echo "${configmap}" | jq -r '.data.Corefile')
    new_corefile=$(echo "${corefile}" | sed "/^\\s*forward/ i \    ${dns_record}")
    kubectl patch configmap coredns -n kube-system --type merge -p "$(jq -n --arg corefile "${new_corefile}" '{data: {Corefile: $corefile}}')"

    # Restart coredns
    kubectl rollout restart deployment coredns -n kube-system
  fi
}

function install_cert_manager() {
  read_cluster_config

  # If release 'cert-manager' in namespace 'cert-manager' already exists, skip
  res=$(helm list -n cert-manager | grep -c cert-manager)
  if [ $res -eq 0 ]; then
    helm upgrade --install \
      cert-manager \
      cert-manager \
      --repo https://charts.jetstack.io \
      --namespace cert-manager \
      --create-namespace \
      --version v1.14.4 \
      --set 'extraArgs={--dns01-recursive-nameservers-only,--dns01-recursive-nameservers=8.8.8.8:53\,1.1.1.1:53}' \
      --set installCRDs=true
  fi

  # Wait for cert-manager to be up
  while [ "$(kubectl get pod -n cert-manager -l app=cert-manager -o jsonpath="{.items[0].status.phase}" 2> /dev/stdout)" != "Running" ]; do
    echo -e "Waiting for cert-manager to be up"
    echo -e ""
    kubectl get pod -n cert-manager
    echo -e ""
    sleep $WAIT_SLEEP
  done
  
  # If clusterIssuer go-deploy-cluster-issuer already exists, skip
  res=$(kubectl get clusterissuer 2>/dev/stdout | grep -c go-deploy-cluster-issuer)
  if [ $res -eq 0 ]; then
    cert_manager_subst=$(mktemp)
    envsubst < ./manifests/cert-manager.yml > $cert_manager_subst
    kubectl apply -f $cert_manager_subst
  fi
}

function install_hairpin_proxy() {
  read_cluster_config

  # If namespace 'hairpin-proxy' already exists, skip
  res=$(kubectl get ns | grep -c hairpin-proxy)
  if [ $res -eq 0 ]; then
    kubectl apply -f https://raw.githubusercontent.com/compumike/hairpin-proxy/v0.2.1/deploy.yml
  fi
}

function install_storage_classes() {
  read_cluster_config

  # Install CRDs if not already installed, we assume if one does not exist, none of them do
  res=$(kubectl get crd | grep -c volumesnapshots.snapshot.storage.k8s.io)
  if [ $res -eq 0 ]; then
    kubectl apply -f https://raw.githubusercontent.com/kubernetes-csi/external-snapshotter/master/client/config/crd/snapshot.storage.k8s.io_volumesnapshots.yaml
    kubectl apply -f https://raw.githubusercontent.com/kubernetes-csi/external-snapshotter/master/client/config/crd/snapshot.storage.k8s.io_volumesnapshotcontents.yaml
    kubectl apply -f https://raw.githubusercontent.com/kubernetes-csi/external-snapshotter/master/client/config/crd/snapshot.storage.k8s.io_volumesnapshotclasses.yaml
    kubectl apply -f https://raw.githubusercontent.com/kubernetes-csi/external-snapshotter/master/deploy/kubernetes/snapshot-controller/rbac-snapshot-controller.yaml
    kubectl apply -f https://raw.githubusercontent.com/kubernetes-csi/external-snapshotter/master/deploy/kubernetes/snapshot-controller/setup-snapshot-controller.yaml
  fi

  export nfs_server="nfs-server.nfs-server.svc.cluster.local"

  # If storage class 'deploy-vm-disks' does not exist, create it
  export nfs_share="$nfs_base_path/vms"
  res=$(kubectl get sc 2>/dev/null | grep -c "deploy-vm-disks")
  if [ $res -eq 0 ]; then
    envsubst < ./manifests/sc-vm-disks.yml | kubectl apply -f -
  fi

  # If storage class 'deploy-vm-scratch' does not exist, create it
  export nfs_share="$nfs_base_path/scratch"
  res=$(kubectl get sc 2>/dev/null | grep -c "deploy-vm-scratch")
  if [ $res -eq 0 ]; then
    envsubst < ./manifests/sc-vm-scratch.yml | kubectl apply -f -
  fi

  # If volume snapshot class 'deploy-vm-snapshots' does not exist, create it
  export nfs_share="$nfs_base_path/snapshots"
  res=$(kubectl get volumesnapshotclass 2>/dev/null | grep -c "deploy-vm-snapshots")
  if [ $res -eq 0 ]; then
    envsubst < ./manifests/vsc-vm-snapshots.yml | kubectl apply -f -
  fi
}

function install_kubevirt() {
  read_cluster_config

  # If namespace 'kubevirt' already exists, skip
  res=$(kubectl get ns | grep -c kubevirt)
  if [ $res -eq 0 ]; then
    export version=$(curl -s https://storage.googleapis.com/kubevirt-prow/release/kubevirt/kubevirt/stable.txt)
    kubectl create -f https://github.com/kubevirt/kubevirt/releases/download/$version/kubevirt-operator.yaml
    kubectl create -f https://github.com/kubevirt/kubevirt/releases/download/$version/kubevirt-cr.yaml
  fi
  
  while [ "$(kubectl get kubevirt.kubevirt.io/kubevirt -n kubevirt -o=jsonpath="{.status.phase}")" != "Deployed" ]; do
    echo -e "Waiting for KubeVirt to be up"
    echo -e ""
    kubectl get pod -n kubevirt
    echo -e ""
    sleep $WAIT_SLEEP
  done

  # Add feature gates DateVolumes, LiveMigration, GPU and Snapshot (spec.configuration.developerConfiguration.featureGates)
  if [ "$(kubectl get kubevirt.kubevirt.io/kubevirt -n kubevirt -o=jsonpath="{.spec.configuration.developerConfiguration.featureGates}")" == "" ]; then
    kubectl patch kubevirt.kubevirt.io/kubevirt -n kubevirt --type='json' -p='[{"op": "add", "path": "/spec/configuration/developerConfiguration/featureGates", "value": []}]'
  fi

  # Add feature gates
  feature_gates=("DataVolumes" "GPU" "Snapshot")
  for feature in "${feature_gates[@]}"; do
    if [[ "$(kubectl get kubevirt.kubevirt.io/kubevirt -n kubevirt -o=jsonpath="{.spec.configuration.developerConfiguration.featureGates}")" != *"$feature"* ]]; then
      kubectl patch kubevirt.kubevirt.io/kubevirt -n kubevirt --type='json' -p='[{"op": "add", "path": "/spec/configuration/developerConfiguration/featureGates/-", "value": "'$feature'"}]'
    fi
  done
}

function install_cdi() {
  read_cluster_config

  # If namespace 'cdi' already exists, skip
  res=$(kubectl get ns | grep -c cdi)
  if [ $res -eq 0 ]; then
    export tag=$(curl -s -w %{redirect_url} https://github.com/kubevirt/containerized-data-importer/releases/latest)
    export version=$(echo ${tag##*/})
    kubectl apply -f https://github.com/kubevirt/containerized-data-importer/releases/download/$version/cdi-operator.yaml
    kubectl apply -f https://github.com/kubevirt/containerized-data-importer/releases/download/$version/cdi-cr.yaml
  fi

  # Ensure that spec.config.scratchSpaceStorageClass: deploy-vm-scratch, if not set it
  if [ "$(kubectl get cdi -n cdi -o=jsonpath="{.items[0].spec.config.scratchSpaceStorageClass}")" != "deploy-vm-scratch" ]; then
    kubectl patch cdi cdi -n cdi --type='json' -p='[{"op": "replace", "path": "/spec/config/scratchSpaceStorageClass", "value": "deploy-vm-scratch"}]'
  fi
}

function install_kubemacpool() {
  # If namespace 'kubemacpool-system' already exists, skip
  res=$(kubectl get ns | grep -c kubemacpool-system)
  if [ $res -eq 0 ]; then
    local tmp_file=$(mktemp)
    wget https://raw.githubusercontent.com/k8snetworkplumbingwg/kubemacpool/master/config/release/kubemacpool.yaml -O $tmp_file
    sed -i "s/02:00:00:00:00:00/02:5e:6c:00:00:00/" $tmp_file
    sed -i "s/02:FF:FF:FF:FF:FF/02:5e:6c:FF:FF:FF/" $tmp_file

    kubectl apply -f $tmp_file
    rm -f $tmp_file
  fi
}

function print_result() {
  read_cluster_config

  echo -e ""
  echo -e "dnsmasq is used to allow the names to resolve. See the following guides for help configuring it:"
  echo -e " - WSL2 (Windows): https://github.com/absolunet/pleaz/blob/production/documentation/installation/wsl2/dnsmasq.md"
  echo -e " - systemd-resolved (Linux): https://gist.github.com/frank-dspeed/6b6f1f720dd5e1c57eec8f1fdb2276df"
  echo -e ""
  echo -e "The following services are now available:"
  echo -e " - ${BLUE_BOLD}Harbor${RESET}: http://127.0.0.1:$harbor_port (admin:Harbor12345)"
  echo -e " - ${TEAL_BOLD}Keycloak${RESET}: http://keycloak.$domain:$keycloak_port (admin:admin)"
  echo -e "      Users: admin:admin, base:base, power:power"
  echo -e "      Clients: go-deploy:(no secret), go-deploy-storage:$keycloak_deploy_storage_secret"
  echo -e " - ${GREEN_BOLD}MongoDB${RESET}: mongodb://admin:admin@localhost:$mongo_db_port"
  echo -e " - ${RED_BOLD}Redis${RESET}: redis://localhost:$redis_port"
  echo -e " - ${ORANGE_BOLD}NFS${RESET}: nfs://localhost:$nfs_port"
  echo -e ""
  echo -e "To start the application, go the the top directory and run the following command:"
  echo -e ""
  echo -e "    ${WHITE_BOLD}DEPLOY_CONFIG_FILE=config.local.yml go run main.go [--mode dev|test|prod] [--<worker name>]${RESET}"
  echo -e ""
  echo -e "Happy coding! 🚀"
  echo -e ""
}

parse_flags $@

check_dependencies

if [ ! -f "./cluster-config.rc" ]; then
  generate_cluster_config
fi

read_cluster_config
if [ "$cluster_name" != "$CLUSTER_NAME" ]; then
  echo -e "$RED_CROSS Another local cluster is already running, and multiple local clusters are not yet supported"
  exit 1
fi

# Pre-requisites

# If CONFIGURE_DNS is true, configure dnsmasq
if [ "$CONFIGURE_DNS" == "true" ]; then
  run_task "Install and configure dnsmasq" install_and_configure_dnsmasq
fi

run_task "Increase max file watchers" increase_max_file_watchers
run_task "Install kubectl" install_kubectl
run_task "Install helm" install_helm
run_task "Install kind" install_kind
run_task "Install jq" install_jq
run_task "Add dnsmasq entry" add_dns_masq_entry
run_task "Waiting for DNS" wait_for_dns

# Base
run_task "Set up kind cluster" create_kind_cluster
run_task "Install NFS Server" install_nfs_server
run_task "Install NFS CSI" install_nfs_csi

# Apps
run_task "Install Ingress Nginx" install_ingress_nginx
run_task "Install Harbor" install_harbor
run_task "Install MongoDB" install_mongodb
run_task "Install Redis" install_redis
run_task "Install Keycloak" install_keycloak

# Dependencies
run_task "Install Cert Manager" install_cert_manager
run_task "Install Hairpin Proxy" install_hairpin_proxy
run_task "Install Storage Classes" install_storage_classes
run_task "Install KubeVirt" install_kubevirt
run_task "Install CDI" install_cdi
run_task "Install kubemacpool" install_kubemacpool

# Post-install
run_task "Seed Harbor with images" seed_harbor_with_images


# If exists ../../config.local.yml, ask if user want to replace it
read_cluster_config
if [ -f "../../config.local.yml" ]; then

  # Check if SKIP_CONFIRMATIONS
  if [ "$SKIP_CONFIRMATIONS" == "false" ]; then
    echo ""
    read -p "config.local.yml already exists. Do you want to replace it? [y/n]: " -n 1 -r
    echo
  else
    REPLY="y"
  fi
else
  REPLY="y"
fi

# If reply is either y or Y, generate config.local.yml
if [[ "$REPLY" =~ ^[Yy]$ ]]; then
  echo "Generating config.local.yml"
  cp config.yml.tmpl ../../config.local.yml

  # Core
  export external_url="http://localhost:8080"
  export port="8080"

  # Zone
  export deployment_domain="app.$domain"
  export sm_domain="storage.$domain"
  export vm_domain="vm.$domain"
  export vm_app_domain="vm-app.$domain"

  export kubeconfig_path="./kube/$cluster_name.yml"
  export nfs_server=$nfs_cluster_ip
  export nfs_parent_path_app="$nfs_base_path/deployments"
  export nfs_parent_path_vm="$nfs_base_path/vms"
  export port_range_start="$port_range_start"
  export port_range_end="$port_range_end"

  # VM
  export admin_ssh_public_key=$(cat ~/.ssh/id_rsa.pub)
  export vm_image="$vm_image"

  # Deployment


  # Registry
  export registry_url="localhost:$harbor_port"
  export placeholder_image="$registry_url/library/go-deploy-placeholder"

  # Keycloak
  export keycloak_url="http://keycloak.deploy.localhost:$keycloak_port"
  export keycloak_realm="master"
  export keycloak_admin_group="admin"
  export keycloak_storage_client_id="go-deploy-storage"
  export keycloak_storage_client_secret=$keycloak_deploy_storage_secret

  # MongoDB
  export mongodb_url="mongodb://admin:admin@localhost:$mongo_db_port"
  export mongodb_name="deploy"

  # Redis
  export redis_url="localhost:$redis_port"
  export redis_password=

  # Harbor
#    export harbor_url="http://harbor.deploy.localhost:$harbor_port"
  export harbor_url="http://$domain:$harbor_port"
  export harbor_user="admin"
  export harbor_password="Harbor12345"
  export harbor_webhook_secret="secret"

  envsubst < config.yml.tmpl > ../../config.local.yml

  echo -e ""
  echo -e ""
  echo -e "$GREEN_CHECK config.local.yml generated"
else
  echo "Skipping config.local.yml generation"
fi


print_result




