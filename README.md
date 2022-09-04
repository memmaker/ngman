# ngman

A simple CLI tool for managing nginx sites.

**Homepage / Demo:** https://textzentrisch.de/pages/ngman/

## Concept

A site is uniquely identified by the domain name.

This tool supports two types of locations: static & proxy.

A site can have multiple static and proxy locations.

The configuration and state of this tool is kept under its config directory.
By default, that is **"~/.ngman/"** and the tool will create it on first start.

The config directory also needs to contain **"nginx.txt"**, the file with all the partial templates.
You can easily adapt that file to your needs to adjust the nginx configurations created.

For every site that it manages the tool will create a file in the **~/.ngman/sites/** directory.

You can always re-create all the nginx config files by running **ngman write-all**.

## Global settings (config.json) 

Example config.json file in a production environment:

    {
        "CertificateRootPath": "/ssl/certificates",
        "SiteStorageDirectory": "/root/.ngman/sites",
        "NginxSiteConfigDirectory": "/etc/nginx/sites-enabled",
        "TemplateFile": "/root/.ngman/nginx.txt",
        "PostRunCommand": "service nginx reload",
        "GenerateCertCommand": "create_ssl_cert"
    }

### CertificateRootPath

The path to the directory where the SSL certificates are stored.
The files are expected to conform to the following naming scheme:

    <domain>.key
    <domain>.crt

### SiteStorageDirectory

The path to the directory where the JSON site configuration files are stored. 
Must be writable by the user executing the tool.

### NginxSiteConfigDirectory

This is the main output directory for the nginx config files.

### TemplateFile

The path to the file that contains the nginx config templates.

### PostRunCommand

ngman will try to execute this command after it has made any changes to nginx configuration files.

### GenerateCertCommand

ngman will try to execute this command when it needs to generate a new SSL certificate.
It will pass the respective domain name as the first argument.

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

    {
        "Domain": "example.com",
        "UsePHP": true,
        "RootPath": "/var/www/example.com"
    }

ngman will then insert the template called **"php-fpm-support"** from the **"nginx.txt"** file into the nginx configuration of that site.

### Misc. Options

The JSON files for the site configuration also allow adding an array of
miscellaneous options to the nginx config file.

Every string to be found in the array called **"MiscOptions"** in a site configuration
will be inserted as a single line into the nginx config file.

Example:

    {
        "Domain": "example.com",
        "RootPath": "/var/www/example.com",
        "MiscOptions": [
            "gzip on;",
            "gzip_disable "msie6";",
            "gzip_vary on;",
            "gzip_proxied any;"
        ]
    }

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