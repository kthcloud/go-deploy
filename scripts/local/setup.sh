#!/bin/bash

# Ensure this script is run from the script folder by checking if the parent folder contains mod.go
if [ ! -f "../../go.mod" ]; then
  echo "$RED_CROSS Please run this script from the scripts folder"
  exit 1
fi

source ./common.sh

domain="deploy.localhost"

cluster_name="go-deploy-dev"
kubeconfig_output_path="../../kube"

placeholder_git_repo="https://github.com/kthcloud/go-deploy-placeholder.git"
vm_image="https://cloud-images.ubuntu.com/noble/current/noble-server-cloudimg-amd64.img"
nfs_base_path="/mnt/nfs"

# Ports
ingress_http_port=9080
ingress_https_port=9443
mongo_db_port=9027
redis_port=9079
nfs_port=9049
harbor_port=9081
port_range_start=29000
port_range_end=30000

# Dynamic variables
keycloak_deploy_secret=""
keycloak_deploy_storage_secret=""


# Check if Docker is installed, if not exit
if ! [ -x "$(command -v docker)" ]; then
  echo -e "$RED_CROSS Docker is not installed. Please install Docker"
  exit 1
fi

# Check if Helm is installed, if not exit
if ! [ -x "$(command -v helm)" ]; then
  echo -e "$RED_CROSS Helm is not installed. Please install Helm"
  exit 1
fi

# Check if dnsmasq is installed, if not exit
if ! [ -x "$(command -v dnsmasq)" ]; then
  echo -e "$RED_CROSS dnsmasq is not installed. Please install dnsmasq"
  exit 1
fi

# Check if /etc/dnsqmasq.d exists, if not exit
if ! [ -d "/etc/dnsmasq.d" ]; then
  echo -e "$RED_CROSS /etc/dnsmasq.d does not exist. This is usually caused by dnsmasq not being installed correctly"
  exit 1
fi

function configure_local_dns() {
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

# If not exists, install k3d
function install_k3d() {
  k3s_install_path="curl -s https://raw.githubusercontent.com/rancher/k3d/main/install.sh | bash"
  if ! [ -x "$(command -v k3d)" ]; then
    eval $k3s_install_path
  fi
}

function create_k3d_cluster() {
  local current=$(k3d cluster list | grep -c $cluster_name)
  if [ "$current" -eq 0 ]; then
    k3d cluster create $cluster_name --agents 2 \
      -p "$ingress_http_port:80@loadbalancer" \
      -p "$ingress_https_port:443@loadbalancer" \
      -p "$port_range_start-$port_range_end,:29000-30000@server:0" \
      -p "$redis_port:6379@server:0" \
      -p "$mongo_db_port:27017@server:0" \
      -p "$nfs_port:2049@server:0" \
      -p "$harbor_port:9081@server:0"
  fi

  # Wait for kubeconfig to change
  while [ "$(kubectl config current-context)" != "k3d-$cluster_name" ]; do
    sleep 5
  done

  # Copy kubeconfig to local folder
  if [ ! -d $kubeconfig_output_path ]; then
    mkdir -p $kubeconfig_output_path
  fi

  # If already exists, remove it
  if [ -f "$kubeconfig_output_path/$cluster_name.yml" ]; then
    rm -f "$kubeconfig_output_path/$cluster_name.yml"
  fi

  k3d kubeconfig get $cluster_name > "$kubeconfig_output_path/$cluster_name.yml"
}

function install_harbor() {
  # If helm release 'harbor' in namespace 'harbor' already exists, skip
  res=$(helm list -n harbor | grep -c harbor)
  if [ $res -eq 0 ]; then
    helm install harbor harbor \
      --repo https://helm.goharbor.io \
      --namespace harbor \
      --create-namespace \
      --values ./helmvalues/harbor.values.yml
  fi

  # Setup an ingress for Harbor
  res=$(kubectl get ingress -n harbor -o yaml | grep -c harbor)
  export domain=$domain
  export harbor_port=$harbor_port
  if [ $res -eq 0 ]; then
    envsubst < ./manifests/harbor.yml | kubectl apply -f -
  fi

  # Wait for Harbor to be up
  while [ "$(curl -s -o /dev/null -w "%{http_code}" http://harbor.$domain:$ingress_http_port)" != "200" ]; do
    sleep 5
  done

}

function seed_harbor_with_images() {
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

  # Download repo and build the image
  if [ ! -d "go-deploy-placeholder" ]; then
    git clone $placeholder_git_repo
  fi

  # Use 'library' so we don't need to create our own (library is the default namespace in Harbor)
  docker build go-deploy-placeholder/ -t $domain/library/go-deploy-placeholder:latest
  docker login $domain -u $robot_user -p $robot_password
  docker push $domain/library/go-deploy-placeholder:latest

  # Remove the placeholder repo
  rm -rf go-deploy-placeholder
}

function install_mongodb() {
  # If namespace 'mongodb' already exists, skip
  res=$(kubectl get ns | grep -c mongodb)
  if [ $res -eq 0 ]; then
    kubectl apply -f ./manifests/mongodb.yml
  fi
}

function install_redis() {
  # If namespace 'redis' already exists, skip
  res=$(kubectl get ns | grep -c redis)
  if [ $res -eq 0 ]; then
    kubectl apply -f ./manifests/redis.yml
  fi
}

function install_keycloak() {
  # If namespace 'keycloak' already exists, skip
  res=$(kubectl get ns | grep -c keycloak)
  if [ $res -eq 0 ]; then
    kubectl apply -f ./manifests/keycloak.yml
  fi

  rm -f keycloak.values.yml

  # Wait for Keycloak to be up
  while [ "$(curl -s -o /dev/null -w "%{http_code}" http://keycloak.$domain:$ingress_http_port/health/ready)" != "200" ]; do
    sleep 5
  done

  local token=$(curl -s \
    -X POST \
    -H "Content-Type: application/x-www-form-urlencoded" \
    -d "client_id=admin-cli&username=admin&password=admin&grant_type=password" \
    http://keycloak.$domain:$ingress_http_port/realms/master/protocol/openid-connect/token \
    | jq -r '.access_token')
  
  # Check if go-deploy client exists, if not create it
  local check_exists=$(curl -s \
    -H "Content-Type: application/json" \
    -H \"Accept: application/json\" \
    -H "Authorization: Bearer $token" \
    -X GET http://keycloak.$domain:$ingress_http_port/admin/realms/master/clients?clientId=go-deploy)
  local exists=$(echo $check_exists | jq -r '.[] | select(.clientId=="go-deploy") | .clientId')
  if [ "$exists" != "go-deploy" ]; then
    local payload='{
      "protocol":"openid-connect",
      "clientId":"go-deploy",
      "name":"go-deploy",
      "description":"go-deploy",
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
      "redirectUris":["http://*"]
      }'
    curl -s \
      -H "Content-Type: application/json" \
      -H "Accept: application/json" \
      -H "Authorization: Bearer $token" \
      -X POST http://keycloak.$domain:$ingress_http_port/admin/realms/master/clients -d "$payload"
  fi

  # Fetch created client's secret
  keycloak_deploy_secret=$(curl -s \
    -H "Content-Type: application/json" \
    -H "Accept: application/json" \
    -H "Authorization: Bearer $token" \
    -X GET http://keycloak.$domain:$ingress_http_port/admin/realms/master/clients?clientId=go-deploy \
    | jq -r '.[0].clientSecret')



  # Check if go-deploy-storage client exists, if not create it
  local check_exists=$(curl -s \
    -H "Content-Type: application/json" \
    -H \"Accept: application/json\" \
    -H "Authorization: Bearer $token" \
    -X GET http://keycloak.$domain:$ingress_http_port/admin/realms/master/clients?clientId=go-deploy-storage)
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
      "redirectUris":["http://*"]
      }'
    curl -s \
      -H "Content-Type: application/json" \
      -H "Accept: application/json" \
      -H "Authorization: Bearer $token" \
      -X POST http://keycloak.$domain:$ingress_http_port/admin/realms/master/clients -d "$payload"
  fi

  # Fetch created client's secret
  keycloak_deploy_storage_secret=$(curl -s \
    -H "Content-Type: application/json" \
    -H "Accept: application/json" \
    -H "Authorization: Bearer $token" \
    -X GET http://keycloak.$domain:$ingress_http_port/admin/realms/master/clients?clientId=go-deploy-storage \
    | jq -r '.[0].clientSecret')
}

function install_nfs_server() {
  res=$(kubectl get ns | grep -c nfs-server)
  if [ $res -eq 0 ]; then
    kubectl apply -f ./manifests/nfs-server.yml
  fi

  # Wait for NFS server to be up
  while [ "$(kubectl get pod -n nfs-server -l app=nfs-server -o jsonpath="{.items[0].status.phase}")" != "Running" ]; do
    sleep 5
  done

  # Create subfolders deployments, vms, scratch and snapshots
  pod=$(kubectl get pod -n nfs-server -l app=nfs-server -o jsonpath="{.items[0].metadata.name}")
  kubectl exec -n nfs-server $pod -- mkdir -p  $nfs_base_path/deployments $nfs_base_path/vms $nfs_base_path/scratch $nfs_base_path/snapshots
}

function install_cert_manager() {
  # If cert-manager namespace already exists, skip
  res=$(kubectl get ns | grep -c cert-manager)
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

    # Wait for cert-manager to be up
    while [ "$(kubectl get pod -n cert-manager -l app=cert-manager -o jsonpath="{.items[0].status.phase}")" != "Running" ]; do
      sleep 5
    done
  
    kubectl apply -f ./manifests/cert.yml
  fi
}

function install_hairpin_proxy() {
  # If namespace 'hairpin-proxy' already exists, skip
  res=$(kubectl get ns | grep -c hairpin-proxy)
  if [ $res -eq 0 ]; then
    kubectl apply -f https://raw.githubusercontent.com/compumike/hairpin-proxy/v0.2.1/deploy.yml
  fi
}

function install_nfs_csi() {
  # If deployment 'csi-nfs-controller' in namespace 'kube-system' already exists, skip
  res=$(kubectl get deploy -n kube-system | grep -c csi-nfs-controller)
  if [ $res -eq 0 ]; then
    helm install csi-driver-nfs csi-driver-nfs \
      --repo https://raw.githubusercontent.com/kubernetes-csi/csi-driver-nfs/master/charts \
      --namespace kube-system \
      --version v4.6.0
  fi
}

function install_storage_classes() {
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
  res=$(kubectl get sc 2>/dev/null | grep -c nfs)
  if [ $res -eq 0 ]; then
    envsubst < ./manifests/sc-vm-disks.yml | kubectl apply -f -
  fi

  # If storage class 'deploy-vm-scratch' does not exist, create it
  export nfs_share="$nfs_base_path/scratch"
  res=$(kubectl get sc 2>/dev/null | grep -c scratch)
  if [ $res -eq 0 ]; then
    envsubst < ./manifests/sc-vm-scratch.yml | kubectl apply -f -
  fi

  # If volume snapshot class 'deploy-vm-snapshots' does not exist, create it
  export nfs_share="$nfs_base_path/snapshots"
  res=$(kubectl get volumesnapshotclass 2>/dev/null | grep -c deploy-vm-snapshots)
  if [ $res -eq 0 ]; then
    envsubst < ./manifests/vsc-vm-snapshots.yml | kubectl apply -f -
  fi
}

function install_kubevirt() {
  # If namespace 'kubevirt' already exists, skip
  res=$(kubectl get ns | grep -c kubevirt)
  if [ $res -eq 0 ]; then
    export version=$(curl -s https://storage.googleapis.com/kubevirt-prow/release/kubevirt/kubevirt/stable.txt)
    kubectl create -f https://github.com/kubevirt/kubevirt/releases/download/$version/kubevirt-operator.yaml
    kubectl create -f https://github.com/kubevirt/kubevirt/releases/download/$version/kubevirt-cr.yaml
  fi
  
  while [ "$(kubectl get kubevirt.kubevirt.io/kubevirt -n kubevirt -o=jsonpath="{.status.phase}")" != "Deployed" ]; do
    sleep 5
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

run_with_spinner "Configuring local DNS" configure_local_dns
run_with_spinner "Waiting for DNS" wait_for_dns

run_with_spinner "Install k3d" install_k3d
run_with_spinner "Set up k3d cluster" create_k3d_cluster

run_with_spinner "Install Harbor" install_harbor
run_with_spinner "Install MongoDB" install_mongodb
run_with_spinner "Install Redis" install_redis
run_with_spinner "Install Keycloak" install_keycloak
run_with_spinner "Install NFS Server" install_nfs_server
run_with_spinner "Install Cert Manager" install_cert_manager
run_with_spinner "Install Hairpin Proxy" install_hairpin_proxy
run_with_spinner "Install NFS CSI" install_nfs_csi
run_with_spinner "Install Storage Classes" install_storage_classes
run_with_spinner "Install KubeVirt" install_kubevirt
run_with_spinner "Install CDI" install_cdi

run_with_spinner "Seed Harbor with images" seed_harbor_with_images



# If exists ../../config.local.yml, ask if user want to replace it
if [ -f "../../config.local.yml" ]; then
  read -p "config.local.yml already exists. Do you want to replace it? [y/n]: " -n 1 -r
  echo
  if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Skipping config.local.yml generation"
    exit 0
  fi
fi


echo "Generating config.local.yml"
cp config.yml.tmpl ../../config.local.yml

# Core
export external_url="$domain:$ingress_http_port"
export port="8080"
export mode="dev"

# Zone
export kubeconfig_path="./kube/$cluster_name.yml"
export nfs_server="nfs-server.nfs-server.svc.cluster.local"
export nfs_parent_app_path="$nfs_base_path/deployments"
export nfs_path_vm="$nfs_base_path/vms"
export port_range_start="$port_range_start"
export port_range_end="$port_range_end"

# VM
export admin_ssh_public_key=$(cat ~/.ssh/id_rsa.pub)
export vm_image="$vm_image"

# Deployment


# Registry
export registry_url="harbor.harbor.svc.cluster.local:$harbor_port"
export placeholder_image="$registry_url/library/go-deploy-placeholder:latest"

# Keycloak
export keycloak_url="http://keycloak.deploy.localhost:$ingress_http_port"
export keycloak_realm="master"
export keycloak_admin_group="admin"
export keycloak_storage_client_id="go-deploy-storage"
export keycloak_storage_client_secret="secret"

# MongoDB
export mongodb_url="mongodb://admin:admin@localhost:$mongo_db_port"
export mongodb_name="deploy"

# Redis
export redis_url="redis://localhost:6379"
export redis_password=

# Harbor
export harbor_url="http://harbor.deploy.localhost:$ingress_http_port"
export harbor_user="admin"
export harbor_password="Harbor12345"
export harbor_webhook_secret="secret"

envsubst < config.yml.tmpl > ../../config.local.yml

echo ""
echo ""
echo -e "$GREEN_CHECK config.local.yml generated"
echo ""
echo "The following services are now available:"
echo " - Harbor: http://harbor.$domain:$ingress_http_port (admin:Harbor12345)"
echo " - Keycloak: http://keycloak.$domain:$ingress_http_port (admin:admin)"
echo " - MongoDB: mongodb://admin:admin@localhost:$mongo_db_port"
echo " - Redis: redis://redis.localhost:$redis_port"
echo ""
echo "dnsmasq is used to allow the names to resolve. See the following guides for help configuring it:"
echo " - WSL2 (Windows): https://github.com/absolunet/pleaz/blob/production/documentation/installation/wsl2/dnsmasq.md"
echo " - systemd-resolved (Linux): https://gist.github.com/frank-dspeed/6b6f1f720dd5e1c57eec8f1fdb2276df"
echo ""
echo "Please review the generated config.local.yml file and make any necessary changes"
echo ""
echo "To start the application, go the the top directory and run the following command:"
echo ""
echo -e "\033[1mDEPLOY_CONFIG_FILE=config.local.yml go run main.go\033[0m"
echo ""


