#!/bin/bash

renew_ssl_cert_http () {
  DOMAIN="$1"
#  lego -s https://acme-staging-v02.api.letsencrypt.org/directory \
  lego -m "$ACMEMAIL" -a -d "$DOMAIN" --path /ssl --http renew
}

renew_ssl_cert_dns () {
  DOMAIN="$1"
#  lego -s https://acme-staging-v02.api.letsencrypt.org/directory \
  lego -m "$ACMEMAIL" -a -d "$DOMAIN" --path /ssl --dns renew
}

for keyfile in /ssl/certificates/*.key
do
    replaced="${keyfile/_/*}"
    if [[ -z "$DNS_PROVIDER" ]]; then
      renew_ssl_cert_http "${replaced%.*}" >> /tmp/cert.log
    else
      renew_ssl_cert_dns "${replaced%.*}" >> /tmp/cert.log
    fi
done