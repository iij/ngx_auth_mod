# Authentication Parameter Rewriting

The following module calls authentication with the rewritten HTTP request value.

- [ngx\_rewrite\_auth](ngx_rewrite_auth.md)
- [ngx\_rewrite\_and\_auth](ngx_rewrite_and_auth.md)
- [ngx\_rewrite\_switch\_auth](ngx_rewrite_switch_auth.md)

This document describes the format used to rewrite this authentication parameter.

## Authentication parameter format

The value of the HTTP request element is written in the following format.

| format | converted string | 
| :--- | :--- | 
| `$$` | `$` | 
| `$u` | Basic authentication username | 
| `$p` | Basic authentication password | 
| `$h{header name}` | HTTP header string for the given header name | 
| `$0` | **username\_re** string matching the whole pattern of regular expression of `$0` | 
| `$1` to `$9` | string matching the regular expression group of **username\_re** (corresponds to the 1st to 9th of the regular expression group) | 
| other than above | described string |

## Configurable Authentication Parameters

The authentication parameters that can be set are as follows:

| HTTP request elements | Parameter names in the configuration file | Supplemental |
| :--- | :--- | :--- |
| Basic authentication username | **set\_usename** | If you want to set the original username, set **set\_usename** to `$u`. The default value is empty. | 
| Basic authentication password | **set\_password** | If you want to set the original password, set **set\_password** to `$p`. The default value is empty. |
| HTTP Header | **\[\[set\_header\]\]**<br>**\[\[and.set\_header\]\]**<br>**\[\[swtich.set\_header\]\]** | By default, the only headers to be set are `User-Agent` and `Cache-Control` |

## Conversion examples

Below are conversion examples.

| **username\_re** regular expression | input value | setting format | converted value |
| :--- | :--- | :--- | :--- |
| empty string | username is `user1` | `$u@example.com` | `user1@example.com` |
| empty string | password is `secret` |`$p` | `secret` | 
| `^([^@]+)@example.com$` | username is `user2@example.com` | `$1` | `user2` | 
| `^[^@]+@(. +)$` | username is `user3@example.com` | `$1` | `example.com` | 
| empty string | `X-Authz-Path` header value is `markdown` |`/$h{X-Authz-Path}/` | `/markdown/` |
