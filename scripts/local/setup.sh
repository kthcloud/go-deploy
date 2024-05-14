#!/bin/bash

# Ensure this script is run from the script folder by checking if the parent folder contains mod.go
if [ ! -f "../../go.mod" ]; then
  echo "$RED_CROSS Please run this script from the scripts folder"
  exit 1
fi

source ./common.sh


function check_dependencies() {
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
}

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
export cluster_name=go-deploy-dev
export kubeconfig_output_path=../../kube

# Domain configuration
export domain=deploy.localhost

# Placeholder git repo
export placeholder_git_repo="https://github.com/kthcloud/go-deploy-placeholder.git"
export vm_image="https://cloud-images.ubuntu.com/noble/current/noble-server-cloudimg-amd64.img"

# NFS configuration
export nfs_base_path="/nfs"
export nfs_cluster_ip="10.0.200.2"

# IAM configuration
export keycloak_deploy_secret=
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

  mkdir -p "./data"

  # Add NFS mount container /mnt/nfs to ./data
  config="$config
  extraMounts:
  - hostPath: "./data"
    containerPath: /mnt/nfs"
  
  echo "$config" > ./manifests/kind-config.yml
}

function create_kind_cluster() {
  read_cluster_config

  local current=$(kind get clusters 2> /dev/stdout | grep -c $cluster_name)
  if [ "$current" -eq 0 ]; then
    generate_kind_cluster_config
    kind create cluster --name $cluster_name --config ./manifests/kind-config.yml --quiet
  fi

  read_cluster_config

  # Wait for kubeconfig to change
  while [ "$(kubectl config current-context)" != "kind-$cluster_name" ]; do
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

  kind get kubeconfig --name $cluster_name > "$kubeconfig_output_path/$cluster_name.yml"
}

function install_nfs_server() {
  read_cluster_config

  res=$(kubectl get ns | grep -c nfs-server)
  if [ $res -eq 0 ]; then
    nfs_server_values_subst=$(mktemp)
    export nfs_cluster_ip=$nfs_cluster_ip
    envsubst < ./manifests/nfs-server.yml > $nfs_server_values_subst
    kubectl apply -f $nfs_server_values_subst
  fi

  # Wait for NFS server to be up
  while [ "$(kubectl get pod -n nfs-server -l app=nfs-server -o jsonpath="{.items[0].status.phase}" 2> /dev/stdout)" != "Running" ]; do
    sleep 5
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
      --set controller.service.nodePorts.https=$ingress_https_port
  fi

  # Wait for ingress-nginx to be up
  # while [ "$(kubectl get pod -n ingress-nginx -l app.kubernetes.io/name=ingress-nginx -o jsonpath="{.items[0].status.phase}")" != "Running" ]; do
  #   sleep 5
  # done
}

function install_harbor() {
  read_cluster_config

  # If helm release 'harbor' in namespace 'harbor' already exists, skip
  res=$(helm list -n harbor | grep -c harbor)
  if [ $res -eq 0 ]; then
    harbor_values_subst=$(mktemp)
    envsubst < ./helmvalues/harbor.values.yml > $harbor_values_subst

    helm install harbor harbor \
      --repo https://helm.goharbor.io \
      --namespace harbor \
      --create-namespace \
      --values - < $harbor_values_subst

    # Allow namespace to be created
    sleep 5
  fi

  # Setup an ingress for Harbor
  res=$(kubectl get ingress -n harbor -o yaml | grep -c harbor)
  if [ $res -eq 0 ]; then
    export domain=$domain
    export harbor_port=$harbor_port
    harbor_subst=$(mktemp)
    envsubst < ./manifests/harbor.yml > $harbor_subst
    kubectl apply -f $harbor_subst
  fi

  # Wait for Harbor to be up
  while [ "$(curl -s -o /dev/null -w "%{http_code}" http://harbor.$domain:$ingress_http_port)" != "200" ]; do
    sleep 5
  done
}

function seed_harbor_with_images() {
  read_cluster_config

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
    git clone $placeholder_git_repo --quiet
  fi

  # Use 'library' so we don't need to create our own (library is the default namespace in Harbor)
  docker build go-deploy-placeholder/ -t $domain/library/go-deploy-placeholder:latest 2> /dev/null
  docker login $domain -u $robot_user -p $robot_password 2> /dev/null
  docker push $domain/library/go-deploy-placeholder:latest 2> /dev/null

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
    sleep 5
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
  keycloak_deploy_secret=$(curl -s \
    -H "Content-Type: application/json" \
    -H "Accept: application/json" \
    -H "Authorization: Bearer $token" \
    -X GET http://keycloak.$domain:$keycloak_port/admin/realms/master/clients?clientId=go-deploy \
    | jq -r '.[0].secret')

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
      "redirectUris":["http://*"]
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

  # Write keycloak_deploy_secret and keycloak_deploy_storage_secret to cluster-config.rc
  # Overwrite if the row already exists
  sed -i "/export keycloak_deploy_secret=/c\export keycloak_deploy_secret=$keycloak_deploy_secret" ./cluster-config.rc
  sed -i "/export keycloak_deploy_storage_secret=/c\export keycloak_deploy_storage_secret=$keycloak_deploy_storage_secret" ./cluster-config.rc
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
  while [ "$(kubectl get pod -n cert-manager -l app=cert-manager -o jsonpath="{.items[0].status.phase}")" != "Running" ]; do
    sleep 5
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

check_dependencies
if [ ! -f "./cluster-config.rc" ]; then
  generate_cluster_config
fi

# Pre-requisites
run_with_spinner "Configuring local DNS" configure_local_dns
run_with_spinner "Waiting for DNS" wait_for_dns

# Base
run_with_spinner "Set up kind cluster" create_kind_cluster
run_with_spinner "Install NFS Server" install_nfs_server
run_with_spinner "Install NFS CSI" install_nfs_csi

# Apps
run_with_spinner "Install Ingress Nginx" install_ingress_nginx
run_with_spinner "Install Harbor" install_harbor
run_with_spinner "Install MongoDB" install_mongodb
run_with_spinner "Install Redis" install_redis
run_with_spinner "Install Keycloak" install_keycloak

# Dependencies
run_with_spinner "Install Cert Manager" install_cert_manager
run_with_spinner "Install Hairpin Proxy" install_hairpin_proxy
run_with_spinner "Install Storage Classes" install_storage_classes
run_with_spinner "Install KubeVirt" install_kubevirt
run_with_spinner "Install CDI" install_cdi

# Post-install
run_with_spinner "Seed Harbor with images" seed_harbor_with_images



# If exists ../../config.local.yml, ask if user want to replace it
if [ -f "../../config.local.yml" ]; then
  echo ""
  read -p "config.local.yml already exists. Do you want to replace it? [y/n]: " -n 1 -r
  echo
  if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Skipping config.local.yml generation"
    exit 0
  fi
fi


echo "Generating config.local.yml"

read_cluster_config
cp config.yml.tmpl ../../config.local.yml

# Core
export external_url="$domain:$ingress_https_port"
export port="8080"
export mode="dev"

# Zone
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
export harbor_url="http://harbor.deploy.localhost:$harbor_port"
export harbor_user="admin"
export harbor_password="Harbor12345"
export harbor_webhook_secret="secret"

envsubst < config.yml.tmpl > ../../config.local.yml

echo -e ""
echo -e ""
echo -e "$GREEN_CHECK config.local.yml generated"
echo -e ""
echo -e "dnsmasq is used to allow the names to resolve. See the following guides for help configuring it:"
echo -e " - WSL2 (Windows): https://github.com/absolunet/pleaz/blob/production/documentation/installation/wsl2/dnsmasq.md"
echo -e " - systemd-resolved (Linux): https://gist.github.com/frank-dspeed/6b6f1f720dd5e1c57eec8f1fdb2276df"
echo -e ""
echo -e "The following services are now available:"
echo -e " - ${BLUE_BOLD}Harbor${RESET}: http://harbor.$domain:$harbor_port (admin:Harbor12345)"
echo -e " - ${TEAL_BOLD}Keycloak${RESET}: http://keycloak.$domain:$keycloak_port (admin:admin)" 
echo -e "      Users: admin:admin, base:base, power:power"
echo -e "      Clients: go-deploy:$keycloak_deploy_secret, go-deploy-storage:$keycloak_deploy_storage_secret"
echo -e " - ${GREEN_BOLD}MongoDB${RESET}: mongodb://admin:admin@localhost:$mongo_db_port"
echo -e " - ${RED_BOLD}Redis${RESET}: redis://localhost:$redis_port"
echo -e " - ${ORANGE_BOLD}NFS${RESET}: nfs://localhost:$nfs_port"
echo -e ""
echo -e "To start the application, go the the top directory and run the following command:"
echo -e ""
echo -e "    ${WHITE_BOLD}DEPLOY_CONFIG_FILE=config.local.yml go run main.go${RESET}"
echo -e ""
echo -e "Happy coding! 🚀"
echo -e ""


