# ngx\_header\_path\_auth

**ngx\_header\_path\_auth** is a module for [nginx auth request module](http://nginx.org/en/docs/http/ngx_http_auth_request_module.html) that authorizes with the user name and path information set in the HTTP header.

## Error handling

On error, the process terminates with an unsuccessful status. 

## How to start

Run it on the command line like this:

```
ngx_header_path_auth <config file>
```

Since it does not provide background execution functions such as daemonization,
start it via a process management system such as systemd.

## Configuration file format

See the [auth request module documentation](http://nginx.org/en/docs/http/ngx_http_auth_request_module.html) for how to configure nginx.

The **ngx\_header\_path\_auth** configuration file is in TOML format, and the following is a sample configuration file.


```ini
socket_type = "tcp"
socket_path = "127.0.0.1:9202"
path_header = "X-Authz-Path"
user_header = "X-Forwarded-User"

[authz]
user_map = "/etc/ngx_auth_mod/usermap.conf"

path_pattern = "^/([^/]+)/"
nomatch_right = "@admin"
default_right = "*/

[authz.path_right]
"test" = "@dev"
```

Each parameter of the configuration file is as follows.

* **socket\_type** - Set this parameter to tcp(TCP socket) or unix(UNIX domain socket).
* **socket\_path** - Set the IP address and port number for tcp, and UNIX domain socket file path for unix.
* **path\_header** - A HTTP header that sets the path used for authorization processing. The default value is `X-Authz-Path`. In the appropriate place of the nginx configuration file, use `proxy_set_header` directive to set the HTTP header. (Eg `proxy_set_header X-Authz-Path $request_uri;`)
* **user\_header** - A HTTP header to set the user name. The default value is `X-Forwarded-User`. In the appropriate place of the nginx configuration file, use `proxy_set_header` directive to set the HTTP header. (Eg `proxy_set_header X-Forwarded-User $remote_user;`)

* \[authz\] part
	* **user_map** - User name and group name mapping file. More on this in the "_**user\_map** file details_" section.
	* **path\_pattern** - A regular expression that extracts the authorization judgment string from the path of the header specified by **path_header**. The extracted string is used for the key in **path\_right**. Use the `()` subexpression regular expression only once to specify the extraction location.
	* **nomatch\_right** - Authorization rights when the **path\_pattern** regular expression is not matched. For more information on authorization rights, see "_Authorization rights details_" section.
	* **default\_right** - Authorization rights when it matches the **path\_pattern**の regular expression and is not specified in **path\_right**. For more information on authorization rights, see "_Authorization rights details_".
	* **path\_right** - Authorization rights map for each extracted string when matching **path\_pattern** regular expression. Specify the extraction string as the key. For more information on authorization rights, see "_Authorization rights details_" section.

## Authorization rights details

In \[authz\] part, **nomatch\_right**, **default\_right**, and **path\_right** table value specify a character string that combines the following judgment descriptions with `|`. The combined judgment process is calculated by logical disjunction("OR"). If the result is true, it is authorized.

* '' - (empty string) Always considers true regardless of the user name.
* '!' - Always considers false regardless of the user name.
* '*' - If the user name exists, it is considered true.
* '@groupname` - The character string after @ is treated as a group name. True if the group contains users. Groups are defined in the **user_map** file.
* '@' - (no group name)　True if the user is described in the **user_map** file.
* 'username` True if the user name matches.

## **user\_map** file details

**user\_map** is a text file that defines users and groups.
This text file defines a user-group mapping, with a user name and group names (None or more is possible) on each line, as shown below.

``` plaintext
user1:group1 group2 ...
...
```

Separate the user name and group name with `:`. If there are multiple group names, separate them with ` ` (space character).

