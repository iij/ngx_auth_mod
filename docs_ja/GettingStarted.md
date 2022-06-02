# はじめに
## 概要

このドキュメントは ngx\_auth\_mod のインストール チュートリアルを提供します。  
以下の手順でインストールを進めていきます。

 1. nginx のインストール
 2. ngx\_auth\_mod のビルド
 3. ngx\_auth\_mod モジュールのインストール
 4. 動作確認

サンプルを提示しながら、上記の順番で説明していきます。  
以降では、ディストリビューションを Ubuntu として説明していきます。  
また提示するサンプルは、リバースプロキシと Web サーバの構成に ngx\_auth\_mod の認証処理を追加するものです。

## 1. nginx のインストール

nginx は Web サーバやリバースプロキシなどの機能が提供できます。  
nginx の概要は[こちら](https://nginx.org/en/)を参照ください。

### nginx のソースからインストール

nginx の簡易インストールは[こちら](https://nginx.org/en/linux_packages.html)を参照することで実施できます。  
ここでは ngx\_auth\_mod の前提となる[auth request module](https://nginx.org/en/docs/http/ngx_http_auth_request_module.html)を含めたインストール手順を示します。

nginx のビルドを行うために以下の依存関係ファイルをインストールします。

 - `apt install build-essential libpcre3 libpcre3-dev zlib1g zlib1g-dev libssl-dev unzip`

nginx のソースコードは[こちら](https://nginx.org/download/)からダウンロードします。  
ダウンロードしたものに、以下のオプションを付けてコンパイルします。

 - `./configure --prefix=/usr/local/nginx --pid-path=/run/nginx.pid --error-log-path=/var/log/nginx/error.log --http-log-path=/var/log/nginx/access.log --with-http_ssl_module --with-http_v2_module --with-http_auth_request_module`
 - `make install`

参考までに nginx を systemd で起動するためのファイルを示します。

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

このファイルを `nginx.service` として `/lib/systemd/system/` の配下に配置します。

## 2. ngx\_auth\_mod のビルド

ここでは ngx\_auth\_mod のモジュールをビルドする手順を示します。

### Go のインストール

ngx\_auth\_mod は Go で記述されたプログラムです。  
ngx\_auth\_mod のモジュールをビルドするには Go をインストールする必要があります。

Go のダウンロードは[こちら](https://go.dev/doc/install)からダウンロードでき、インストールの手順も掲載されています。

### ngx\_auth\_mod のビルド手順

ngx\_auth\_mod のビルドは以下の手順で進めます。

1. git clone コマンドで本リポジトリのクローン
    - `git clone <THIS REPOSITORY>`
2. build.sh で ngx\_auth\_mod のモジュールをビルド
    - `bash /<YOUR WORKING DIRECTORY PATH>/ngx_auth_mod/build.sh`

ビルドが成功すると ngx\_auth\_mod/bin/ ディレクトリの配下に各モジュールのバイナリが作成されます。  
各モジュールは以下の通りです。

 - ngx\_ldap\_auth
    - [ngx\_ldap\_auth ドキュメント](../docs_ja/README.md#ngx_ldap_auth)
 - ngx\_ldap\_path\_auth
    - [ngx\_ldap\_path\_auth ドキュメント](../docs_ja/README.md#ngx_ldap_path_auth)
 - ngx\_header\_path\_auth
    - [ngx\_header\_path\_auth ドキュメント](../docs_ja/README.md#ngx_header_path_auth)
 - ngx\_ldap\_path2ldap\_auth
    - [ngx\_ldap\_path2ldap\_auth ドキュメント](../docs_ja/README.md#ngx_ldap_path2ldap_auth)

認証(/認可)のモジュールを用途に応じて選択し、任意の場所へ配置してください。  
以降の説明で扱うサンプルは ngx\_ldap\_auth を例に挙げます。

## 3. ngx\_auth\_mod モジュールのインストール

ここでは ngx\_auth\_mod モジュールのインストールを以下の流れで説明します。

1. LDAP 情報から認証(/認可)を設定する
2. nginx に認証モジュールを設定する
3. ngx\_auth\_mod モジュールを systemd 設定する

### LDAP 情報から認証(/認可)を設定する

ngx\_auth\_mod モジュールの認証(/認可)の制約を設定ファイルに記載します。  
設定ファイルのドキュメントは以下の通りで、サンプルは[こちら](../conf/)から入手できます。

 - ngx\_ldap\_auth モジュール
    - [ngx\_ldap\_auth 設定ファイル ドキュメント](../docs_ja/ngx_ldap_auth.md)
 - ngx\_ldap\_path\_auth モジュール
    - [ngx\_ldap\_path\_auth 設定ファイル ドキュメント](../docs_ja/ngx_ldap_path_auth.md)
 - ngx\_header\_path\_auth モジュール
    - [ngx\_header\_path\_auth 設定ファイル ドキュメント](../docs_ja/ngx_header_path_auth.md)
 - ngx\_ldap\_path2ldap\_auth モジュール
    - [ngx\_ldap\_path2ldap\_auth 設定ファイル ドキュメント](../docs_ja/ngx_ldap_path2ldap_auth.md)

ここでは ngx\_ldap\_auth モジュールを扱うため `auth-ldap.conf` を自身の環境にあわせて編集してください。  
また、`auth-ldap.conf` を任意の場所に配置してください。

### nginx に認証モジュールを設定する

nginx の設定ファイルは `/usr/local/nginx/conf/nginx.conf` に配置されています。  
`nginx.conf` に以下の行を追加します。

 - `include /usr/local/nginx/sites-enabled/*.conf;`

このディレクトリは、nginx を利用する上での慣習的なディレクトリであり、自身で用意してください。

 - `mkdir /usr/local/nginx/sites-available`
 - `mkdir /usr/local/nginx/sites-enabled`

次に、リバースプロキシ(443/TCP)、Webサーバ（80/TCP)、ngx\_ldap\_auth モジュール（任意のポート/TCP）を nginx で起動する設定ファイルを作成します。  
以下に、サンプルを示します。

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

認証情報を扱うため、SSL/TLSで認証を行うことをおすすめします。  
また、HTML ファイルの用意、ディレクトリパスのパラメータは自身の環境に併せて変更してください。

上述した nginx の設定ファイルを、以下のディレクトリに配置してください。

 - /usr/local/nginx/sites-available
 - /usr/local/nginx/sites-enabled

### ngx\_auth\_mod モジュールを systemd 設定する

参考までに ngx\_auth\_ldap モジュールを systemd で起動するためのファイルを示します。

```
[Unit]
Description=LDAP authentication service for nginx
After=nginx.service

[Service]
ExecStart=/<YOUR MODULE PATH>/ngx_ldap_auth /<YOUR MODULE CONFIG FILE PATH>/auth-ldap.conf
User=<ACCOUNT NAME>
Group=<GROUP NAME>
```

このファイルを `ngx_auth_ldap.service` として `/lib/systemd/system/` 配下に配置します。  

## 4. 動作確認

最後に、nginx および ngx\_auth\_mod モジュールを起動して動作確認を行います。

### nginx および ngx\_auth\_mod モジュールの起動

以下のコマンドで、デーモンとして起動します。

 - `systemctl start nginx.service`
 - `systemctl start ngx_auth_ldap.service`

### 認証処理の確認

以下のコマンドで簡易的に認証処理を確認できます。

 - `curl https://127.0.0.1 -H --basic -u <USER>:<PASS>`

良い認証認可ライフを<3
