#!/bin/sh

EMAIL="$1"

DH_PARAM_BITS=2048

# Check for needed arguments
if [ -z "$EMAIL" ]; then
  echo "Please provide an email address as first argument"
  exit 1
fi

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

WEBROOT="$HOME/www"
CERTROOT="$HOME/ssl"

mkdir -p "$CERTROOT" "$WEBROOT"/_acme-challenges "$HOME"/.ngman "$HOME"/keys "$HOME"/nginx-conf

if ! command -v podman 1> /dev/null 2> /dev/null
then
    echo "podman not found, installing it"
    sudo apt-get update 1> /dev/null 2> /dev/null && sudo apt-get install -y podman > /dev/null
fi

if [ ! -f /etc/sysctl.d/99-rootless.conf ]; then
  echo "Setting up sysctl for rootless http services"
  echo "net.ipv4.ip_unprivileged_port_start=80" | sudo tee /etc/sysctl.d/99-rootless.conf
  sudo sysctl --system
fi

if [ ! -f "$HOME"/.ngman/dnsprovider.env ] || [ ! -s "$HOME"/.ngman/dnsprovider.env ]; then
  echo "Could not find dnsprovider.env, please create it and add your dns provider credentials"
  echo "ACME DNS Challenge is currently disabled, no wildcard certificate support."
  touch "$HOME"/.ngman/dnsprovider.env
else
  echo "Found dnsprovider.env, enabling ACME DNS Challenge and wildcard support"
fi

if [ ! -f "$HOME"/keys/dhparam.pem ]; then
  echo "Generating dhparam.pem for nginx https (this may take a while)"
  openssl dhparam -out "$HOME"/keys/dhparam.pem "$DH_PARAM_BITS"
fi


if ! podman network exists podnet; then
  echo "Creating podman network podnet"
  podman network create podnet > /dev/null
fi

if podman image exists ghcr.io/memmaker/nginx:latest; then
  echo "Pulling latest image"
  podman pull ghcr.io/memmaker/nginx:latest
fi

if podman container exists ngx; then
  echo "Removing existing container ngx"
  podman rm -f ngx > /dev/null
fi

PROXY_RESOLVER=$(podman network inspect podnet | jq -r '.[0].plugins[0].ipam.ranges[0][0].gateway')

echo "Starting container ngx"
podman run \
  -d \
  -e ACMEMAIL="$EMAIL" \
  -e NGMAN_PROXY_RESOLVER="$PROXY_RESOLVER" \
  --env-file="$HOME"/.ngman/dnsprovider.env \
  --name ngx \
  -p 80:1080 \
  -p 443:10443 \
  -v "$HOME"/.ngman:/home/nginx/.ngman \
  -v "$HOME"/keys/dhparam.pem:/etc/nginx/dhparam.pem \
  -v "$HOME"/nginx-conf:/etc/nginx/conf.d/ \
  -v "$CERTROOT":/ssl \
  -v "$WEBROOT":/var/www \
  --network podnet \
  ghcr.io/memmaker/nginx:latest

RENEWCMD="podman exec ngx ssl-renew.sh"
newcron "0 4 1 */2 * ${RENEWCMD} >/dev/null 2>&1"
