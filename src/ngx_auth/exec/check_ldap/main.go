package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/bgentry/speakeasy"

	"ngx_auth/ldap_auth"
)

func die(format string, v ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", v...)
	os.Exit(1)
}

func warn(format string, v ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", v...)
}

var SocketType string
var SocketPath string
var AuthRealm string
var LdapAuthConfig *ldap_auth.Config
var Username string

func get_passwd() (string, error) {
	return speakeasy.Ask("Password: ")
}

func auth(user, pass string) bool {
	la, err := ldap_auth.NewLdapAuth(LdapAuthConfig)
	if err != nil {
		warn("Authenticate error: %s", err.Error())
		return false
	}
	defer la.Close()

	ok, _, err := la.Authenticate(user, pass)
	if err != nil {
		warn("Authenticate error: %s", err.Error())
		return false
	}

	return ok
}

func init() {
	flag.CommandLine.SetOutput(os.Stderr)
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(),
			"Usage: %s [options ...] <config_file> <user>\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.CommandLine.SetOutput(os.Stderr)

	flag.Parse()

	if flag.NArg() != 2 {
		flag.Usage()
		os.Exit(1)
	}

	cfg, err := load_config(flag.Arg(0))
	if err != nil {
		die("Config file parse error: %s", err)
	}

	SocketType = cfg.SocketType
	SocketPath = cfg.SocketPath

	if SocketType != "tcp" && SocketType != "unix" {
		die("Bad socket type: %s", SocketType)
	}

	if cfg.AuthRealm == "" {
		die("relm is required")
	}
	AuthRealm = cfg.AuthRealm

	LdapAuthConfig = &ldap_auth.Config{
		HostUrl:        cfg.HostUrl,
		StartTls:       cfg.StartTls != 0,
		SkipCertVerify: cfg.SkipCertVerify != 0,
		RootCaFiles:    cfg.RootCaFiles,
		BaseDn:         cfg.BaseDn,
		BindDn:         cfg.BindDn,
		UniqueFilter:   cfg.UniqFilter,
		Timeout:        cfg.Timeout,
	}

	Username = flag.Arg(1)
}

func main() {
	pass, err := get_passwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	ok := auth(Username, pass)
	fmt.Println("Result:", ok)
	if !ok {
		os.Exit(2)
	}
}
