[auth request module]: http://nginx.org/en/docs/http/ngx_http_auth_request_module.html
# ngx\_ldap\_auth

nginxの[auth request module]に、LDAPのbind処理の結果を流用して認証を提供するモジュールです。

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

nginx側の設定方法については、[auth request module]のドキュメントを参照してください。

ngx\_ldap\_authの設定ファイルは、TOMLフォーマットで、以下がサンプルです。

```ini
socket_type = "tcp"
socket_path = "127.0.0.1:9200"
#cache_seconds = 0
#neg_cache_seconds = 0
#use_etag = false
#use_serialized_auth = false
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

#[response.ok]
#code=200
#message="Authorized"

#[response.unauth]
#code=401
#message="Not authenticated"
```

設定ファイルの各パラメータの意味は以下のとおりです。

### ルート部分

|パラメータ名|意味|
| :--- | :--- |
| **socket\_type** | tcp(TCPソケット)とunix(Unixドメインソケット)が指定できます。 |
| **socket\_path** | tcpの場合はIPアドレスとポート番号、unixの場合はソケットファイルのファイルパスを指定します。 |
| **cache\_seconds** | 認証成功時にnginxに渡される秒のキャッシュ期間です。その値が0の場合、キャッシュを利用しなくなります。<br>詳細については[認証キャッシュ制御](proxy_cache.md)を参照してください。 |
| **neg\_cache\_seconds** | 認証失敗時にnginxに渡される秒のキャッシュ期間です。その値が0の場合、キャッシュを利用しなくなります。<br>詳細については[認証キャッシュ制御](proxy_cache.md)を参照してください。 |
| **use\_etag** | `ETag`タグを使ったキャッシュの検証を行いたい場合は、`true`に設定してください。<br>詳細については[認証キャッシュ制御](proxy_cache.md)を参照してください。 |
| **use\_serialized\_auth** | 認証を各アカウント毎に直列化したい場合は、`true`に設定してください。<br>同じアカウントの認証が衝突した場合、ブロックして遅延させます。 |
| **auth\_realm** | HTTPのrealmの文字列です。 |
| **host\_url** | LDAPサーバの接続アドレスのURLです。パス部分は利用しません。 |
| **start\_tls** | TLSのStartTLSを利用する場合は1を指定します。 |
| **skip\_cert\_verify** | 証明書のチェック結果を無視する場合は1を指定します。 |
| **root\_ca\_files** | CA証明書のPEMファイルのリストです。LDAPサーバが、プライベートCAによる証明書を利用している時に使います。 |
| **base\_dn** | LDAPサーバに接続するときのbase DNです。 |
| **bind\_dn** | LDAPのbind処理を行う時に使うbind DNです。`%s`が含まれているとリモートユーザ名を埋め込みます。`%%`が含まれていると`%`に変換します |
| **uniq\_filter** | 設定された場合、bind処理のあとこの値をフィルターに指定してsearch処理が実施されます。その結果応答されたDNが1つだった場合以外は、認証の失敗として扱います。この値を指定しない場合は、bind処理の結果だけで判定が行われます。 |
| **timeout** | LDAPサーバとの通信に利用するタイムアウト時間(単位はms)です。(デフォルト値は`1000`) |

### **\[response.ok\]** 部分

|パラメータ名|意味|
| :--- | :--- |
| **code** | 認可された時のHTTPレスポンスステータスコード(デフォルト値は`200`)<br>この値は[auth request module]によって利用されるため、変更すると誤動作の可能性があります。 |
| **message** | 認可された時のHTTPレスポンスメッセージ(デフォルト値は`"Authorized"`) |

### **\[response.unauth\]** 部分

|パラメータ名|意味|
| :--- | :--- |
| **code** | 未認証時のHTTPレスポンスステータスコード(デフォルト値は`401`)<br>この値は[auth reque  st module]によって利用されるため、変更すると誤動作の可能性があります。 | 
| **message** | 未認証時のHTTPレスポンスメッセージ(デフォルト値は`"Not authenticated"`) |
