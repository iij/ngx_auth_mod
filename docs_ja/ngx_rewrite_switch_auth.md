[auth request module]: http://nginx.org/en/docs/http/ngx_http_auth_request_module.html

# ngx\_rewrite\_switch\_auth

ユーザー名のパターンに基づいて選ばれた認証を、HTTP経由で呼び出す認証モジュールです。  
この認証は、呼び出された認証が成功した時だけ成功します。
このモジュールは、ユーザ名に対する正規表現で認証を制限も可能です。

利用できる外部認証には、以下の種類があります。

- HTTP Basic認証
    - キャッシュ制御のヘッダーやHTTPステータスを書き換えて、[auth request module]認証に変換されます。
    - 別のサイトのBasic認証を利用できます。
- [auth request module]用認証
    - キャッシュ設定やユーザ名やヘッダーが書き換え可能です。

以下の項目が変更可能です。

- ユーザおよびパスワード
- HTTPヘッダー
- HTTPキャッシュの設定

## エラー処理

エラー時は、プロセスが異常ステータスで終了します。 

## 実行方法

コマンドラインは、以下の通りです。

```
ngx_rewrite_switch_auth 設定ファイル名
```

自前ではdaemon化等のバックグラウンド実行の機能は提供しません。  
systemd等のプロセス管理のシステムから起動してください。

## 設定ファイル書式

nginx側の設定方法については、[auth request module]のドキュメントを参照してください。

ngx\_rewrite\_authの設定ファイルは、TOMLフォーマットで、以下がサンプルです。

```ini
socket_type = "tcp"
socket_path = "127.0.0.1:9200"
cache_seconds = 5
neg_cache_seconds = 2
use_etag = true
#use_serialized_auth = false
auth_realm = "TEST Authentication"
#user_agent = "ngx_auth_mod"

[response.ok]
#code=200
message="Authorized0"

[[switch]]
username_re = '^[^@]+$'
auth_url = "http://127.0.0.1:9300"
set_username = "$0"
set_password = "$p"
#skip_cert_verify = false
#root_ca_files = [
#	"/etc/ssl/certs/Local-CA-Chain.cer",
#]
timeout = 5000

[switch.set_header]
"X-Remote-User" = "$u"

[switch.response.unauth]
#code=401
message="Not authenticated1"

[switch.response.forbidden]
#code=403
message="Forbidden1"

[switch.response.bad_request]
#code=400
message="Bad request1"

[switch.response.bad_gateway]
#code=501
message="Bad gateway1"

[switch.response.timeout]
#code=504
message="Timeout1"

[switch.response.invalid_setting]
#code=500
message="Invalid setting1"

[[switch]]
username_re = '^([^@]+)@example\.com$'
auth_url = "http://127.0.0.1:9200"
set_username = "$1"
set_password = "$p"
#skip_cert_verify = false
#root_ca_files = [
#	"/etc/ssl/certs/Local-CA-Chain.cer",
#]
timeout = 5000

[switch.set_header]
"X-Remote-User" = "$u"

[switch.response.unauth]
#code=401
message="Not authenticated2"

[switch.response.forbidden]
#code=403
message="Forbidden2"

[switch.response.bad_request]
#code=400
message="Bad request2"

[switch.response.bad_gateway]
#code=502
message="Bad gateway2"

[switch.response.timeout]
#code=504
message="Timeout2"

[switch.response.invalid_setting]
#code=500
message="Invalid setting2"
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
| **user\_agent** | 認証を呼び出す時の`User-Agent`ヘッダーの値です。 |

### **\[response.ok\]** 部分

|パラメータ名|意味|
| :--- | :--- |
| **code** | 認可された時のHTTPレスポンスステータスコード(デフォルト値は`200`)<br>この値は[auth request module]によって利用されるため、変更すると誤動作の可能性があります。 |
| **message** | 認可された時のHTTPレスポンスメッセージ(デフォルト値は`"Authorized"`) |

### **\[\[switch\]\]** 部分

|パラメータ名|意味|
| :--- | :--- |
| **username\_re** | ユーザ名の正規表現です。マッチした時だけ認証を呼びだします。<br>マッチ結果は[認証パラメータの書き換え](auth_rewriting.md)でも使われます。|
| **auth\_url** | 認証用URLです。`http`、`https`および`unix`(Unixドメインソケット経由のHTTP)のスキームをサポートしています。<br>`unix`スキームは、`unix:/run/backend.socket:/path/`もしくは`unix:/run/backend.socket`の形式で書きます。|
| **set\_username** | 設定されるBasic認証ユーザー名の形式です。記述方法については、[認証パラメータの書き換え](auth_rewriting.md)の説明を参照してください。 |
| **set\_password** | 設定されるBasic認証パスワードの形式です。記述方法については、[認証パラメータの書き換え](auth_rewriting.md)の説明を参照してください。 |
| **skip\_cert\_verify** | 証明書のチェック結果を無視する場合は1を指定します。 |
| **root\_ca\_files** | CA証明書のPEMファイルのリストです。認証サーバが、プライベートCAによる証明書を利用している時に使います。 |
| **timeout** | 認証のHTTP呼び出しのタイムアウト(単位はms)です。(デフォルト値は`1000`) |

### **\[switch.set\_header\]** 部分

認証に渡すHTTPヘッダーのテーブルです。テーブルのパラメータ名と値の意味は以下の通りです。

- **パラメータ名(キー**
    - HTMLヘッダー名です。
- **値**
    - HTMLヘッダーの値の形式です。記述方法については、[認証パラメータの書き換え](auth_rewriting.md)の説明を参照してください。


### **\[switch.response.unauth\]** 部分

|パラメータ名|意味|
| :--- | :--- |
| **code** | 未認証時のHTTPレスポンスステータスコード(デフォルト値は`401`)<br>この値は[auth reque  st module]によって利用されるため、変更すると誤動作の可能性があります。 | 
| **message** | 未認証時のHTTPレスポンスメッセージ(デフォルト値は`"Not authenticated"`) |

### **\[switch.response.forbidden\]** 部分

|パラメータ名|意味|
| :--- | :--- |
| **code** | 認可失敗時のHTTPレスポンスステータスコード(デフォルト値は`403`)<br>この値は[auth request module]によって利用されるため、変更すると誤動作の可能性があります。 |
| **message** | 認可失敗時のHTTPレスポンスメッセージ(デフォルト値は`"Forbidden"`) |

### **\[switch.response.bad\_request\]** 部分

|パラメータ名|意味|
| :--- | :--- |
| **code** | 認証リクエストに問題がある時のHTTPレスポンスステータスコード(デフォルト値は`400`)<br>この値は[auth request module]によって利用されるため、変更すると誤動作の可能性があります。 | 
| **message** | 認証リクエストに問題がある時のHTTPレスポンスメッセージ(デフォルト値は`"Bad request"`) |

### **\[switch.response.bad\_gateway\]** 部分

|パラメータ名|意味|
| :--- | :--- |
| **code** | 認証の通信が失敗した時のHTTPレスポンスステータスコード(デフォルト値は`502`)<br>この値は[auth request module]によって利用されるため、変更すると誤動作の可能性があります。 | 
| **message** | 認証の通信が失敗した時のHTTPレスポンスメッセージ(デフォルト値は`"Bad gateway"`) |

### **\[switch.response.timeout\]** 部分

|パラメータ名|意味|
| :--- | :--- |
| **code** | タイムアウトした時のHTTPレスポンスステータスコード(デフォルト値は`504`)<br>この値は[auth request module]によって利用されるため、変更すると誤動作の可能性があります。 | 
| **message** | タイムアウトした時のHTTPレスポンスメッセージ(デフォルト値は`"Timeout"`) |

### **\[switch.response.invalid\_setting\]** 部分

|パラメータ名|意味|
| :--- | :--- |
| **code** | 認証パラメータに問題がある時のHTTPレスポンスステータスコード(デフォルト値は`500`)<br>この値は[auth request module]によって利用されるため、変更すると誤動作の可能性があります。 | 
| **message** | 認証パラメータに問題がある時のHTTPレスポンスメッセージ(デフォルト値は`"Invalid setting"`) |
