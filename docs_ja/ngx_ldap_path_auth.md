# ngx\_ldap\_path\_auth

nginxの[auth request module](http://nginx.org/en/docs/http/ngx_http_auth_request_module.html)に、LDAPのbind処理の結果を流用して認証と、ヘッダーに設定したパス情報を元にした認可を提供するモジュールです。

## エラー処理

エラー時は、プロセスが異常ステータスで終了します。 

## 実行方法

コマンドラインは、以下の通りです。

```
ngx_ldap_path_auth 設定ファイル名
```

自前ではdaemon化等のバックグラウンド実行の機能は提供しません。  
systemd等のプロセス管理のシステムから起動してください。

LDAPの情報で認証ユーザを制限したい場合は、LDAPのsearch処理のフィルタを使って(**uniq\_filter**の設定)、工夫してください。

パスごとの認可処理が必要ない場合は、[ngx\_ldap\_auth](ngx_ldap_auth.md)を使用してください。

## 設定ファイル書式

nginx側の設定方法については、[auth request moduleのドキュメント](http://nginx.org/en/docs/http/ngx_http_auth_request_module.html)を参照してください。

**ngx\_ldap\_path\_auth**の設定ファイルは、TOMLフォーマットで、以下がサンプルです。

```ini
socket_type = "tcp"
socket_path = "127.0.0.1:9201"
#cache_seconds = 0
#use_etag = true
auth_realm = "TEST Authentication"
path_header = "X-Authz-Path"

[ldap]
host_url = "ldaps://ldap.example.com"
start_tls = 0
#skip_cert_verify = 0
root_ca_files = [
	"/etc/ssl/certs/Local-CA-Chain.cer",
]

base_dn = "DC=group,DC=example,DC=com"
bind_dn = "CN=%s,OU=Users,DC=group,DC=example,DC=com"
uniq_filter = "(&(objectCategory=person)(objectClass=user)(memberOf=CN=Group1,DC=example,DC=com)(userPrincipalName=%s@example.com))"
timeout = 5000

[authz]
user_map_config = "/etc/ngx_auth_mod/usermap_config.conf"
user_map = "/etc/ngx_auth_mod/usermap.conf"

path_pattern = "^/([^/]+)/"
nomatch_right = "@admin"
default_right = "*"

[authz.path_right]
"test" = "@dev"
```

設定ファイルの各パラメータの意味は以下のとおりです。

### ルート部分

| パラメータ名 | 意味 | 
| :--- | :--- |
| **socket\_type** | tcp(TCPソケット)とunix(Unixドメインソケット)が指定できます。 |
| **socket\_path** | tcpの場合はIPアドレスとポート番号、unixの場合はソケットファイルのファイルパスを指定します。 |
| **cache\_seconds** | nginxに渡すキャッシュ期間の秒数です。ただし、その値が0の場合、キャッシュを利用しなくなります。<br>詳細については[認証キャッシュ制御](proxy_cache.md)を参照してください。 |
| **use_etag** | `true`に設定すると`ETag`タグを使ったキャッシュの検証を行なうようになります。<br>詳細については[認証キャッシュ制御](proxy_cache.md)を参照してください。 |
| **auth\_realm** | HTTPのrealmの文字列です。 |
| **path\_header** | 認可処理の使うパスを設定するHTTPヘッダーです。デフォルト値は`X-Authz-Path`です。nginxの設定ファイルの適切な箇所で、`proxy_set_header X-Authz-Path $request_uri;`などのように、ヘッダーの値を設定してください。 |

### **\[ldap\]** 部分

| パラメータ名 | 意味 |
| :--- | :--- |
| **host\_url** | LDAPサーバの接続アドレスのURLです。パス部  分は利用しません。 |
| **start\_tls** | TLSのStartTLSを利用する場合は1を指定します。 |
| **skip\_cert\_verify** | 証明書のチェック結果を無視する場合は1を指定します。 |
| **root\_ca\_files** | CA証明書のPEMファイルのリストです。LDAPサーバが、プライベートCAによる証明書を利用している時に使います。 |
| **base\_dn** | LDAPサーバに接続するときのbase DNです。 |
| **bind\_dn** | LDAPのbind処理を行う時に使うbind DNです。\%sが含まれているとリモートユーザ名を埋め込みます。\%\%が含まれていると\%に変換します |
| **uniq\_filter** | 設定された場合、bind処理のあとこの値をフィルターに指定してsearch処理が実施されます。その結果応答されたDNが1つだった場合以外は、認証の失敗として扱います。この値を指定しない場合は、bind処理の結果だけで判定が行われます。 |
| **timeout** | LDAPサーバとの通信に利用するタイムアウト時間(単位はms)です。 |

### **\[authz\]** 部分

| パラメータ名 | 意味 |
| :--- | :--- |
| **user\_map\_config** | user\_mapでの、ユーザ名とグループ名の扱いを指定するファイルです。ファイルの書式は別途説明します。 |
| **user\_map** | ユーザ名とグループ名のマッピングファイルです。ファイルの書式は別途説明します。 |
| **path\_pattern** | **path\_header**のヘッダで渡されたパス情報から認可判定を行う文字列を抽出する正規表現です。抽出された文字列は、**path\_right**で権限を指定するために使われます。`()`の正規表現を１つだけ使って、認可権限の判断に使う文字列部分を指定してください。抽出箇所の指定に`()`の正規表現を1回だけ使ってください。 |
| **nomatch\_right** | **path\_pattern**の正規表現のマッチが失敗した場合の認可権限です。認可権限の詳細は、「認可権限の詳細」の説明を見てください。 |
| **default\_right** | **path\_pattern**の正規表現のマッチが成功し、かつ、**path\_right**に該当のキーが無い場合の、認可権限です。認可権限の詳細は、「認可権限の詳細」の説明を見てください。 |
| **path\_right** | **path\_pattern**の正規表現のマッチに成功したときの、抽出文字列ごとの認可権限の設定です。抽出文字列をキーとして指定します。認可権限の詳細は、「認可権限の詳細」の説明を見てください。 |
 
## 認可権限の詳細

**\[authz\]**の**nomatch\_right**、**default\_right**、**path\_right**のテーブルの各要素の値は、以下の判定処理の記述を|で結合した文字列を指定します。結合された判定処理は、倫理和(or)で計算します。結果が真の場合は、認可されます。

|認可方法|意味|
| :--- | :--- |
| 空文字 | ユーザー名に関係なく真と判断します。 |
| `!` | ユーザ名に関係なく偽と判断します。 |
| `*` | ユーザ名が存在すれば、真と判断します。 |
| `@グループ名` | @の後ろをグループ名として扱い、そのグループにユーザが含まれる場合に真と判断します。グループは**user_map**ファイルで定義します。 |
| `@` | (@のみ、グループ名無し) **user_map**ファイルに利用者のユーザ名が記述されていれば、真と判断します。 |
| ユーザ名 | 利用者のユーザ名と一致する場合に真と判断します。 |

## **user\_map**ファイルの詳細
**user\_map**は、ユーザとグループを定義するテキストファイルを指定します。
このテキストファイルは、以下のように、各行にユーザ名と所属グループ名(無し及び複数も可能)を記述して、ユーザ名とグループ名のマッピングを表現します。  

``` plaintext
ユーザ名1:グループ名1 グループ名2 ...
...
```

ユーザ名とグループ名の間は`:`で区切ります。グループ名が複数の場合は` `(空白文字)で区切ります。`:`と` `(空白文字)をユーザ名やグループ名で使いたいときは、`\`でエスケープします。

## user\_map\_configファイルの詳細
**user\_map\_config**は、ユーザ名とグループ名の扱いを定義するファイルです。  
許容されるユーザ名とグループ名を、以下のように、正規表現で表現します。

```
user_regex = '^[a-z_][0-9a-z_\-]{0,32}$'
group_regex = '^[a-z_][0-9a-z_\-]{0,32}$'
```

| パラメータ名 | 意味 |
| :--- | :--- |
| **user\_regex** | ユーザ名として許可する文字列の正規表現です。 |
| **group\_regex** | グループ名として許可する文字列の正規表現です。 |
