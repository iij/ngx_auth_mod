# ngx\_ldap\_path2ldap\_auth

**ngx\_ldap\_path2ldap\_auth** is a module for is a module for [nginx auth request module](http://nginx.org/en/docs/http/ngx_http_auth_request_module.html) that authenticates using an LDAP bind operation, and authorizes by file path and LDAP information.

## Error handling

On error, the process terminates with an unsuccessful status. 

## How to start

Run it on the command line like this:

```
ngx_ldap_path2ldap_auth <config file>
```

Since it does not provide background execution functions such as daemonization,
start it via a process management system such as systemd.

## Configuration file format

See the [auth request module documentation](http://nginx.org/en/docs/http/ngx_http_auth_request_module.html) for how to configure nginx.

The **ngx\_ldap\_path2ldap\_auth** configuration file is in TOML format, and the following is a sample configuration file.

```ini
socket_type = "tcp"
socket_path = "127.0.0.1:9203"
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
uniq_filter = "(&(objectCategory=person)(objectClass=user)(userPrincipalName=%s@example.com))"
timeout = 5000

[authz]
path_pattern = "^/([^/]+)/"
#ban_nomatch = false

nomatch_filter = "" # for root directory files
ban_default = true
#default_filter = ""

[authz.path_filter]
"test" = "(&(objectCategory=person)(objectClass=user)(memberOf=CN=Group1,DC=example,DC=com)(userPrincipalName=%s@example.com))"
```

Each parameter of the configuration file is as follows.

* **socket\_type** - Set this parameter to tcp(TCP socket) or unix(UNIX domain socket).
* **socket\_path** - Set the IP address and port number for tcp, and UNIX domain socket file path for unix.
* **auth\_realm** - HTTP realm string.
* **path\_header** - A HTTP header that sets the path used for authorization processing. The default value is `X-Authz-Path`. In the appropriate place of the nginx configuration file, use `proxy_set_header` directive to set the HTTP header. (Eg `proxy_set_header X-Authz-Path $request_uri;`)

* \[ldap\] part
	* **host\_url** - The URL of the LDAP server connection address. The pass part is not used.
	* **start\_tls** - Set to 1 when using TLS STARTTLS.
	* **skip\_cert\_verify** - Set to 1 to ignore the certificate check result.
	* **root\_ca\_files** - A list of PEM files for the CA certificate. Used when the LDAP server is using a certificate from a private CA.
	* **base\_dn** - The base DN when connecting to the LDAP server.
	* **bind\_dn** - This is the bind DN when performing LDAP bind processing. Rewrite `%s` as the remote user name and `%%` as `%`.
	* **uniq\_filter** - Only if this value is set, search with this value filter. If the search result is one DN, the authentication will be successful.
	* **timeout** - Communication timeout(unit: ms) with the LDAP server.

* \[authz\] part
	* **path\_pattern** - A regular expression that extracts the authorization judgment string from the path of the header specified by **path_header**. The extracted string is used for the key in **path\_filter**. Use the `()` subexpression regular expression only once to specify the extraction location.
	* **ban\_nomatch** - If true, authorization will fail if the **path\_pattern** regular expression does not match. (As a result, **nomatch\_filter** is disabled.)
	* **nomatch\_filter** - LDAP filter for authorization when the **path\_pattern** regular expression is not matched. **nomatch\_filter** results is processed in the same way as **uniq\_filter**.
	* **ban\_default** - If true, authorization will fail if the **path\_pattern** regular expression does not match. (As a result, **default\_filter** is disabled.)
	* **default\_filter** - LDAP filter for authorization rights when it matches the **path\_pattern** regular expression and is not specified in **path\_filter**. **default\_filter** results is processed in the same way as **uniq\_filter**.