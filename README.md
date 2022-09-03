# ngman

A simple CLI tool for managing nginx sites.

**Homepage:** https://textzentrisch.de/pages/ngman/

## Concept

The configuration and state of this tool is kept under its config directory.
By default, that is **"~/.ngman/"** and the tool will create it on first start.

The config directory also contains **nginx.txt**, the file with all the partial templates.
You can easily adapt that file to your needs.

For every site that it manages the tool will create a file in the **~/.ngman/sites/** directory.

You can always re-create all the nginx config file by running **ngman write-all**.

## Config.json Options

    {
        "CertificateRootPath": "/ssl/certificates",
        "SiteStorageDirectory": "/root/.ngman/sites",
        "NginxSiteConfigDirectory": "/etc/nginx/sites-enabled",
        "TemplateFile": "/root/.ngman/nginx.txt",
        "PostRunCommand": "service nginx reload",
        "GenerateCertCommand": "create_ssl_cert"
    }

**NOTE:**
I suggest using [lego](https://github.com/go-acme/lego) in combination with [podman](https://podman.io/) for certificate generation.
You can then do something like this

    create_ssl_cert () {
        podman run \
        --env [YOUR-DOMAIN-API-TOKEN] \
        -v /ssl:/lego \
        goacme/lego \
        --accept-tos \
        --path /lego \
        --email [YOUR-EMAIL] \
        --dns dode \
        --domains "$@" \
        run
    }

Which will create a command as expected by **ngman**, where you just have to provide a domain name as argument.

## Installation

    ARCH=darwin_amd64;
    mkdir ~/.ngman > /dev/null 2>&1;
    pushd;
    cd ~/.ngman && \
    wget https://github.com/memmaker/ngman/releases/latest/download/nginx.txt && \
    wget https://github.com/memmaker/ngman/releases/latest/download/ngman_${ARCH}.zip && \ 
    unzip ngman_${ARCH}.zip && rm ngman_${ARCH}.zip && \
    mv ngman_${ARCH} /usr/local/bin/ngman && popd

## Uninstall

    rm -rf ~/.ngman && rm /usr/local/bin/ngman

## Usage

    ngman list                                                               
    ngman edit <domain>                                                      
    ngman add-static <domain> <root-path> <uri-location>                     
    ngman add-proxy <domain> <endpoint> <uri-location>                       
    ngman delete <domain>                                                    
    ngman write-all      