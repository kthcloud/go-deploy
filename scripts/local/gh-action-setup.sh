#!/bin/bash

GREEN="\033[92m"
RED="\033[91m"
RESET="\033[0m"
BLUE="\033[94m"

CHECK="\u2713"
CROSS="\u2717"
RIGHT_ARROW="\u27A1"

GREEN_CHECK="${GREEN}${CHECK}${RESET}"
RED_CROSS="${RED}${CROSS}${RESET}"
BLUE_RIGHT_ARROW="${BLUE}${RIGHT_ARROW}${RESET}"

# Function to update /etc/systemd/resolved.conf
update_resolved_conf() {
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

    sudo cat $RESOLVED_CONF

    echo -e "${GREEN_CHECK} Updated /etc/systemd/resolved.conf"
}

# Function to update /etc/dnsmasq.conf
update_dnsmasq_conf() {
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
update_default_dnsmasq() {
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

increase_files() {
  echo -e "${BLUE_RIGHT_ARROW} Increasing inotify file watch limits..."

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

  echo -e "${GREEN_CHECK} Increased inotify file watch limits"
}

function install_kubectl() {
  echo -e "${BLUE_RIGHT_ARROW} Installing kubectl..."
  curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
  chmod +x kubectl
  sudo mv kubectl /usr/local/bin/kubectl
  echo -e "${GREEN_CHECK} kubectl installed"
}

function install_helm() {
  echo -e "${BLUE_RIGHT_ARROW} Installing Helm..."
  curl https://baltocdn.com/helm/signing.asc | gpg --dearmor | sudo tee /usr/share/keyrings/helm.gpg > /dev/null
  sudo apt-get install apt-transport-https --yes
  echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/helm.gpg] https://baltocdn.com/helm/stable/debian/ all main" | sudo tee /etc/apt/sources.list.d/helm-stable-debian.list > /dev/null
  sudo apt-get update
  sudo apt-get install helm -y
  echo -e "${GREEN_CHECK} Helm installed"
}

function install_kind() {
  echo -e "${BLUE_RIGHT_ARROW} Installing kind..."
  curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.11.1/kind-linux-amd64
  chmod +x ./kind
  sudo mv ./kind /usr/local/bin/kind
  echo -e "${GREEN_CHECK} kind installed"
}

function install_dnsmaq() {
  echo -e "${BLUE_RIGHT_ARROW} Installing dnsmasq..."

  # Make systemd-resolved no longer listen on 127.0.0.1:53
  update_resolved_conf
  sudo systemctl restart systemd-resolved

  # Install dnsmasq
  sudo apt-get install dnsmasq -y
  
  # Make dnsmasq fallback to 127.0.0.2:53
  update_dnsmasq_conf
  update_default_dnsmasq

  sudo systemctl restart dnsmasq

  echo -e "${GREEN_CHECK} dnsmasq installed"
}

increase_files
install_kubectl
install_helm
install_kind
install_dnsmaq