# ngman

A lightweight abstraction layer around [nginx](https://www.nginx.com/) and [lego](https://github.com/go-acme/lego)

**Homepage / Demo:** https://textzentrisch.de/pages/ngman/

## Features

 * Launch a new website with a single command
 * Supports static locations
 * Supports reverse proxy locations
 * Simplified declarative or imperative configuration
 * Automatic SSL certificate generation and renewal

It basically aims at making [nginx](https://www.nginx.com/) as easy to configure as [Caddy](https://caddyserver.com/).
At least regarding the specific use-cases of static site hosting and reverse proxying.

## Requirements

1. A linux (Ubuntu 22.04) based web-server with root shell access
2. A domain name pointing to the ip address of the web-server

**NOTE:** Currently the setup.sh script uses **apt** to install **podman**.
It *should* also work correctly if you just pre-install podman via your package manager of choice and then run
**setup.sh** script.

In combination with [podman](https://podman.io/) and a pre-configured nginx container, you can do some pretty cool stuff.
These examples use a container that has been built from the [ngman/Nginx subdirectory](https://github.com/memmaker/ngman/tree/main/Nginx).

## Self-hosted HTTPS static content in three steps

    1. Setup a Web Server
    curl -sL https://raw.githubusercontent.com/memmaker/ngman/main/setup.sh | bash -s <your-acme-mail>

    2. Add a site with the respective domain
    ngman add-site <your-domain>

    3. Publish your content
    echo "It Works" > /var/www/<your-domain>/index.html

You can now visit https://&lt;your-domain&gt;/ in the browser and will see "It Works".

## Self-hosted HTTPS reverse proxy in three steps

    1. Setup a Web Server
    curl -sL https://raw.githubusercontent.com/memmaker/ngman/main/setup.sh | bash -s <your-acme-mail>

    2. Startup your service container
    podman run --name webserver --network podnet -dt docker.io/library/httpd:alpine

    3. Add your service to ngman
    ngman add-proxy <your-domain> http://webserver:80

You can now visit https://&lt;your-domain&gt;/ in the browser and will see "It Works".

## Adding new sites locations

You can add additional virtual hosts to your web server by using the respective command:

    ngman add-site <your-domain>
    or
    ngman add-location <your-domain> /static /var/www/<your-domain>/static
    or
    ngman add-proxy <your-domain> http://webserver:80

These will

 * update your site's configuration,
 * write a new nginx configuration file, 
 * make sure that the SSL certificate is available 
 * and reload the nginx container.

Alternatively you can also edit the configuration file directly:

    ngman edit <your-domain>

## What does setup.sh do?

1. Installs [podman](https://podman.io/)
2. Installs [ngman](https://github.com/memmaker/ngman)
3. Generate DH parameters for HTTPS
4. Setup a container network with DNS support
5. Start an pre-configured nginx container that includes [lego](https://github.com/go-acme/lego)
6. Setup a cronjob for automatic SSL certificate renewal

## Concepts of ngman

A site is uniquely identified by the domain name.

This tool supports two types of locations: static & proxy.

A site can have multiple static and proxy locations.

The configuration and state of this tool is kept under its config directory.
By default, that is **"~/.ngman/"** and the tool will create it on first start.
I am using [TOML](https://github.com/toml-lang/toml) as the configuration format.

The config directory also needs to contain **"nginx.txt"**, the file with all the partial templates.
You can easily adapt that file to your needs to adjust the nginx configurations created.

For every site that it manages the tool will create a file in the **~/.ngman/sites/** directory.

You can always re-create all the nginx config files by running **ngman write-all**.

## Global settings (config.toml) 

Example config.toml file in a production environment:

    CertificateRootPath = '/ssl/certificates'
    SiteStorageDirectory = '/root/.ngman/sites'
    NginxSiteConfigDirectory = '/etc/nginx/sites-enabled'
    TemplateFile = '/root/.ngman/nginx.txt'
    PostRunCommand = 'service nginx reload'
    GenerateCertCommand = 'create_ssl_cert'

### CertificateRootPath

The path to the directory where the SSL certificates are stored.
The files are expected to conform to the following naming scheme:

    <domain>.key
    <domain>.crt

### SiteStorageDirectory

The path to the directory where the TOML site configuration files are stored. 
Must be writable by the user executing the tool.

### NginxSiteConfigDirectory

This is the main output directory for the nginx config files.

### TemplateFile

The path to the file that contains the nginx config templates.

The template language used is [Go's text/template](https://golang.org/pkg/text/template/).

### PostRunCommand

ngman will try to execute this command after it has made any changes to nginx configuration files.

### GenerateCertCommand

ngman will try to execute this command when it needs to generate a new SSL certificate.
It will pass the respective domain name as the first argument.


## Usage

    ngman list                                                               
    ngman create <domain> <root-path>                   
    ngman add-static <domain> <root-path> <uri-location>                     
    ngman add-proxy <domain> <endpoint> <uri-location>                       
    ngman edit <domain>                                                      
    ngman delete <domain>                                                    
    ngman write-all

## Advanced Usage

### PHP-FPM Support

You can add the key UsePHP to a site config to enable PHP-FPM support.

Example:

    Domain = 'example.org'
    RootPath = '/var/www/example'
    UsePHP = true # <-- this is the important part

ngman will then insert the template called **"php-fpm-support"** from the **"nginx.txt"** file into the nginx configuration of that site.

### Misc. Options

The TOML files for the site configuration also allow adding an array of
miscellaneous options to the nginx config file.

Every string to be found in the array called **"MiscOptions"** in a site configuration
will be inserted as a single line into the nginx config file.

Example:

    Domain = 'example.com'
    RootPath = '/var/www/example.com'
    MiscOptions = [
        'gzip on',
        'gzip_disable "msie6"',
        'gzip_vary on',
        'gzip_proxied any'
    ]

**Note:** The semicolon is appended automatically in the config template.

### Chunks

When creating the nginx configuration files, ngman also looks into
a directory called chunks for files named like the domain name.

    config.SiteStorageDirectory + "/chunks/" + domain
    eg. example.org -> /root/.ngman/sites/chunks/example.org

These configuration chunks are inserted into the nginx site configuration file as is.
This mechanism can be used for further customizations of individual sites.

### Wildcard certificate support

ngman will assume that any subdomain will require a wildcard certificate.

So if you add a site with a domain like **"example.org"** a normal LetsEncrypt certificate will be generated.

However, if you add a site with a domain like **"sub.example.org"** a wildcard certificate will be generated and used.

Subsequent sites with a domain like **"foo.example.org"** will then also use the same wildcard certificate.

NOTE: In order to use wildcard support, you will have to provide the file **"~/.ngman/dnsprovider.env"**.
This file should contain the credentials for your DNS provider.

See [lego's documentation](https://go-acme.github.io/lego/dns/) for more information.

Example:

    root@dallas:~/.ngman# cat dnsprovider.env
    DNS_PROVIDER=dode    
    DODE_TOKEN=12345678901234677
    

## Standalone usage of ngman

**NOTE: ALL THE FOLLOWING IS ALREADY INCLUDED AND AUTOMATED IN THE [setup.sh file](https://github.com/memmaker/ngman/blob/main/setup.sh) and [mmx23/nginx image](https://hub.docker.com/repository/docker/mmx23/nginx)**

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

For certificate renewal, I suggest something like this

    for keyfile in $(sudo ls /ssl/certificates/ | grep key)
    do
        replaced="${keyfile/_/*}"
        renew_ssl_cert "${replaced%.*}" >> /tmp/cert.log
    done

Where renew_ssl_cert is the same as create_ssl_cert, but with the **run** command replaced by **renew**.

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
