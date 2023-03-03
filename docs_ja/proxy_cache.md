
# 認証キャッシュ制御

ngx\_auth\_modは、アクセス負荷が高い状況向けに、[nginx proxy module](http://nginx.org/en/docs/http/ngx_http_proxy_module.html)へ認証結果をキャッシュさせる仕組みを持っています。

# キャッシュの無効化

処理速度が十分であれば、キャッシュする必要はありません。
そして、nginxの[auth request module](http://nginx.org/en/docs/http/ngx_http_auth_request_module.html)向けの認証結果を、理解なくキャッシュすると危険です。  
ですから、もし、設定が難しいのなら、認証結果のキャッシュは無効にすべきです。

## nginxキャッシュの無効化

[proxy\_cacheディレクティブ](http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_cache)を使えば、以下の例のように認証処理のキャッシュだけを無効にできます。

```
server {
    # ...
    auth_request /auth;
    location = /auth {
        proxy_cache off;
        # ....
    }
    # ...
}
```

# キャッシュの有効化

認証結果のキャッシュを有効にするには、nginxとngx\_auth\_modの両方の設定が必要です。

以下で、nginxとngx\_auth\_modの設定を個別に説明します。

## nginx側のキャッシュの有効方法

nginxのキャッシュを有効する為には、以下のディレクティブの設定が必要です。

1. [proxy_temp_path](http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_temp_path) 一時保存領域の設定をする。
2. [proxy_cache_path](http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_cache_path) 認証結果専用の保存領域を設定する。
3. [proxy_cache_revalidate](http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_cache_revalidate) nginxのキャッシュ検証を有効化する。
4. [proxy_cache_key](http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_cache_key) キャッシュキーを設定する。

セキュリティ上、認証結果のキャッシュを他のキャッシュと混ぜるのは危険です。
また、認証結果のキャッシュの保存場所を分けておけば、認証結果だけを簡単に削除できます。

利用するモジュールによって設定が違うので、以下のように分けて説明します。

-  nginxのキャッシュ領域の設定（全モジュール共通の設定）
- ngx_ldap_auth向け`proxy_cache_key`設定
- ngx_ldap_auth以外の`proxy_cache_key`設定

### nginxのキャッシュ領域の設定（全モジュール共通の設定）

以下は、`proxy_temp_path`、`proxy_cache_path`および`proxy_cache_revalidate`の設定例です。

```
proxy_cache_path /var/cache/nginx/auth levels=2:2 keys_zone=zone_auth:1m max_size=1g inactive=1m;
proxy_temp_path  /var/cache/nginx_tmp;
proxy_cache_revalidate on;
```

この用途では、[proxy_cache_path](http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_cache_path)の`inactive`パラメータは、長くしてはいけません。
長すぎる`inactive`パラメータは、認証情報の変更時に問題を起こします。
例えば、長期間、古いパスワードで認証できてしまったりします。

### ngx_ldap_auth向け`proxy_cache_key`設定

[ngx_ldap_auth](ngx_ldap_auth.md)は認証処理のみしかしていません。
そのため、キャッシュキーをアカウント名にするとキャッシュの重複が減り、キャッシュの効率が良くなります。
具体的には、`proxy_cache_key "$remote_user";`のように設定します。

以下は、その設定例です。

```
proxy_cache_path /var/cache/nginx/auth levels=2:2 keys_zone=zone_auth:1m max_size=1g inactive=1m;
proxy_temp_path  /var/cache/nginx_tmp;

server {
    # ...
    set $auth_req 127.0.0.1:9200;

    # ...

    auth_request /auth;
    location = /auth {
            proxy_cache_key "$remote_user";
            proxy_cache zone_auth;
            proxy_cache_revalidate on;

            proxy_set_header Context-Length "";
            proxy_pass_request_body off;
            proxy_pass http://$auth_req;
f
            deny all;
    }
    # ...
}
```

### ngx\_ldap\_auth以外の`proxy_cache_key`設定

ngx_ldap_auth以外のモジュール([ngx_ldap_path_auth](ngx_ldap_path_auth.md)など)は、URLのパスによって認可結果が変わります。
その為、URLパスをキャッシュキーに含めることが必要です。
例えば、`proxy_cache_key "$request_uri $remote_user";`のように設定します。

以下で、2つの設定例を説明します。

#### 例1: `$request_uri`を利用

この例では、`proxy_cache_key "$request_uri $remote_user";`を設定します。  
以下が、設定サンプルです。

```
proxy_cache_path /var/cache/nginx/auth levels=2:2 keys_zone=zone_auth:1m max_size=1g inactive=1m;
proxy_temp_path  /var/cache/nginx_tmp;

server {
    # ...
    set $auth_req 127.0.0.1:9200;
    # ...

    auth_request /auth;
    location = /auth {
            proxy_intercept_errors off;
            proxy_cache_key "$request_uri $remote_user";
            proxy_cache zone_auth;

            proxy_set_header Context-Length "";
            proxy_pass_request_body off;
            proxy_set_header X-Authz-Path $request_uri;

            proxy_pass http://$auth_req;

            deny all;
    }

    # ...
}
```

#### 例2: 変換したURLを利用

この例では、正規表現により`$request_uri`をキャッシュキーに変換します。
利用する正規表現は、ngx\_auth\_modの`path_pattern`パラメータと同じものにします。
この構成では、第一階層ディレクトリの粒度で荒くキャッシュを行うため、例1よりも効率的です。
ただし、黒魔術的な[rewrite module](http://nginx.org/en/docs/http/ngx_http_rewrite_module.html)を使っているので、慎重に設定してください。

処理の概要は以下の通りです。

1. [if](http://nginx.org/en/docs/http/ngx_http_rewrite_module.html#if)および[set](http://nginx.org/en/docs/http/ngx_http_rewrite_module.html#set)のnginxディレクティブをつかって、`$auth_path_id`変数を設定する。
2. キャッシュキーに、`$auth_path_id`変数を利用する。

この設定例では、ngx\_auth\_modのモジュールにに`path_pattern = "^/([^/]+)/"`の設定がされていることを前提にしていますので、適宜読み替えてください。  
以下が、その設定例です。

```
proxy_cache_path /var/cache/nginx/auth levels=2:2 keys_zone=zone_auth:1m max_size=1g inactive=1m;
proxy_temp_path  /var/cache/nginx_tmp;

server {
    # ...
    set $auth_req 127.0.0.1:9200;
    # ...

    auth_request /auth;
    location = /auth {
            proxy_intercept_errors off;
            proxy_cache_key "$auth_path_id $remote_user";
            proxy_cache zone_auth;

            proxy_set_header Context-Length "";
            proxy_pass_request_body off;
            proxy_set_header X-Authz-Path $request_uri;

            proxy_pass http://$auth_req;

            deny all;
    }

    set $auth_path_id "/";
    if ($request_uri ~ "^/([^/]+)/") {
        set $auth_path_id $1;
    }

    # ...
}
```

## ngx\_auth\_mod側のキャッシュ設定

ngx\_auth\_modのキャッシュ設定は、以下に分けて説明します。

- キャッシュ期間の設定
- キャッシュ検証の有効化方法

### キャッシュ期間の設定

ngx\_auth\_modの以下のモジュールがキャッシュ期間の設定に対応しています。

- [ngx_header_path_auth](ngx_header_path_auth.md)
- [ngx_ldap_auth](ngx_ldap_auth.md)
- [ngx_ldap_path_auth](ngx_ldap_path_auth.md)
- [ngx_ldap_path2ldap_auth](ngx_ldap_path2ldap_auth.md)

キャッシュ期間は、設定ファイルの**cache\_seconds**パラメータで秒数で設定できます。
ただし、**cache\_seconds**パラメータを0にした場合はキャッシュを利用しなくなります。

以下は、5秒に設定する場合の例です。

```
cache_seconds = 5
```

### キャッシュ検証の有効化方法

注意: 理解せずにキャッシュ検証を有効にすると危険です。

以下のモジュールでは、`ETag`および`If-None-Match`ヘッダーを使ったキャッシュ検証を有効化できます。

- [ngx_ldap_auth](ngx_ldap_auth.md)
- [ngx_ldap_path_auth](ngx_ldap_path_auth.md)
- [ngx_ldap_path2ldap_auth](ngx_ldap_path2ldap_auth.md)

キャッシュ検証が成功すると、LDAPサーバの呼び出しを省略し、nginxのキャッシュ期間が更新されます。
そのため、LDAPサーバの負荷をさらに減らしたいときに使います。

キャッシュ検証を有効にする場合は、以下の設定を追加します。

```
use_etag = true
```

上記の設定をした場合は、後述の「キャッシュの強制更新」の作業が必須です。

### キャッシュの強制更新

キャッシュ検証が有効な場合に、同じアカウントから、キャッシュ期間未満の間隔でアクセスされ続けると、キャッシュが更新されなくなります。  
このため、キャッシュ検証を有効にした場合は、キャッシュを強制的に更新させるために、以下の処理の定期的実行が必要です。

- ngx\_auth\_modの認証モジュールを再起動する。

この処理でのキャッシュ更新はキャッシュ期間分だけ遅延するので、キャッシュ期間も短めにする必要があります。  
再起動によってキャッシュの強制更新が出来る理由は、起動時刻で`ETag`が変更されるからです。
その`ETag`の変更で、キャッシュ期間後のキャッシュ検証は常に失敗するので、キャッシュが更新されるのです。
