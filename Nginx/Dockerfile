# idea
# create a nginx container that includes lego & ngman
# it should expose ngman as an interface
# and use lego in the background to create SSL certs
FROM nginx:stable

ARG USERNAME=nginx
ARG USER_UID=1000
ARG USER_GID=$USER_UID

# Create the user
#RUN groupadd --gid $USER_GID $USERNAME \
#    && useradd --uid $USER_UID --gid $USER_GID -m $USERNAME

ARG NGMAN_VERSION="v1.1.2"
ARG HOME=/home/$USERNAME

COPY *.conf /etc/nginx/
COPY ssl-* /usr/local/bin/

# will include all configs in /etc/nginx/conf.d/*.conf;
RUN rm /etc/nginx/conf.d/default.conf

# install wget && clean up
RUN mkdir -p /var/www "$HOME"&& chown -R $USERNAME:$USERNAME "$HOME" /var/www /etc/nginx/

USER $USERNAME

RUN mkdir -p "$HOME"/bin "$HOME"/.ngman
RUN    curl -sL "https://github.com/memmaker/ngman/releases/download/${NGMAN_VERSION}/ngman_${NGMAN_VERSION}_linux_amd64.tgz" | tar xzO > "$HOME"/bin/ngman
RUN    curl -sL "https://github.com/memmaker/ngman/releases/download/${NGMAN_VERSION}/nginx.txt" > "$HOME"/.ngman/nginx.txt
RUN    printf "CertificateRootPath = '/ssl/certificates'\nSiteStorageDirectory = '%s/.ngman/sites'\nNginxSiteConfigDirectory = '/etc/nginx/conf.d'\nTemplateFile = '%s/.ngman/nginx.txt'\nPostRunCommand = 'service nginx reload'\nWebRootPath = '/var/www'\nGenerateCertCommand = 'ssl-create.sh'" "$HOME" "$HOME" > "$HOME"/.ngman/config.toml

# install lego
RUN curl -sL https://github.com/go-acme/lego/releases/download/v4.8.0/lego_v4.8.0_linux_amd64.tar.gz | tar xzO lego > "$HOME"/bin/lego

VOLUME /ssl /var/www /etc/nginx/conf.d "$HOME"/.ngman /etc/nginx/dhparam.pem

EXPOSE 80 443

