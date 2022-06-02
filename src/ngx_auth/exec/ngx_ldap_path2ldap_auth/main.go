package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"syscall"

	"github.com/l4go/task"
	"github.com/naoina/toml"

	"ngx_auth/ldap_auth"
)

func die(format string, v ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", v...)
	os.Exit(1)
}

func warn(format string, v ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", v...)
}

type NgxLdapPathAuthConfig struct {
	SocketType string
	SocketPath string
	AuthRealm  string `toml:",omitempty"`
	PathHeader string `toml:",omitempty"`

	Ldap struct {
		HostUrl        string
		StartTls       int      `toml:",omitempty"`
		SkipCertVerify int      `toml:",omitempty"`
		RootCaFiles    []string `toml:",omitempty"`
		BaseDn         string
		BindDn         string
		UniqFilter     string `toml:",omitempty"`
		Timeout        int    `toml:",omitempty"`
	}

	Authz struct {
		PathPattern   string
		BanNomatch    bool              `toml:",omitempty"`
		NomatchFilter string            `toml:",omitempty"`
		BanDefault    bool              `toml:",omitempty"`
		DefaultFilter string            `toml:",omitempty"`
		PathFilter    map[string]string `toml:",omitempty"`
	}
}

var SocketType string
var SocketPath string
var AuthRealm string
var LdapAuthConfig *ldap_auth.Config

var PathHeader = "X-Authz-Path"
var PathPatternReg *regexp.Regexp

var UniqueFilter string
var BanNomatch bool
var NomatchFilter string
var BanDefault bool
var DefaultFilter string
var PathFilter map[string]string

func init() {
	flag.CommandLine.SetOutput(os.Stderr)
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(),
			"Usage: %s [options ...] <config_file>\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.CommandLine.SetOutput(os.Stderr)

	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	cfg_f, err := os.Open(flag.Arg(0))
	if err != nil {
		die("Config file open error: %s", err)
	}
	defer cfg_f.Close()

	cfg := &NgxLdapPathAuthConfig{}
	if err := toml.NewDecoder(cfg_f).Decode(&cfg); err != nil {
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

	if cfg.PathHeader != "" {
		PathHeader = cfg.PathHeader
	}

	UniqueFilter = cfg.Ldap.UniqFilter
	LdapAuthConfig = &ldap_auth.Config{
		HostUrl:        cfg.Ldap.HostUrl,
		StartTls:       cfg.Ldap.StartTls != 0,
		SkipCertVerify: cfg.Ldap.SkipCertVerify != 0,
		RootCaFiles:    cfg.Ldap.RootCaFiles,
		BaseDn:         cfg.Ldap.BaseDn,
		BindDn:         cfg.Ldap.BindDn,
		UniqueFilter:   UniqueFilter,
		Timeout:        cfg.Ldap.Timeout,
	}

	PathPatternReg, err = regexp.Compile(cfg.Authz.PathPattern)
	if err != nil {
		die("path pattern error: %s", cfg.Authz.PathPattern)
		return
	}

	BanNomatch = cfg.Authz.BanNomatch
	NomatchFilter = cfg.Authz.NomatchFilter
	if BanNomatch && NomatchFilter != "" {
		warn("nomatch_filter is not used because ban_nomatch is true.")
	}

	BanDefault = cfg.Authz.BanDefault
	DefaultFilter = cfg.Authz.DefaultFilter
	if BanDefault && DefaultFilter != "" {
		warn("default_filter is not used because ban_default is true.")
	}

	PathFilter = cfg.Authz.PathFilter
}

var ErrUnsupportedSocketType = errors.New("unsupported socket type.")

func listen(cc task.Canceller, stype string, spath string) (net.Listener, error) {
	lcnf := &net.ListenConfig{}

	switch stype {
	default:
		return nil, ErrUnsupportedSocketType
	case "unix":
	case "tcp":
	}

	return lcnf.Listen(cc.AsContext(), stype, spath)
}

func main() {
	signal_chan := make(chan os.Signal, 1)
	signal.Notify(signal_chan, syscall.SIGINT, syscall.SIGTERM)

	srv := &http.Server{Addr: SocketPath}

	cc := task.NewCancel()
	defer cc.Cancel()
	go func() {
		select {
		case <-cc.RecvCancel():
		case <-signal_chan:
			cc.Cancel()
		}
		srv.Close()
	}()

	http.HandleFunc("/", TestAuthHandler)

	lstn, lerr := listen(cc, SocketType, SocketPath)
	switch lerr {
	case nil:
	case context.Canceled:
	default:
		die("socket listen error: %v.", lerr)
	}
	if SocketType == "unix" {
		defer os.Remove(SocketPath)
		os.Chmod(SocketPath, 0777)
	}

	serr := srv.Serve(lstn)
	switch serr {
	case nil:
	case http.ErrServerClosed:
	default:
		die("HTTP server error: %v.", serr)
	}
}
