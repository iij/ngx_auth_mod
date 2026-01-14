[auth request module]: http://nginx.org/en/docs/http/ngx_http_auth_request_module.html
# ngx\_ldap\_path\_auth

**ngx\_ldap\_path\_auth** is a module for nginx [auth request module] that authenticates using an LDAP bind operation, and authorizes by file path.

## Error handling

On error, the process terminates with an unsuccessful status. 

## How to start

Run it on the command line like this:

```
ngx_ldap_path_auth <config file>
```

Since it does not provide background execution functions such as daemonization,
start it via a process management system such as systemd.

If you want to limit authenticated users by LDAP information, use the LDAP search processing filter(use the **uniq\_filter** config parameter).

If you don't need file path authorization, use the [ngx\_ldap\_auth](ngx_ldap_auth.md) module.

## Configuration file format

See the [auth request module] documentation for how to configure nginx.

The **ngx\_ldap\_path\_auth** configuration file is in TOML format, and the following is a sample configuration file.

```ini
socket_type = "tcp"
socket_path = "127.0.0.1:9201"
#cache_seconds = 0
#use_etag = false
#use_serialized_auth = false
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
user_map = "/etc/ngx_auth_mod/usermap.conf"

path_pattern = "^/([^/]+)/"
nomatch_right = "@admin"
default_right = "*"

[authz.path_right]
"test" = "@dev"

#[response.ok]
#code=200
#message="Authorized"

#[response.unauth]
#code=401
#message="Not authenticated"

#[response.forbidden]
#code=403
#message="Forbidden"

#[response.nopath]
#code=403
#message="No path header"
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
| **path\_header** | A HTTP header that sets the path used for authorization processing. The default value is `X-Authz-Path`. In the appropriate place of the nginx configuration file, use `proxy_set_header` directive to set the HTTP header. (Eg `proxy_set_header X-Authz-Path $request_uri;`) |

### **\[ldap\]** part

| Parameter | Description |
| :--- | :--- |
| **host\_url** | The URL of the LDAP server connection address. The pass part is not used. |
| **start\_tls** | Set to 1 when using TLS STARTTLS. |
| **skip\_cert\_verify** | Set to 1 to ignore the certificate check result. |
| **root\_ca\_files** | A list of PEM files for the CA certificate. Used when the LDAP server is using a certificate from a private CA. |
| **base\_dn** | The base DN when connecting to the LDAP server. |
| **bind\_dn** | This is the bind DN when performing LDAP bind processing. Rewrite `%s` as the remote user name and `%%` as `%`. |
| **uniq\_filter** | Only if this value is set, search with this value filter. If the search result is one DN, the authentication will be successful. |
| **timeout** | Communication timeout(unit: ms) with the LDAP server. (Default value: `1000`) |

### **\[authz\]** part

| Parameter | Description |
| :--- | :--- |
| **user\_map\_config** | A file that specifies how user names and group names are handled in **user\_map**.  More on this in the "_**user\_map\_config** file details_" section. |
| **user\_map** | User name and group name mapping file. More on this in the "_**user\_map** file details_" section. |
| **path\_pattern** | A regular expression that extracts the authorization judgment string from the path of the header specified by **path\_header**. The extracted string is used for the key in **path\_right**. Use the `()` subexpression regular expression only once to specify the extraction location. |
| **nomatch\_right** | Authorization rights when the **path\_pattern** regular expression is not matched. For more information on authorization rights, see "_Authorization rights details_" section. |
| **default\_right** | Authorization rights when it matches the **path\_pattern** regular expression and is not specified in **path\_right**. For more information on authorization rights, see "_Authorization rights details_". |
| **path\_right** | Authorization rights map for each extracted string when matching **path\_pattern** regular expression. Specify the extraction string as the key. For more information on authorization rights, see "_Authorization rights details_" section. |

### **\[response.ok\]** part

| Parameter | Description |
| :--- | :--- |
| **code** | The HTTP response status code indicates authorized requests. (Default value: `200`)<br>This value is used by the [auth request module]. Therefore, Malfunctions may be caused by the incorrect setting value. |
| **message** |  The HTTP response message indicates authorized requests. (Default value: `"Authorized"`) |

### **\[response.unauth\]** part

| Parameter | Description |
| :--- | :--- |
| **code** | The HTTP response status code indicates unauthenticated requests. (Default value: `401`)<br>This value is used by the [auth request module]. Therefore, Malfunctions may be caused by the incorrect setting value. |
| **message** | The HTTP response message indicates unauthenticated requests. (Default value: `"Not authenticated"`) |

### **\[response.forbidden\]** part

| Parameter | Description |
| :--- | :--- |
| **code** | The HTTP response status code indicates failed authorization requests. (Default value: `403`)<br>This value is used by the [auth request module]. Therefore, Malfunctions may be caused by the incorrect setting value. |
| **message** | The HTTP response message indicates failed authorization requests. (Default value: `"Forbidden"`) |

### **\[response.nopath\]** part

| Parameter | Description |
| :--- | :--- |
| **code** | The HTTP response status code indicates an unexpected HTTP header in **path\_header**. (Default value: `403`)<br>This value is used by the [auth request module]. Therefore, Malfunctions may be caused by the incorrect setting value. |
| **message** | The HTTP response message indicates an unexpected HTTP header in **path\_header**. (Default value: `"No path header"`) |

## Authorization rights details

In **\[authz\]** part, **nomatch\_right**, **default\_right**, and **path\_right** table value specify a character string that combines the following judgment descriptions with `|`. The combined judgment process is calculated by logical disjunction("OR"). If the result is true, it is authorized.


| Authorization method | Description|
| :--- | :--- |
| empty string | Always considers true regardless of the user name. |
| `!` | Always considers false regardless of the user name. |
| `*` | If the user name exists, it is considered true. |
| `@groupname` | The character string after @ is treated as a group name. True if the group contains users. Groups are defined in the **user\_map** file. |
| `@` (no group name) | True if the user is described in the **user\_map** file. |
| user name | True if the user name matches. |

## **user\_map** file details

**user\_map** is a text file that defines users and groups.
This text file defines a user-group mapping, with a user name and group names (None or more is possible) on each line, as shown below.  

``` plaintext
user1:group1 group2 ...
...
```

Separate the user name and group name with `:`. If there are multiple group names, separate them with ` ` (space character). If you want to use`:` and ` `(space character) in user or group names, escape them with `\`.

## **user\_map\_config** file details

**user\_map\_config** is a file that defines the handling of user names and group names.   
This text file defines the available usernames and group names in regular expressions, as shown below.

```
user_regex = '^[a-z_][0-9a-z_\-]{0,32}$'
group_regex = '^[a-z_][0-9a-z_\-]{0,32}$'
```

| Parameter| Description|
| :--- | :--- |
| **user\_regex** | A regular expression of strings to allow as usernames. |
| **group\_regex** | Regular expression of strings to allow as group names. |
