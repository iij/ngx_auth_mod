# GettingStarted

## About

This document describes the installation procedure of the ngx\_auth\_mod.  
The outline of the installation procesure is shown below.

 1. Install nginx
 2. Build ngx\_auth\_mod modules
 3. Install ngx\_auth\_mod modules
 4. Confirm the operation

We will explain each steps with examples.  
In the rest of the document, we assume the target distribution is Ubuntu.  
The example shown here is to add an authentication function to the configuration of the reverse proxy and web server.


## 1. Install nginx

By using nginx, you can build a HTTP(S) and reverse proxy server etc.  
Click [here](https://nginx.org/en/) for an overview of nginx.

### How to install nginx from source code

This section explains how to install nginx from source code.  
Note that you need to include [auth request module](https://nginx.org/en/docs/http/ngx_http_auth_request_module.html) during the installation procedure, because ngx\_auth\_mod adds an authentication processing to the nginx auth request module.  
(Easy installation document can be found from [here](https://nginx.org/en/linux_packages.html))

Install the prerequisites for building nginx:

- `apt install build-essential libpcre3 libpcre3-dev zlib1g zlib1g-dev libssl-dev unzip`

Source code for nginx can be downloaded from [here](https://nginx.org/download/).  
To install nginx, run the following command:

 - `./configure --prefix=/usr/local/nginx --pid-path=/run/nginx.pid --error-log-path=/var/log/nginx/error.log --http-log-path=/var/log/nginx/access.log --with-http_ssl_module --with-http_v2_module --with-http_auth_request_module`
 - `make install`

Finally, place `nginx.service` as a config file of systemd under `/lib/systemd/system/`.  
This enables systemd to launch nginx.  
FYI, the following is an example of the `nginx.service` file.

```
[Unit]
Description=A high performance web server and a reverse proxy server
After=network.target

[Service]
Type=forking
PIDFile=/run/nginx.pid
ExecStartPre=/usr/local/nginx/sbin/nginx -t -q -g 'daemon on; master_process on;'
ExecStart=/usr/local/nginx/sbin/nginx -g 'daemon on; master_process on;'
ExecReload=/usr/local/nginx/sbin/nginx -g 'daemon on; master_process on;' -s reload
ExecStop=-/sbin/start-stop-daemon --quiet --stop --retry QUIT/5 --pidfile /run/nginx.pid
TimeoutStopSec=5
KillMode=mixed
User=<USER NAME>
Group=<GROUP NAME>

[Install]
WantedBy=multi-user.target
```

## 2. Build ngx\_auth\_mod modules

This section explains how to build ngx\_auth\_mod modules.

### How to install Go

ngx\_auth\_mod modules is written in Go.  
You need to install Go to build ngx\_auth\_mod modules.

To download and install Go, click [here](https://go.dev/doc/install).

### How to build ngx\_auth\_mod modules

Follow the steps below to build ngx\_auth\_mod modules.

1.  Clone ngx\_auth\_mod repository with the following command
    - `git clone <THIS REPOSITORY>`

2. Run the following command to build ngx\_auth\_mod modules:
    - `bash /<YOUR WORKING DIRECTORY PATH>/ngx_auth_mod/build.sh`

If the build succeeded, the executable files of each ngx\_auth\_mod module will be placed on ngx\_auth\_mod/bin/.  
The documents of each modules are available in separate pages.

 - ngx\_ldap\_auth module
    - [documentation of ngx\_ldap\_auth](../docs/README.md#ngx_ldap_auth)
 - ngx\_ldap\_path\_auth module
    - [documentation of ngx\_ldap\_path\_auth](../docs/README.md#ngx_ldap_path_auth)
 - ngx\_header\_path\_auth module
    - [documentation of ngx\_header\_path\_auth](../docs/README.md#ngx_header_path_auth)
 - ngx\_ldap\_path2ldap\_auth module
    - [documentation of ngx\_ldap\_path2ldap\_auth](../docs/README.md#ngx_ldap_path2ldap_auth)

Select any of the authentication module depending on your requirement and place it to the location you prefer.  
In the next section, the ngx\_ldap\_auth module will be used as an example.

## 3. Install ngx\_auth\_mod modules

In this section, we will explain how to install the ngx\_auth\_mod modules by following the steps shown below.

1. How to configure LDAP authentication(/authorization)
2. How to configure authentication of nginx
3. How to create a systemd configuration file for ngx\_auth\_mod

### How to configure LDAP authentication(/authorization)

Create authentication(/authorization) restrictions according to the module you are using.  
You can find an example config file from [here](../conf/).  
The documents of each configuration file are provided separately.

 - ngx\_ldap\_auth module
    - [documentation of ngx\_ldap\_auth configuration file](../docs/ngx_ldap_auth.md)
 - ngx\_ldap\_path\_auth module
    - [documentation of ngx\_ldap\_path\_auth configuration file](../docs/ngx_ldap_path_auth.md)
 - ngx\_header\_path\_auth module
    - [documentation of ngx\_header\_path\_auth configuration file](../docs/ngx_header_path_auth.md)
 - ngx\_ldap\_path2ldap\_auth module
    - [documentation of ngx\_ldap\_path2ldap\_auth configuration file](../docs/ngx_ldap_path2ldap_auth.md)

In this GettingStarted documentation, ngx\_ldap\_auth module is used as an example.  
Therefore, edit the `auth-ldap.conf` file to fit your LDAP schema.  
Also, place the `auth-ldap.conf` file anywhere.

### How to configure authentication of nginx

The configuration file (`nginx.conf`) is placed under `/usr/local/nginx/conf/`.  
Add the following line to the `nginx.conf` file.

 - `include /usr/local/nginx/sites-enabled/*.conf;`

Create a directory with the following command:

 - `mkdir /usr/local/nginx/sites-available`
 - `mkdir /usr/local/nginx/sites-enabled`

These directories are popular directries when customising nginx.

Next, create a new configuration file of nginx as `auth-webserver.conf`.  
This configuration file is used to launch reverse proxy (443/TCP), web server (80/TCP), and ngx\_auth\_mod module (Any port number/TCP).  
FYI, an example `auth-webserver.conf` is shown below.

```
# ngx_ldap_auth module service
upstream auth_req { server 127.0.0.1:<MODULE PORT>; }

# Reverse proxy(443/TCP)
server {
	listen 443 default ssl;
	server_name localhost;
	ssl_certificate <YOUR SSL CERTIFICATE FILE PATH>;
	ssl_certificate_key <YOUR SSL CERTIFICATE KEY FILE PATH>;
	ssl_protocols TLSv1.2 TLSv1.3;
	ssl_ciphers HIGH:!aNULL:!MD5;
	ssl_prefer_server_ciphers on;
	ssl_session_cache shared:SSL:10m;
	ssl_session_timeout 10m;
	client_max_body_size 1G;

	root /<YOUR WEB SERVER's ROOT DIRECTORY PATH>/var/www/dummy; # Don't create a dummy file.

	# Authentication URL path
	auth_request /auth;
	proxy_set_header Authorization "";
	location = /auth {
		proxy_set_header X-Forwarded-User $remote_user; # Required if using modules other than ngx_ldap_auth.
		proxy_set_header Context-Length "";
		proxy_pass_request_body off;
		proxy_pass http://auth_req;
	}

	proxy_intercept_errors on;
	error_page 400 403 404 500 502 503 /error/error.html;
	error_page 405 415 /error/;
	location /error/ {
		auth_request off;
		access_log off;
		add_header Cache-Control "maxage-86400, public";
	}

	location / {
		proxy_set_header Context-Length "";
		proxy_pass http://localhost:80/;
	}
}
# Web Server(80/TCP)
server {
	listen 80 default;

	server_name localhost;
	client_max_body_size 1G;
	root /<YOUR WEB SERVER's ROOT DIRECTORY PATH>/var/www;

	location / {
		index index.html;
	}

	proxy_intercept_errors on;
	error_page 400 403 404 500 502 503 /error/error.html;
	error_page 405 415 /error/;
	location /error/ {
		auth_request off;
		access_log off;
		add_header Cache-Control "maxage-86400, public";
	}
}
```

We recommend to use SSL/TLS protocols to use authentication information.  
Prepare HTML files by yourself.  
Finally, place the `auth-webserver.conf` file under the `sites-available` and `sites-enabled` directories.

### How to create a systemd configuration file of ngx\_auth\_mod module

Finally, place the `ngx_auth_ldap.service` file used by systemd under `/lib/systemd/system/`.  
This enables systemd to launch the ngx\_auth\_ldap module.  
FYI, the following is an example of the `ngx_auth_ldap.service` file.

```
[Unit]
Description=LDAP authentication service for nginx
After=nginx.service

[Service]
ExecStart=/<YOUR MODULE PATH>/ngx_ldap_auth /<YOUR MODULE CONFIG FILE PATH>/auth-ldap.conf
User=<ACCOUNT NAME>
Group=<GROUP NAME>
```

## 4. Confirm the operation

Finally, make sure the operation of nginx and ngx\_auth\_mod module.

### Launch of nginx and ngx\_auth\_mod service

Run the following commands to launch nginx and the ngx\_auth\_mod module:

 - `systemctl start nginx.service`
 - `systemctl start ngx_auth_ldap.service`

### Confirmation of authentication process

Run the following command to confirm the authentication process is working:

 - `curl https://127.0.0.1 -H --basic -u <USER>:<PASS>`

Have a good Authentication life <3
