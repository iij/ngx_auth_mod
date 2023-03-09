# Authentication cache control

ngx\_auth\_mod has a mechanism to cache authentication results in the [nginx proxy module](http://nginx.org/en/docs/http/ngx_http_proxy_module.html) for busy server.

# Disable cache

If the processing speed is sufficient, caching is not necessary.
And caching authentication results for nginx [auth request module](http://nginx.org/en/docs/http/ngx_http_auth_request_module.html) without understanding is dangerous.  
Therefore, if it is difficult to configure, caching of authentication results should be disabled.

## Disable nginx cache

Using the [proxy\_cache directive](http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_cache), only the cache of the authentication process can be disabled, as in the below example.

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

# Enable cache

To enable the caching of authentication results, both nginx and ngx\_auth\_mod must be configured.

Below, configurations for nginx and ngx\_auth\_mod are described individually.

## How to enable cache on nginx side

To enable nginx caching, the following directives require setting.

1. Use [proxy_temp_path](http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_temp_path) to set the temporary storage area.
2. Use [proxy_cache_path](http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_cache_path) to set a dedicated storage area for authentication results.
3. Use [proxy_cache_revalidate](http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_cache_revalidate) to enable nginx cache validation.
4. Use [proxy_cache_key](http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_cache_key) to set the cache key.

For security reasons, It is dangerous to mix the cache of authentication results with other caches.
In addition, if the cache of authentication results is stored in a separate location, you may flush only cache the authentication results easily.

Since the configuration differs depending on the module used, the explanation is divided as follows.

- nginx cache area settings (common settings for all modules)
- `proxy_cache_key` setting for ngx_ldap_auth
- `proxy_cache_key` settings other than ngx_ldap_auth

### nginx cache area settings (common settings for all modules)

Below is an example configuration of `proxy_temp_path`, `proxy_cache_path` and `proxy_cache_revalidate`.

```
proxy_cache_path /var/cache/nginx/auth levels=2:2 keys_zone=zone_auth:1m max_size=1g inactive=1m;
proxy_temp_path  /var/cache/nginx_tmp;
proxy_cache_revalidate on;
```

In this case, the `inactive` time of [proxy_cache_path](http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_cache_path) should not set long.
Too long `inactive` time will cause problems when changing authentication information.
For example, the old password can be used to authenticate for a long time.

### `proxy_cache_key` setting for ngx_ldap_auth

[ngx_ldap_auth](ngx_ldap_auth.md) only handles the authentication process.
Therefore, using the account name as the cache key reduces cache duplication and makes caching more efficient.
Specifically, set it like `proxy_cache_key "$remote_user";`.

Below is an example configuration.

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

            proxy_set_header Context-Length "";
            proxy_pass_request_body off;
            proxy_pass http://$auth_req;

            deny all;
    }
    # ...
}
```

### `proxy_cache_key` settings other than ngx\_ldap\_auth

Modules other than ngx_ldap_auth (such as [ngx_ldap_path_auth](ngx_ldap_path_auth.md)) result in different authorization results depending on the URL path.
Therefore, the URL path must be included in the cache key.
For example, set it like `proxy_cache_key "$request_uri $remote_user";`.

Two example configurations are described below.

#### Example 1: Using `$request_uri`

In this example, set `proxy_cache_key "$request_uri $remote_user";`.  
The following is a sample configuration.

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

#### Example 2: Using Converted URL

In this example, a regular expression is used to convert `$request_uri` to a cache key.
The regular expression used should be the same as the `path_pattern` parameter of ngx\_auth\_mod.
This configuration is more efficient than Example 1 because it roughly caches at the granularity of the first-level directory.
However, please set this up carefully, using the mysterious [rewrite module](http://nginx.org/en/docs/http/ngx_http_rewrite_module.html).

The process is outlined below.

1. Set the `$auth_path_id` variable using the [if](http://nginx.org/en/docs/http/ngx_http_rewrite_module.html#if) and [set](http://nginx.org/en/docs/http/ngx_http_rewrite_module.html#set) nginx directives.
2. Use `$auth_path_id` variable for cache key.

This example assumes that `path_pattern = "^/([^/]+)/"` is set in the module of ngx\_auth\_mod, so please read it accordingly.  
Below is an example of such a configuration.

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

## How to set cache on ngx\_auth\_mod side

The ngx\_auth\_mod cache configuration is divided into the following sections.

- Cache duration setting
- Enabling cache validation

### Cache duration setting

The following modules of ngx\_auth\_mod support cache period setting.

- [ngx_header_path_auth](ngx_header_path_auth.md)
- [ngx_ldap_auth](ngx_ldap_auth.md)
- [ngx_ldap_path_auth](ngx_ldap_path_auth.md)
- [ngx_ldap_path2ldap_auth](ngx_ldap_path2ldap_auth.md)

The cache duration can be set in seconds with **cache\_seconds** parameter in the configuration file.
However, if **cache\_seconds** parameter is set to 0, the cache will not be used.

Below is an example of setting it to 5 seconds.

```
cache_seconds = 5
```

### Enabling cache validation

Caution: Without understanding, enabling cache validation can be dangerous.

The following modules can enable cache validation using the `ETag` and `If-None-Match` headers.

- [ngx_ldap_auth](ngx_ldap_auth.md)
- [ngx_ldap_path_auth](ngx_ldap_path_auth.md)
- [ngx_ldap_path2ldap_auth](ngx_ldap_path2ldap_auth.md)

If the cache validation succeeds, the LDAP server call is skipped and the nginx cache period is updated.
Therefore, it is used to further reduce the load on the LDAP server.

To enable cache verification, add the below setting.

```
use_etag = true
```

If the above setting is made, the "Force cache update" process described below is necessary.

### Force cache update

If cache validation is enabled and the cache is continually accessed by the same account at intervals of less than the cache duration, the cache will no longer be updated.  
Therefore, if cache validation is enabled, the following process must be executed periodically to force the cache to be updated.

- Restart ngx\_auth\_mod authentication module.

Cache updates in this process are delayed by the cache duration, so the cache duration must also be short.  
The reason why a restart can force a cache update is because the `ETag` is modified at the startup time.
With that `ETag` modification, the cache validation always fails after the cache period, and thus the cache is updated.
