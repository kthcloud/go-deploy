#!/bin/bash

source "./common.sh"

# Ensure this script is run from the script folder by checking if the parent folder contains mod.go
if [ ! -f "../../go.mod" ]; then
  echo "$RED_CROSS Please run this script from the scripts folder"
  exit 1
fi

function print_usage() {
  echo -e "Usage: $0 [options]"
  echo -e "Options:"
  echo -e "  -h, --help\t\t\tPrint this help message"
  echo -e "  --non-interactive\t\tSkip all user input and fancy output. Default: false"
}

function parse_flags() {
  local args=("$@")
  local index=0

  NON_INTERACTIVE=false

  while [[ $index -lt ${#args[@]} ]]; do
    case "${args[$index]}" in
      -h|--help)
        print_usage
        exit 0
        ;;
      --non-interactive)
        NON_INTERACTIVE=true
        ((index++))
        ;;
      *)
        echo "Error: Unrecognized argument: ${args[$index]}"
        print_usage
        exit 1
        ;;
    esac
  done
}


function delete_kind_cluster() {
  name="go-deploy-dev"
  current=$(kind get clusters 2> /dev/stdout | grep -c $name)
  if [ $current -eq 1 ]; then
    kind delete cluster --name $name > /dev/null 2>&1
  fi

  rm -f ./manifests/kind-config.yml
}

function delete_local_dns_record() {
  sudo rm -f /etc/dnsmasq.d/50-go-deploy-dev.conf
}

parse_flags "$@"

run_task "Delete Kind cluster" delete_kind_cluster
run_task "Delete local DNS record" delete_local_dns_record