#!/bin/bash

GREEN="\033[92m"
RED="\033[91m"
RESET="\033[0m"
CHECK="\u2713"
CROSS="\u2717"

GREEN_CHECK="${GREEN}${CHECK}${RESET}"
RED_CROSS="${RED}${CROSS}${RESET}"

# Function to update /etc/systemd/resolved.conf
update_resolved_conf() {
    RESOLVED_CONF="/etc/systemd/resolved.conf"
    DNS_LINE="DNS=127.0.0.2"

    if ! grep -q "^$DNS_LINE" $RESOLVED_CONF; then
        sed -i "/^\[Resolve\]/a $DNS_LINE" $RESOLVED_CONF
    fi
}

# Function to update /etc/dnsmasq.conf
update_dnsmasq_conf() {
    DNSMASQ_CONF="/etc/dnsmasq.conf"
    LISTEN_ADDRESS="listen-address=127.0.0.2"
    BIND_INTERFACES="bind-interfaces"

    if ! grep -q "^$LISTEN_ADDRESS" $DNSMASQ_CONF; then
        echo "$LISTEN_ADDRESS" >> $DNSMASQ_CONF
    fi

    if ! grep -q "^$BIND_INTERFACES" $DNSMASQ_CONF; then
        echo "$BIND_INTERFACES" >> $DNSMASQ_CONF
    fi
}

# Function to update /etc/default/dnsmasq
update_default_dnsmasq() {
    DEFAULT_DNSMASQ="/etc/default/dnsmasq"
    IGNORE_RESOLVCONF="IGNORE_RESOLVCONF=yes"
    ENABLED="ENABLED=1"

    if ! grep -q "^$IGNORE_RESOLVCONF" $DEFAULT_DNSMASQ; then
        echo "$IGNORE_RESOLVCONF" >> $DEFAULT_DNSMASQ
    fi

    if ! grep -q "^$ENABLED" $DEFAULT_DNSMASQ; then
        echo "$ENABLED" >> $DEFAULT_DNSMASQ
    fi
}

increase_files() {
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

# Increase files
increase_files

# Install kubectl
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
chmod +x kubectl
sudo mv kubectl /usr/local/bin/kubectl

# Install Helm
curl https://baltocdn.com/helm/signing.asc | gpg --dearmor | sudo tee /usr/share/keyrings/helm.gpg > /dev/null
sudo apt-get install apt-transport-https --yes
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/helm.gpg] https://baltocdn.com/helm/stable/debian/ all main" | sudo tee /etc/apt/sources.list.d/helm-stable-debian.list > /dev/null
sudo apt-get update
sudo apt-get install helm -y

# Install Kind
if [ "$(uname -m)" = "x86_64" ]; then
    curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.23.0/kind-linux-amd64
    chmod +x ./kind
    sudo mv ./kind /usr/local/bin/kind
fi

# Install dnsmasq
sudo apt-get install dnsmasq -y

update_resolved_conf
update_dnsmasq_conf
update_default_dnsmasq

systemctl restart systemd-resolved
systemctl restart dnsmasq
