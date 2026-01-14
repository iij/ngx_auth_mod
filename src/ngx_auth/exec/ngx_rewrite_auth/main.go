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

	"ngx_auth/htstat"
	"ngx_auth/http_auth"
	"ngx_auth/htval"
)

const DEFAULT_USER_AGENT = "ngx_auth_mod"
const DEFAULT_TIMEOUT int = 1000

func die(format string, v ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", v...)
	os.Exit(1)
}

func warn(format string, v ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", v...)
}

type RewriteAuthConfig struct {
	SocketType        string
	SocketPath        string
	CacheSeconds      uint `toml:",omitempty"`
	NegCacheSeconds   uint `toml:",omitempty"`
	UseEtag           bool `toml:",omitempty"`
	UseSerializedAuth bool `toml:",omitempty"`
	AuthRealm         string
	UserAgent         string `toml:",omitempty"`

	UsernameRe     htval.HtPattern `toml:",omitempty"`
	AuthUrl        http_auth.HtAuthAddr
	SetUsername    htval.HtValue     `toml:",omitempty"`
	SetPassword    htval.HtValue     `toml:",omitempty"`
	SetHeader      htval.HtSetHeader `toml:",omitempty"`
	SkipCertVerify bool              `toml:",omitempty"`
	RootCaFiles    []string          `toml:",omitempty"`
	Timeout        int               `toml:",omitempty"`

	Response htstat.HttpStatusTbl `toml:",omitempty"`
}

var (
	SocketType         string
	SocketPath         string
	CacheSeconds       uint = 0
	NegCacheSeconds    uint = 0
	UseEtag            bool
	UseSerializedAuth  bool
	AuthRealm          string
	UserAgent          string = DEFAULT_USER_AGENT
	Timeout            int    = DEFAULT_TIMEOUT
	SystemHttpResponse htstat.HttpStatusTbl
)

var StartTimeMS int64

var RewriteAgent *http_auth.RewriteAgent

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

	cfg := &RewriteAuthConfig{}
	if err := toml.NewDecoder(cfg_f).Decode(&cfg); err != nil {
		die("Config file parse error: %s", err)
	}

	SocketType = cfg.SocketType
	SocketPath = cfg.SocketPath

	if SocketType != "tcp" && SocketType != "unix" {
		die("Bad socket type: %s", SocketType)
	}

	CacheSeconds = cfg.CacheSeconds
	NegCacheSeconds = cfg.NegCacheSeconds
	UseEtag = cfg.UseEtag
	UseSerializedAuth = cfg.UseSerializedAuth

	if cfg.AuthRealm == "" {
		die("relm is required")
	}
	AuthRealm = cfg.AuthRealm

	if cfg.UserAgent != "" {
		UserAgent = cfg.UserAgent
	}

	if cfg.AuthUrl.IsZero() {
		die("auth_url is required")
	}

	if cfg.Timeout < 0 {
		die("bad timeout: %d", cfg.Timeout)
	}
	if cfg.Timeout > 0 {
		Timeout = cfg.Timeout
	}

	cfg.Response.SetDefault()
	if !cfg.Response.IsValid() {
		die("response code config error.")
		return
	}
	SystemHttpResponse = cfg.Response

	RewriteAgent = http_auth.NewRewriteAgent(&http_auth.RewriteAgentConfig{
		UsernameRe: cfg.UsernameRe.Regexp,

		UserAgent: UserAgent,
		AuthAddr:  cfg.AuthUrl,

		SetUsernameVal: cfg.SetUsername,
		SetPasswordVal: cfg.SetPassword,
		SetHeadersVal:  cfg.SetHeader,

		SkipCertVerify: cfg.SkipCertVerify,
		RootCaFiles:    cfg.RootCaFiles,

		Timeout: time.Duration(Timeout) * time.Millisecond,
	})

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
		os.Chmod(SocketPath, 0o777)
	}

	serr := srv.Serve(lstn)
	switch serr {
	case nil:
	case http.ErrServerClosed:
	default:
		die("HTTP server error: %v.", serr)
	}
}
