[auth request module]: http://nginx.org/en/docs/http/ngx_http_auth_request_module.html

# ngx\_rewrite\_switch\_auth

**ngx\_rewrite\_switch\_auth** is an authentication module that calls the selected authentication via HTTP based on a username pattern.  
Only when the called authentication is successful will this authentication be successful.
This module can also restrict authentication by regular expressions for usernames.

The following types of external authentication are supported:

- HTTP "Basic" authentication
    - "Basic" authentication is converted to [auth request module] authentication by rewriting cache control headers and HTTP status.
    - You can use "Basic" authentication for another site.
- Authentications for [auth request module]
    - Cache settings, username and headers can be rewritten.

The following items can be changed:

- Username and password
- HTTP Header
- HTTP Cache Settings

## Error handling

On error, the process terminates with an unsuccessful status.

## How to start

Run it on the command line like this:

```
ngx_rewrite_switch_auth <config file>
```

Since it does not provide background execution functions such as daemonization,
start it via a process management system such as systemd.

## Configuration file format

See the [auth request module] documentation for how to configure nginx.

The **ngx\_rewrite\_switch\_auth** configuration file is in TOML format, and the following is a sample configuration file.

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

Each parameter of the configuration file is as follows.

### Root part

| Parameter | Description |
| :--- | :--- |
| **socket\_type** | Set this parameter to tcp(TCP socket) or unix(UNIX domain socket). |
| **socket\_path** | Set the IP address and port number for tcp, and UNIX domain socket file path for unix. |
| **cache\_seconds** | Cache duration in seconds passed to nginx upon successful authentication. If the value is 0, cache will not be used. <br>See [Authentication Cache Control](proxy_cache.md) for details. |
| **neg\_cache\_seconds** | Cache duration in seconds passed to nginx upon failed authentication. If the value is 0, cache will not be used. <br>See [Authentication Cache Control](proxy_cache.md) for details. |
| **use\_etag** | Set to `true` if you want to validate the cache using the `ETag` tag. <br>See [Authentication Cache Control](proxy_cache.md) for details. |
| **use\_serialized\_auth** | Set to `true` if you want authentication to be serialized for each account. <br>When authentications for the same account conflict, the authentication will be blocked and delayed. |
| **auth\_realm** | HTTP realm string. |
| **user\_agent** | `User-Agent` header value when calling for authentication. |

### **\[response.ok\]** part

| Parameter | Description |
| :--- | :--- |
| **code** | The HTTP response status code indicates authorized requests. (Default value: `200`)<br>This value is used by the [auth request module]. Therefore, Malfunctions may be caused by the incorrect setting value. |
| **message** |  The HTTP response message indicates authorized requests. (Default value: `"Authorized"`) |

### **\[\[switch\]\]** 部分

| Parameter | Description |
| :--- | :--- |
| **username_re** | A regular expression for username. Only when a match is found will authentication be called. <br>The match result is also used in [Authentication Parameter Rewriting](auth_rewriting.md). |
| **auth\_url** | URL for authentication. `http`, `https` and `unix`(HTTP over Unix domain sockets) schemes are supported. <br>The `unix` scheme is written in the form `unix:/run/backend.socket:/path/` or `unix:/run/backend.socket`. |
| **set\_username** | The format of the Basic authentication username to be set. See [Authentication Parameter Rewriting](auth_rewriting.md) for details on how to write the parameters. |
| **set\_password** | The format of the Basic authentication password to be set. See [Authentication Parameter Rewriting](auth_rewriting.md) for details on how to write the parameters. |
| **skip\_cert\_verify** | Set to 1 to ignore the certificate check result. |
| **root\_ca\_files** | A list of PEM files for the CA certificate. Used when the LDAP server is using a certificate from a private CA. |
| **timeout** | HTTP call timeout for authentication(unit: ms). (Default value: `1000`) |

### **\[switch.set\_header\]** part

This is a table of HTTP headers to be sent to authentication. The table keys and values have the following meanings:

- **key**.
    - HTML header name.
- **value**.
    - The format of the HTML header value. See [Authentication Parameter Rewriting](auth_rewriting.md) for details on how to write the parameters. 

### **\[switch.response.unauth\]** part

| Parameter | Description |
| :--- | :--- |
| **code** | The HTTP response status code indicates unauthenticated requests. (Default value: `401`)<br>This value is used by the [auth request module]. Therefore, Malfunctions may be caused by the incorrect setting value. |
| **message** | The HTTP response message indicates unauthenticated requests. (Default value: `"Not authenticated"`) |

### **\[switch.response.forbidden\]** part

| Parameter | Description |
| :--- | :--- |
| **code** | The HTTP response status code indicates failed authorization requests. (Default value: `403`)<br>This value is used by the [auth request module]. Therefore, Malfunctions may be caused by the incorrect setting value. |
| **message** | The HTTP response message indicates failed authorization requests. (Default value: `"Forbidden"`) |

### **\[switch.response.bad\_request\]** part

| Parameter | Description |
| :--- | :--- |
| **code** | the HTTP response status code when there is a problem with the authentication request. (Default value: `400`)<br>This value is used by the [auth request module]. Therefore, Malfunctions may be caused by the incorrect setting value. |
| **message** | the HTTP response message when there is a problem with the authentication request. (Default value: `"Bad request"`) |

### **\[switch.response.bad\_gateway\]** part

| Parameter | Description |
| :--- | :--- |
| **code** | the HTTP response status code when authentication communication fails. (default value is `502`)<br>This value is used by the [auth request module]. Therefore, Malfunctions may be caused by the incorrect setting value. | 
| **message** | the HTTP response message when authentication communication fails. (Default value: `"Bad gateway"`) |

### **\[switch.response.timeout\]** part

| Parameter | Description |
| :--- | :--- |
| **code** | HTTP response status code when timeout. (Default value: `504`)<br>This value is used by the [auth request module]. Therefore, Malfunctions may be caused by the incorrect setting value. | 
| **message** | HTTP response message when timeout (default value is `"Timeout"`) |

### **\[switch.response.invalid\_setting\]** part

| Parameter | Description |
| :--- | :--- |
| **code** | the HTTP response status code when there is a problem with the authentication parameters. (Default value: `500`)<br>This value is used by the [auth request module]. Therefore, Malfunctions may be caused by the incorrect setting value. | 
| **message** | the HTTP response message when there is a problem with the authentication parameters. (Default value: `"Invalid setting"`) |
