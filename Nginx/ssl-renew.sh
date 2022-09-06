#!/bin/bash
renew_ssl_cert () {
  DOMAIN="$1"
#  lego -s https://acme-staging-v02.api.letsencrypt.org/directory \
  lego -m "$ACMEMAIL" -a -d "$DOMAIN" --path /ssl --http renew
}
for keyfile in /ssl/certificates/*.key
do
    replaced="${keyfile/_/*}"
    renew_ssl_cert "${replaced%.*}" >> /tmp/cert.log
done