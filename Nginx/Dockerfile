# idea
# create a nginx container that includes lego & ngman
# it should expose ngman as an interface
# and use lego in the background to create SSL certs

FROM nginx:stable

COPY *.conf /etc/nginx/

# will include all configs in /etc/nginx/conf.d/*.conf;
RUN rm /etc/nginx/conf.d/default.conf

# install wget && clean up
RUN apt-get update && apt-get install -y wget && rm -rf /var/lib/apt/lists/*

# install lego
RUN wget -O /tmp/lego.tgz https://github.com/go-acme/lego/releases/download/v4.8.0/lego_v4.8.0_linux_amd64.tar.gz && \
    tar -xzf /tmp/lego.tgz -C /tmp && \
    mv /tmp/lego /usr/local/bin/lego && \
    rm /tmp/lego.tgz

COPY ssl-* /usr/local/bin/
