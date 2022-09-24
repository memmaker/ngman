# idea
# create a nginx container that includes lego & ngman
# it should expose ngman as an interface
# and use lego in the background to create SSL certs
FROM alpine:3.16

ARG USERNAME=nginx
ARG USER_UID=1000
ARG USER_GID=$USER_UID

ARG NGMAN_VERSION="v1.1.4"
ARG HOME=/home/$USERNAME

COPY ssl-* /usr/local/bin/

RUN addgroup -S "$USERNAME" && \
    adduser -u "$USER_UID" -G "$USERNAME" -D -g 'nginx user' -h /home/"$USERNAME"/ "$USERNAME" && \
    apk update && apk add --no-cache nginx curl && \
    mkdir -p /var/www /home/"$USERNAME"/default "$HOME" && \
    echo "<p>Default page</p>" > /home/"$USERNAME"/default/index.html && \
    curl -sL "https://github.com/memmaker/ngman/releases/download/${NGMAN_VERSION}/ngman_${NGMAN_VERSION}_linux_amd64.tgz" | tar xzO > /usr/local/bin/ngman && \
    curl -sL https://github.com/go-acme/lego/releases/download/v4.8.0/lego_v4.8.0_linux_amd64.tar.gz | tar xzO lego > /usr/local/bin/lego && \
    chown -R $USERNAME:$USERNAME "$HOME" /var/www /etc/nginx/ /usr/local/bin/ && chmod +x /usr/local/bin/ngman /usr/local/bin/lego


USER $USERNAME

COPY *.conf /etc/nginx/

RUN mkdir -p "$HOME"/.ngman
VOLUME /ssl /var/www /etc/nginx/conf.d "$HOME"/.ngman /etc/nginx/dhparam.pem

EXPOSE 1080 10443

CMD ["nginx", "-g", "daemon off;", "-c", "/etc/nginx/nginx.conf"]