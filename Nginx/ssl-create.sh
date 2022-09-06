#!/bin/bash
create_ssl_cert () {
  DOMAIN="$1"
  WEBROOT="/var/www"
#  lego -s https://acme-staging-v02.api.letsencrypt.org/directory \
   lego -m "$ACMEMAIL" -a -d "$DOMAIN" --path /ssl --http --http.webroot "$WEBROOT" run
}
create_ssl_cert "$1"