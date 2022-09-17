#!/bin/bash

WEBROOT="/var/www"

# User has to set these values for the container env

#DNS_PROVIDER="cloudflare"
#CLOUDFLARE_EMAIL="you@example.com"
#CLOUDFLARE_API_KEY="yourprivatecloudflareapikey"

create_ssl_cert_http () {
  DOMAIN="$1"
#  lego -s https://acme-staging-v02.api.letsencrypt.org/directory \
   lego -m "$ACMEMAIL" -a -d "$DOMAIN" --path /ssl --http --http.webroot "$WEBROOT" run
}

create_ssl_cert_dns () {
  DOMAIN="$1"
#  lego -s https://acme-staging-v02.api.letsencrypt.org/directory \
   lego -m "$ACMEMAIL" -a -d "$DOMAIN" --path /ssl --dns "$DNS_PROVIDER" run
}

if [[ -z "$DNS_PROVIDER" ]]; then
  create_ssl_cert_http "$1"
else
  create_ssl_cert_dns "$1"
fi