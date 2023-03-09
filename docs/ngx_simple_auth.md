# ngx\_simple\_auth

**ngx\_simple\_auth** is a module for [nginx auth request module](http://nginx.org/en/docs/http/ngx_http_auth_request_module.html) that authenticates with the account information in the configuration file.  
The purpose of this module is only to check the settings of [nginx auth request module](http://nginx.org/en/docs/http/ngx_http_auth_request_module.html).
For "Basic" authentication, please use nginx's [auth basic module](http://nginx.org/en/docs/http/ngx_http_auth_basic_module.html).

Since it does not use external data for authentication, it can be used to check the settings of the [nginx auth request module](http://nginx.org/en/docs/http/ngx_http_auth_request_module.html).

## Error handling

On error, the process terminates with an unsuccessful status. 

## How to start

Run it on the command line like this:

```
ngx_simple_auth <config file>
```

Since it does not provide background execution functions such as daemonization,
start it via a process management system such as systemd.

## Configuration file format

See the [auth request module documentation](http://nginx.org/en/docs/http/ngx_http_auth_request_module.html) for how to configure nginx.

The **ngx\_simple\_auth** configuration file is in TOML format, and the following is a sample configuration file.

```ini
socket_type = "tcp"
socket_path = "127.0.0.1:9200"
#cache_seconds = 0
auth_realm = "TEST Authentication"

[password]
admin1 = "hoge"
user1 = "hogehoge"
```

Each parameter of the configuration file is as follows.

| Parameter | Description |
| :--- | :--- |
| **socket\_type** | Set this parameter to tcp(TCP socket) or unix(UNIX domain socket). |
| **socket\_path** | Set the IP address and port number for tcp, and UNIX domain socket file path for unix. |
| **cache\_seconds** | The cache duration in seconds to pass to nginx. However, if its value is 0, it will not use the cache.<br>See [Authentication Cache Control](proxy_cache.md) for details.|
| **auth\_realm** | HTTP realm string. |
| **[password]** | User-password mapping data in TOML table format. |
