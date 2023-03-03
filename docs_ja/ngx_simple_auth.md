# ngx\_simple\_auth

[nginx auth request module](http://nginx.org/en/docs/http/ngx_http_auth_request_module.html)に、設定ファイルだけの単純な認証を提供するモジュールです。  
このモジュールの用途は、[nginx auth request module](http://nginx.org/en/docs/http/ngx_http_auth_request_module.html)の設定確認のみです。
Basic認証の用途には、nginxの[auth basic module](http://nginx.org/en/docs/http/ngx_http_auth_basic_module.html)を利用してください。

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

nginx側の設定方法については、[auth request moduleのドキュメント](http://nginx.org/en/docs/http/ngx_http_auth_request_module.html)を参照してください。

**ngx\_simple\_auth**の設定ファイルは、TOMLフォーマットで、以下がサンプルです。

```ini
socket_type = "tcp"
socket_path = "127.0.0.1:9200"
#cache_seconds = 0
auth_realm = "TEST Authentication"

[password]
admin1 = "hoge"
user1 = "hogehoge"
```

設定ファイルの各パラメータの意味は以下のとおりです。

| パラメータ名 | 意味 |
| :--- | :--- |
| **socket\_type** | tcp(TCPソケット)とunix(Unixドメインソケット)が指定できます。 |
| **socket\_path** | tcpの場合はIPアドレスとポート番号、unixの場合はソケットファイルのファイルパスを指定します。 |
| **cache\_seconds** | nginxに渡すキャッシュ期間の秒数です。ただし、その値が0の場合、キャッシュを利用しなくなります。<br>詳細については[認証キャッシュ制御](proxy_cache.md)を参照してください。 |
| **auth\_realm** | HTTPのrealmの文字列です。 |
| **[password]** 部分 | TOML table形式のユーザーとパスワードのマッピングデータです。 |

