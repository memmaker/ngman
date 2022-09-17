#!/bin/bash

# Check for needed arguments
if [ -z "$EMAIL" ]; then
  echo "Please provide an email address as first argument"
  exit 1
fi

# Function definitions
get_latest_release() {
  curl --silent "https://api.github.com/repos/$1/releases/latest" | # Get latest release from GitHub api
    grep '"tag_name":' |                                            # Get tag line
    sed -E 's/.*"([^"]+)".*/\1/'                                    # Pluck JSON value
}

newcron () {
  crontab -l > /tmp/crontab_temp 2> /dev/null
  if grep -Fxq "$*" /tmp/crontab_temp; then
    echo "Cronjob already exists, skipping"
  else
    echo "Adding cronjob $*"
    echo "$*" >> /tmp/crontab_temp && \
    crontab /tmp/crontab_temp && \
    rm /tmp/crontab_temp
  fi
}


EMAIL="$1"
NGMAN_VERSION=$(get_latest_release "memmaker/ngman")

WEBROOT="$HOME/www"
CERTROOT="$HOME/ssl"

mkdir -p "$CERTROOT" "$WEBROOT"/_acme-challenges

if ! command -v podman &> /dev/null
then
    echo "podman not found, installing it"
    sudo apt-get update &> /dev/null && sudo apt-get install -y podman > /dev/null
fi

if [ ! -f /etc/sysctl.d/99-rootless.conf ]; then
  echo "Setting up sysctl for rootless http services"
  echo "net.ipv4.ip_unprivileged_port_start=80" | sudo tee /etc/sysctl.d/99-rootless.conf
  sudo sysctl --system
fi


if [ ! -f "$HOME"/bin/ngman ]; then
  echo "ngman not found, installing it"
  mkdir -p "$HOME"/.ngman/nginx-conf && \
  curl -sL https://github.com/memmaker/ngman/releases/download/${NGMAN_VERSION}/ngman_linux_amd64.tgz | tar xzO > "$HOME"/bin/ngman && \
  curl -sL https://github.com/memmaker/ngman/releases/download/${NGMAN_VERSION}/nginx.txt > "$HOME"/.ngman/nginx.txt && \
  printf "CertificateRootPath = '%s/ssl/certificates'\nSiteStorageDirectory = '%s/.ngman/sites'\nNginxSiteConfigDirectory = '%s/.ngman/nginx-conf'\nTemplateFile = '%s/.ngman/nginx.txt'\nPostRunCommand = 'podman exec ngx service nginx reload'\nWebRootPath = '%s/www'\nGenerateCertCommand = 'podman exec ngx ssl-create.sh'" "$HOME" "$HOME" "$HOME" "$HOME" "$HOME" > "$HOME"/.ngman/config.toml
fi


if ! podman network exists podnet; then
  echo "Creating podman network podnet"
  podman network create podnet > /dev/null
fi
if podman container exists ngx; then
  echo "Removing existing container ngx"
  podman rm -f ngx > /dev/null
fi

if [ ! -f "$HOME"/.ngman/dnsprovider.env ] || [ ! -s "$HOME"/.ngman/dnsprovider.env ]; then
  echo "Could not find dnsprovider.env, please create it and add your dns provider credentials"
  echo "ACME DNS Challenge is currently disabled, no wildcard certificate support."
  touch "$HOME"/.ngman/dnsprovider.env
else
  echo "Found dnsprovider.env, enabling ACME DNS Challenge and wildcard support"
fi

echo "Starting container ngx"
podman run \
  -d \
  -e ACMEMAIL="$EMAIL" \
  --env-file="$HOME"/.ngman/dnsprovider.env \
  --name ngx \
  -p 80:80 \
  -p 443:443 \
  -v "$HOME"/.ngman/dhparam.pem:/etc/nginx/dhparam.pem \
  -v "$HOME"/.ngman/nginx-conf:/etc/nginx/conf.d/ \
  -v "$CERTROOT":/ssl \
  -v "$WEBROOT":/var/www \
  --network podnet \
  ghcr.io/memmaker/nginx:latest


RENEWCMD="podman exec ngx ssl-renew.sh"
newcron "0 4 1 */2 * ${RENEWCMD} >/dev/null 2>&1"

