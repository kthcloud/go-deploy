#!/bin/bash

# Ensure this script is run from the script folder by checking if the parent folder contains mod.go
if [ ! -f "../../go.mod" ]; then
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

# Check if Helm is installed, if not exit
if ! [ -x "$(command -v helm)" ]; then
  echo -e "$RED_CROSS Helm is not installed. Please install Helm"
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
    k3d cluster create $name --agents 2 -p '9080:80@loadbalancer' -p '9443:443@loadbalancer' -p '29000:30000@server:0' -p '27017:27017@server:0'
  fi
}

function setup_harbor() {
  curl_result=$(curl -s localhost:11080 -o /dev/null -w "%{http_code}")
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
    sed -i 's/port: 80/port: 11080/g' ./harbor/harbor.yml
    # Set data dir
    mkdir -p ./data/harbor
    sed -i "s|data_volume: /data|data_volume: $(pwd)/data/harbor|g" ./harbor/harbor.yml

    sudo ./harbor/install.sh

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
  res=$(curl -s -u admin:Harbor12345 -X GET http://localhost:11080/api/v2.0/projects/library/repositories | jq -r '.[] | select(.name=="library/go-deploy-placeholder") | .name')
  if [ "$res" == "library/go-deploy-placeholder" ]; then
    return
  fi

  # Download repo and build the image
  if [ ! -d "go-deploy-placeholder" ]; then
    git clone $PLACEHOLDER_GIT_REPO
  fi

  # Use 'library' so we don't need to create our own (library is the default namespace in Harbor)
  docker build go-deploy-placeholder/ -t localhost:8000/library/go-deploy-placeholder:latest
  docker login localhost:8000 -u admin -p Harbor12345
  docker push localhost:8000/library/go-deploy-placeholder:latest

  # Remove the placeholder repo
  rm -rf go-deploy-placeholder
}


function install_mongodb() {
  # If namespace 'mongodb' already exists, skip
  res=$(kubectl get ns | grep -c mongodb)
  if [ $res -eq 0 ]; then
    kubectl apply -f - <<EOF
apiVersion: v1
kind: Namespace
metadata:
  name: mongodb
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mongodb
  namespace: mongodb
spec:
  selector:
    matchLabels:
      app: mongodb
  template:
    metadata:
      labels:
        app: mongodb
    spec:
      containers:
      - name: mongodb
        image: mongo:4.4
        ports:
        - containerPort: 27017
        env:
        - name: MONGO_INITDB_ROOT_USERNAME
          value: root
        - name: MONGO_INITDB_ROOT_PASSWORD
          value: root
        volumeMounts:
        - name: mongodb-data
          mountPath: /data/db
      volumes:
      - name: mongodb-data
        hostPath:
          path: /mnt/nfs/mongodb
          type: DirectoryOrCreate
---
apiVersion: v1
kind: Service
metadata:
  name: mongodb
  namespace: mongodb
spec:
  ports:
    - port: 27017
      targetPort: 27017
  selector:
    app: mongodb
  type: LoadBalancer
EOF
  fi
}

function install_redis() {
  # If namespace 'redis' already exists, skip
  res=$(kubectl get ns | grep -c redis)
  if [ $res -eq 0 ]; then
    kubectl apply -f - <<EOF
apiVersion: v1
kind: Namespace
metadata:
  name: redis
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: redis
  namespace: redis
spec:
  selector:
    matchLabels:
      app: redis
  template:
    metadata:
      labels:
        app: redis
    spec:
      containers:
      - name: redis
        image: redis:6.2
        ports:
        - containerPort: 6379
---
apiVersion: v1
kind: Service
metadata:
  name: redis
  namespace: redis
spec:
  ports:
    - port: 6379
      targetPort: 6379
  selector:
    app: redis
  type: LoadBalancer
EOF
  fi
}

function install_keycloak() {
  # If namespace 'keycloak' already exists, skip
  res=$(kubectl get ns | grep -c keycloak)
  if [ $res -eq 0 ]; then
    values_file=$(mktemp)

    cat > $values_file <<EOF
service:
  httpPort: 12080
  httpsPort: 12443
  type: LoadBalancer
extraEnv: |
  - name: KEYCLOAK_USER
    value: admin
  - name: KEYCLOAK_PASSWORD
    value: admin
EOF
    helm install keycloak keycloak \
      --repo https://codecentric.github.io/helm-charts \
      --namespace keycloak \
      --create-namespace \
      --values $values_file
  fi
}

function install_nfs_server() {
  res=$(kubectl get ns | grep -c nfs-server)
  if [ $res -eq 0 ]; then
    kubectl apply -f - <<EOF
apiVersion: v1
kind: Namespace
metadata:
  name: nfs-server
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nfs-server
  namespace: nfs-server
spec:
  selector:
    matchLabels:
      app: nfs-server
  template:
    metadata:
      labels:
        app: nfs-server
    spec:
      containers:
      - name: nfs-server
        image: k8s.gcr.io/volume-nfs:0.8
        ports:
        - name: nfs
          containerPort: 2049
        - name: mountd
          containerPort: 20048
        - name: rpcbind
          containerPort: 111
        securityContext:
          privileged: true
        volumeMounts:
        - name: storage
          mountPath: /mnt/nfs
      volumes:
      - name: storage
        hostPath:
          path: /mnt/nfs
          type: DirectoryOrCreate
---
apiVersion: v1
kind: Service
metadata:
  name: nfs-server
  namespace: nfs-server
spec:
  ports: 
    - name: nfs
      port: 2049
      targetPort: 2049
    - name: mountd
      port: 20048
      targetPort: 20048
    - name: rpcbind
      port: 111
      targetPort: 111
  selector:
    app: nfs-server
  type: ClusterIP
EOF
  fi

  sleep 10

  # Create subfolders deployments, vms, scratch and snapshots
  pod=$(kubectl get pod -n nfs-server -l app=nfs-server -o jsonpath="{.items[0].metadata.name}")
  kubectl exec -n nfs-server $pod -- mkdir -p /mnt/nfs/deployments /mnt/nfs/vms /mnt/nfs/scratch /mnt/nfs/snapshots
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
  
    kubectl apply -f - <<EOF
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: go-deploy-cluster-issuer
spec:
  selfSigned: {}
---
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: go-deploy-wildcard
  namespace: kube-system
spec:
  secretName: go-deploy-wildcard-secret
  secretTemplate:
    labels:
      app.kubernetes.io/deploy-name: deploy-wildcard-secret
  issuerRef: 
    kind: ClusterIssuer
    name: go-deploy-cluster-issuer
  commonName: ""
  dnsNames:
    - "*.apps.local"
    - "*.vm-app.local"
    - "*.storage.local"
EOF
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

  # If storage class 'deploy-vm-disks' does not exist, create it
  res=$(kubectl get sc | grep -c nfs)
  if [ $res -eq 0 ]; then
    kubectl apply -f - <<EOF
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: deploy-vm-disks
parameters:
  server: nfs-server.nfs-server.svc.cluster.local
  share: /mnt/nfs/vms
provisioner: nfs.csi.k8s.io
reclaimPolicy: Delete
EOF
  fi

  # If storage class 'deploy-vm-scratch' does not exist, create it
  res=$(kubectl get sc | grep -c scratch)
  if [ $res -eq 0 ]; then
    kubectl apply -f - <<EOF
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: deploy-vm-scratch
parameters:
  server: nfs-server.nfs-server.svc.cluster.local
  share: /mnt/nfs/scratch
provisioner: nfs.csi.k8s.io
reclaimPolicy: Delete
EOF
  fi

  # If volume snapshot class 'deploy-vm-snapshots' does not exist, create it
  res=$(kubectl get volumesnapshotclass | grep -c deploy-vm-snapshots)
  if [ $res -eq 0 ]; then
    kubectl apply -f - <<EOF
apiVersion: snapshot.storage.k8s.io/v1
driver: nfs.csi.k8s.io
kind: VolumeSnapshotClass
metadata:
  name: deploy-vm-snapshots
parameters:
  server: nfs-server.nfs-server.svc.cluster.local
  share: /mnt/nfs/snapshots
deletionPolicy: Delete
EOF
  fi
}

function install_kubevirt() {
  # If namespace 'kubevirt' already exists, skip
  res=$(kubectl get ns | grep -c kubevirt)
  if [ $res -eq 0 ]; then
    export VERSION=$(curl -s https://storage.googleapis.com/kubevirt-prow/release/kubevirt/kubevirt/stable.txt)
    kubectl create -f https://github.com/kubevirt/kubevirt/releases/download/$VERSION/kubevirt-operator.yaml
    kubectl create -f https://github.com/kubevirt/kubevirt/releases/download/$VERSION/kubevirt-cr.yaml
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
    export TAG=$(curl -s -w %{redirect_url} https://github.com/kubevirt/containerized-data-importer/releases/latest)
    export VERSION=$(echo ${TAG##*/})
    kubectl apply -f https://github.com/kubevirt/containerized-data-importer/releases/download/$VERSION/cdi-operator.yaml
    kubectl apply -f https://github.com/kubevirt/containerized-data-importer/releases/download/$VERSION/cdi-cr.yaml
  fi

  # Ensure spec.config.scratchSpaceStorageClass: deploy-vm-scratch
  if [ "$(kubectl get cdi.kubevirt.io/cdi -n cdi -o=jsonpath="{.spec.config.scratchSpaceStorageClass}")" != "deploy-vm-scratch" ]; then
    kubectl patch cdi.kubevirt.io/cdi -n cdi --type='json' -p='[{"op": "add", "path": "/spec/config/scratchSpaceStorageClass", "value": "deploy-vm-scratch"}]'
  fi
}

function generate_config() {
  # if [ ! -f "../../config.local.yml" ]; then
    cp config.yml.tmpl ../../config.local.yml

    export port=8080
    envsubst < config.yml.tmpl > ../../config.local.yml

    # sed -i 's/port:/port: 8000/' ../config.local.yml
    # sed -i 's/placeholderImage:/placeholderImage: localhost:8000\/library\/go-deploy-placeholder/g' ../config.local.yml
  
  # fi
}

run_with_spinner "Install k3d" install_k3d
run_with_spinner "Set up k3d cluster" create_k3d_cluster
run_with_spinner "Set up Harbor" setup_harbor
run_with_spinner "Seed Harbor with placeholder images" seed_harbor_with_placeholder_images

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

# TODO: storage classes

run_with_spinner "Generate config.local.yml" generate_config