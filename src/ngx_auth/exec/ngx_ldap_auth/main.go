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
	"syscall"
	"time"

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

type NgxLdapAuthConfig struct {
	SocketType   string
	SocketPath   string
	CacheSeconds uint32 `toml:",omitempty"`
	UseEtag      bool   `toml:",omitempty"`
	AuthRealm    string `toml:",omitempty"`

	HostUrl        string
	StartTls       int      `toml:",omitempty"`
	SkipCertVerify int      `toml:",omitempty"`
	RootCaFiles    []string `toml:",omitempty"`
	BaseDn         string
	BindDn         string
	UniqFilter     string `toml:",omitempty"`
	Timeout        int    `toml:",omitempty"`
}

var SocketType string
var SocketPath string
var CacheSeconds uint32
var UseEtag bool
var AuthRealm string
var LdapAuthConfig *ldap_auth.Config

var StartTimeMS int64

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

	cfg := &NgxLdapAuthConfig{}
	if err := toml.NewDecoder(cfg_f).Decode(&cfg); err != nil {
		die("Config file parse error: %s", err)
	}

	SocketType = cfg.SocketType
	SocketPath = cfg.SocketPath

	if SocketType != "tcp" && SocketType != "unix" {
		die("Bad socket type: %s", SocketType)
	}

	CacheSeconds = cfg.CacheSeconds
	UseEtag = cfg.UseEtag

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

	StartTimeMS = time.Now().UnixMicro()
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
