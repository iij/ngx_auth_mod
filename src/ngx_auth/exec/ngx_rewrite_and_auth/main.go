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

type AndConfig struct {
	UsernameRe htval.HtPattern `toml:",omitempty"`

	AuthUrl        http_auth.HtAuthAddr
	SetUsername    htval.HtValue     `toml:",omitempty"`
	SetPassword    htval.HtValue     `toml:",omitempty"`
	SetHeader      htval.HtSetHeader `toml:",omitempty"`
	SkipCertVerify bool              `toml:",omitempty"`
	RootCaFiles    []string          `toml:",omitempty"`
	Timeout        int               `toml:",omitempty"`

	Response htstat.HttpStatusTbl `toml:",omitempty"`
}

type RewriteAndConfig struct {
	SocketType        string
	SocketPath        string
	CacheSeconds      uint `toml:",omitempty"`
	NegCacheSeconds   uint `toml:",omitempty"`
	UseEtag           bool `toml:",omitempty"`
	UseSerializedAuth bool `toml:",omitempty"`
	AuthRealm         string
	UserAgent         string `toml:",omitempty"`

	And []AndConfig

	Response htstat.HttpStatusTbl `toml:",omitempty"`
}

type AndParam struct {
	RewriteAgent *http_auth.RewriteAgent
	HttpResponse htstat.HttpStatusTbl
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
	SystemHttpResponse htstat.HttpStatusTbl
)

var AndParamList []*AndParam
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

	cfg := &RewriteAndConfig{}
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

	cfg.Response.SetDefault()
	if !cfg.Response.IsValid() {
		die("response code config error.")
		return
	}
	SystemHttpResponse = cfg.Response

	AndParamList = make([]*AndParam, 0, len(cfg.And))
	for _, acfg := range cfg.And {
		if acfg.AuthUrl.IsZero() {
			die("[[and]] auth_url is required")
		}

		acfg.Response.SetDefaultByTbl(&SystemHttpResponse)
		if !acfg.Response.IsValid() {
			die("response code config error.")
			return
		}

		if acfg.Timeout < 0 {
			die("bad timeout: %d", acfg.Timeout)
		}
		tout := DEFAULT_TIMEOUT
		if acfg.Timeout > 0 {
			tout = acfg.Timeout
		}

		rw_agt := http_auth.NewRewriteAgent(&http_auth.RewriteAgentConfig{
			UsernameRe:     acfg.UsernameRe.Regexp,
			UserAgent:      UserAgent,
			AuthAddr:       acfg.AuthUrl,
			SetUsernameVal: acfg.SetUsername,
			SetPasswordVal: acfg.SetPassword,
			SetHeadersVal:  acfg.SetHeader,
			SkipCertVerify: acfg.SkipCertVerify,
			RootCaFiles:    acfg.RootCaFiles,
			Timeout:        time.Duration(tout) * time.Millisecond,
		})

		AndParamList = append(AndParamList, &AndParam{
			RewriteAgent: rw_agt,
			HttpResponse: acfg.Response,
		})
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
