#!/bin/bash

EMAIL="$1"
NGMAN_VERSION=v1.0.2

if [ -z "$EMAIL" ]; then
  echo "Please provide an email address as first argument"
  exit 1
fi

if ! command -v podman &> /dev/null
then
    echo "podman not found, installing it"
    apt-get update &> /dev/null && apt-get install -y unzip podman > /dev/null
fi

if [ ! -f /usr/local/bin/ngman ]; then
  echo "ngman not found, installing it"
  mkdir -p "$HOME"/.ngman/nginx-conf /ssl /var/www/_acme-challenges && \
  wget -qO /tmp/ngman.zip https://github.com/memmaker/ngman/releases/download/${NGMAN_VERSION}/ngman_linux_amd64.zip && \
  unzip /tmp/ngman.zip -d /tmp > /dev/null && \
  mv /tmp/ngman_linux_amd64 /usr/local/bin/ngman && \
  rm /tmp/ngman.zip && \
  wget -qO "$HOME"/.ngman/nginx.txt https://github.com/memmaker/ngman/releases/download/${NGMAN_VERSION}/nginx.txt
  printf "CertificateRootPath = '/ssl/certificates'\nSiteStorageDirectory = '%s/.ngman/sites'\nNginxSiteConfigDirectory = '%s/.ngman/nginx-conf'\nTemplateFile = '%s/.ngman/nginx.txt'\nPostRunCommand = 'podman exec ngx service nginx reload'\nWebRootPath = '/var/www'\nGenerateCertCommand = 'podman exec ngx ssl-create.sh'" "$HOME" "$HOME" "$HOME" > "$HOME"/.ngman/config.toml
fi


if [ ! -f "$HOME"/.ngman/dhparam.pem ]; then
  echo "Generating dhparam.pem for nginx https (this may take a while)"
  openssl dhparam -out "$HOME"/.ngman/dhparam.pem 4096
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
  -v /ssl:/ssl \
  -v /var/www:/var/www \
  --network podnet \
  ghcr.io/memmaker/nginx

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

RENEWCMD="podman exec ngx ssl-renew.sh"
newcron "0 4 1 */2 * ${RENEWCMD} >/dev/null 2>&1"
