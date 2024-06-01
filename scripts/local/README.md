# 🛠️ Local Environment

This directory contains scripts to set up and tear down a local environment for go-deploy.

## `setup.sh`

The `setup.sh` script will set up an entire local environment with every service installed in a Kind Kubernetes cluster on your local machine.

It is only supported on Linux distributions with `apt` or `dnf` package managers.

Run `./setup.sh --help` to see the available options.

### Local Packages
The following packages will be installed on your machine:
| Package | Description |
| --- | --- |
| [Dnsmasq](http://www.thekelleys.org.uk/dnsmasq/doc.html) | DNS forwarder |

### Local CLI Tools
The following CLI tools will be installed:
| Tool | Description | Version |
| --- | --- | --- |
| [kind](https://kind.sigs.k8s.io/) | Kubernetes in Docker | v0.23.0  |
| [kubectl](https://kubernetes.io/docs/reference/kubectl/overview/) | Kubernetes CLI | latest |
| [helm](https://helm.sh/) | Kubernetes package manager | latest |
| [jq](https://stedolan.github.io/jq/) | JSON processor | latest |

### Kubernetes Services
The following services will be installed in the Kubernetes cluster:
| Service | Description | Version |
| --- | --- | --- |
| [NFS Server](https://github.com/kubernetes/examples/blob/master/staging/volumes/nfs/nfs-server-deployment.yaml) | Network File System server | 0.8 |
| [NFS CSI Driver](https://github.com/kubernetes-csi/csi-driver-nfs) | Network File System Container Storage Interface driver | v4.6.0 |
| [Ingress NGINX](https://github.com/kubernetes/ingress-nginx) | Ingress controller | 1.0.0 | latest |
| [Harbor](https://goharbor.io/) | Container image registry | v1.14.2 |
| [MongoDB](https://www.mongodb.com/) | NoSQL database | 6.0 |
| [Redis](https://redis.io/) | In-memory data structure store | 6.2 |
| [Keycloak](https://www.keycloak.org/) | Identity and Access Management | 24.0.1 |
| [cert-manager](https://cert-manager.io/) | Certificate management | v1.14.4 |
| [KubeVirt](https://kubevirt.io/) | VM-extension for Kubernetes | latest |
| [CDI](https://kubevirt.io/user-guide/storage/containerized_data_importer/) | Containerized Data Importer | latest |
| [kubemacpool](https://github.com/k8snetworkplumbingwg/kubemacpool) | Kubernetes MAC address pool | latest |


### Storage
The setup script creates a local directory `$HOME/go-deploy-storage/go-deploy-dev` to store the data for the NFS server.

This is done using a mix of manually provisioning the storage and using the NFS CSI driver to create a PersistentVolumeClaim.

The following storage classes are created:
| Storage Class | Description |
| --- | --- |
| `deploy-misc` | Storage class for miscellaneous data |
| `deploy-vm-disk` | Storage class for VM disks |
| `deploy-vm-scratch` | Storage class scratch space for VMs |

### Port Configuration
The setup scripts generates a `cluster-config.rc` with random prots in the NodePort range `30000-32767`. The ports are used to expose the services in the Kubernetes cluster and VM ports in go-deploy.

## `teardown.sh`

The `teardown.sh` script will remove the entire local environment created by the `setup.sh` script. It will remove the Kind cluster and all the services installed in the cluster.

Run `./teardown.sh --help` to see the available options.

