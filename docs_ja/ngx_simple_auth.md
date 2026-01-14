[auth request module]: http://nginx.org/en/docs/http/ngx_http_auth_request_module.html
[auth basic module]: http://nginx.org/en/docs/http/ngx_http_auth_basic_module.html
# ngx\_simple\_auth

nginx [auth request module]に、設定ファイルだけの単純な認証を提供するモジュールです。  
このモジュールの用途は、nginx [auth request module]の設定確認のみです。
Basic認証の用途には、nginxの[auth basic module]を利用してください。

## エラー処理

エラー時は、プロセスが異常ステータスで終了します。 

## 実行方法

コマンドラインは、以下の通りです。

```
ngx_simple_auth 設定ファイル名
```

自前ではdaemon化等のバックグラウンド実行の機能は提供しません。  
systemd等のプロセス管理のシステムから起動してください。

## 設定ファイル書式

nginx側の設定方法については、[auth request module]のドキュメントを参照してください。

**ngx\_simple\_auth**の設定ファイルは、TOMLフォーマットで、以下がサンプルです。

```ini
socket_type = "tcp"
socket_path = "127.0.0.1:9200"
#cache_seconds = 0
#neg_cache_seconds = 0
#use_etag = false
auth_realm = "TEST Authentication"

[password]
admin1 = "hoge"
user1 = "hogehoge"

#[response.ok]
#code=200
#message="Authorized"

#[response.unauth]
#code=401
#message="Not authenticated"
```

設定ファイルの各パラメータの意味は以下のとおりです。

### ルート部分

| パラメータ名 | 意味 |
| :--- | :--- |
| **socket\_type** | tcp(TCPソケット)とunix(Unixドメインソケット)が指定できます。 |
| **socket\_path** | tcpの場合はIPアドレスとポート番号、unixの場合はソケットファイルのファイルパスを指定します。 |
| **cache\_seconds** | 認証成功時にnginxに渡される秒のキャッシュ期間です。その値が0の場合、キャッシュを利用しなくなります。<br>詳細については[認証キャッシュ制御](proxy_cache.md)を参照してください。 |
| **neg\_cache\_seconds** | 認証失敗時にnginxに渡される秒のキャッシュ期間です。その値が0の場合、キャッシュを利用しなくなります。<br>詳細については[認証キャッシュ制御](proxy_cache.md)を参照してください。 |
| **use\_etag** | `ETag`タグを使ったキャッシュの検証を行いたい場合は、`true`に設定してください。<br>詳細については[認証キャッシュ制御](proxy_cache.md)を参照してください。 |
| **auth\_realm** | HTTPのrealmの文字列です。 |
| **[password]** 部分 | TOML table形式のユーザーとパスワードのマッピングデータです。 |

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
