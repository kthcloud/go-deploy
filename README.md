# go-deploy

[![ci](https://github.com/kthcloud/go-deploy/actions/workflows/docker-image.yml/badge.svg)](https://github.com/kthcloud/go-deploy/actions/workflows/docker-image.yml)

A simple API to create deployments similar to Heroku and spin up virtual machines built on top of kthcloud

To start the project, the following envs must be defined

```
DEPLOY_SESSION_SECRET
User generated secret

DEPLOY_EXTERNAL_URL
External URL of the API

DEPLOY_PARENT_DOMAIN
Base URL for deployments

DEPLOY_PARENT_DOMAIN_VM
Base URL for virtual machines

DEPLOY_DOCKER_REGISTRY_URL
URL to push Docker images to

DEPLOY_PLACEHOLDER_DOCKER_IMAGE
Image for newly created deployments

DEPLOY_APP_PORT
Internal port for deployments

DEPLOY_APP_PREFIX
Internal prefix for deployment names

DEPLOY_KEYCLOAK_URL
Keycloak URL (only domain name)

DEPLOY_KEYCLOAK_REALM
Keycloak Realm to use for authentication

DEPLOY_DB_URL
URL for MongoDB (only domain name)

DEPLOY_DB_NAME
Database name where a collection 'deploy' will be put

DEPLOY_DB_USERNAME
Username for database authentication

DEPLOY_DB_PASSWORD
Password for database authentication

DEPLOY_K8S_CONFIG
Base64-encoded kube config to a cluster where deployments will be hosted

DEPLOY_NPM_API_URL
URL to NPM (usually under /api)

DEPLOY_NPM_ADMIN_IDENTITY
Username for NPM authentication

DEPLOY_NPM_ADMIN_SECRET
Password for NPM authentication

DEPLOY_HARBOR_API_URL
URL to Harbor API (usually under /api/v2.0)

DEPLOY_HARBOR_WEBHOOK_SECRET
Secret used to authenticate webhook requests

DEPLOY_HARBOR_ADMIN_IDENTITY
Username for Harbor authentication

DEPLOY_HARBOR_ADMIN_SECRET
Password for Harbor authentication

DEPLOY_CS_API_URL
URL to CloudStack API (usually under /client/api)

DEPLOY_CS_API_KEY
API key for CloudStack authentication

DEPLOY_CS_SECRET_KEY
Secret key for CloudStack authentication

DEPLOY_CS_VM_PASSWORD
Password for newly created virtual machines (temporary solution)

DEPLOY_PFSENSE_API_URL
URL to pfSense API (usually under /api/v1)

DEPLOY_PFSENSE_ADMIN_IDENTITY
Username for pfSense authentication

DEPLOY_PFSENSE_ADMIN_SECRET
Password for pfSense authentication

DEPLOY_PFSENSE_PUBLIC_IP
Public IP to access virtual machines through

DEPLOY_PFSENSE_PORT_RANGE
Port range to sample virtual machine ports from
```

