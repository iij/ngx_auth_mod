# check\_ldap

**check\_ldap** is a program to check the operation of LDAP authentication process using [ngx\_ldap\_auth](ngx_ldap_auth.md) or [ngx\_ldap\_path\_auth](ngx_ldap_path_auth.md) configuration file.  

## Error handling

On error, the process terminates with an unsuccessful status. 

## How to start

Run it on the command line like this:

```
check_ldap <config file> <user name>
```

Execute with the configuration file and user name as arguments.
After execution, enter the password and the result will be output.

The configuration file is the same as [ngx\_ldap\_auth](ngx_ldap_auth.md).
