# ngx\_ldap\_auth

nginxの[auth request module](http://nginx.org/en/docs/http/ngx_http_auth_request_module.html)に、LDAPのbind処理の結果を流用して認証を提供するモジュールです。

## エラー処理

エラー時は、プロセスが異常ステータスで終了します。 

## 実行方法

コマンドラインは、以下の通りです。

```
ngx_ldap_auth 設定ファイル名
```

自前ではdaemon化等のバックグラウンド実行の機能は提供しません。  
systemd等のプロセス管理のシステムから起動してください。

LDAPの情報で認証ユーザを制限したい場合は、LDAPのsearch処理のフィルタを使って(**uniq\_filter**の設定)、工夫してください。

## 設定ファイル書式

nginx側の設定方法については、[auth request moduleのドキュメント](http://nginx.org/en/docs/http/ngx_http_auth_request_module.html)を参照してください。

ngx_ldap_authの設定ファイルは、TOMLフォーマットで、以下がサンプルです。

```ini
socket_type = "tcp"
socket_path = "127.0.0.1:9200"
#cache_seconds = 0
#use_etag = true
auth_realm = "TEST Authentication"

host_url = "ldaps://ldap.example.com"
start_tls = 0
#skip_cert_verify = 0
root_ca_files = [
	"/etc/ssl/certs/Local-CA-Chain.cer",
]

base_dn = "DC=example,DC=com"
bind_dn = "CN=%s,OU=Users,DC=example,DC=com"
uniq_filter = "(&(objectCategory=person)(objectClass=user)(memberOf=CN=Group1,DC=example,DC=com)(userPrincipalName=%s@example.com))"
timeout = 5000
```

設定ファイルの各パラメータの意味は以下のとおりです。

|パラメータ名|意味|
| :--- | :--- |
| **socket\_type** | tcp(TCPソケット)とunix(Unixドメインソケット)が指定できます。 |
| **socket\_path** | tcpの場合はIPアドレスとポート番号、unixの場合はソケットファイルのファイルパスを指定します。 |
| **cache\_seconds** | nginxに渡すキャッシュ期間の秒数です。ただし、その値が0の場合、キャッシュを利用しなくなります。<br>詳細については[認証キャッシュ制御](proxy_cache.md)を参照してください。 |
| **use_etag** | `true`に設定すると`ETag`タグを使ったキャッシュの検証を行なうようになります。<br>詳細については[認証キャッシュ制御](proxy_cache.md)を参照してください。 |
| **auth\_realm** | HTTPのrealmの文字列です。 |
| **host\_url** | LDAPサーバの接続アドレスのURLです。パス部分は利用しません。 |
| **start\_tls** | TLSのStartTLSを利用する場合は1を指定します。 |
| **skip\_cert\_verify** | 証明書のチェック結果を無視する場合は1を指定します。 |
| **root\_ca\_files** | CA証明書のPEMファイルのリストです。LDAPサーバが、プライベートCAによる証明書を利用している時に使います。 |
| **base\_dn** | LDAPサーバに接続するときのbase DNです。 |
| **bind\_dn** | LDAPのbind処理を行う時に使うbind DNです。`%s`が含まれているとリモートユーザ名を埋め込みます。`%%`が含まれていると`%`に変換します |
| **uniq\_filter** | 設定された場合、bind処理のあとこの値をフィルターに指定してsearch処理が実施されます。その結果応答されたDNが1つだった場合以外は、認証の失敗として扱います。この値を指定しない場合は、bind処理の結果だけで判定が行われます。 |
| **timeout** | LDAPサーバとの通信に利用するタイムアウト時間(単位はms)です。 |
