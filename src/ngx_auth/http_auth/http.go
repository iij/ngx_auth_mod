package http_auth

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"
)

type HttpAuthConfig struct {
	AuthAddr HtAuthAddr

	Username  string
	Password  string
	SetHeader map[string]string

	SkipCertVerify bool
	RootCaFiles    []string

	Timeout time.Duration
}

type HttpAuthAgent struct {
	HttpAuthConfig
}

const MIN_TIMEOUT = 100 * time.Millisecond

func NewHttpAuthAgent(cfg *HttpAuthConfig) *HttpAuthAgent {
	ha_agt := &HttpAuthAgent{}
	ha_agt.HttpAuthConfig = *cfg

	if ha_agt.Timeout < MIN_TIMEOUT {
		ha_agt.Timeout = MIN_TIMEOUT
	}

	return ha_agt
}

func verify_user(user string) bool {
	return !bytes.ContainsFunc([]byte(user),
		func(b rune) bool {
			if b >= 0 && b <= 0x1f {
				return true
			}
			if b == 0x7f {
				return true
			}
			if b == ':' {
				return true
			}

			return false
		})
}

func verify_pass(pass string) bool {
	return !bytes.ContainsFunc([]byte(pass),
		func(b rune) bool {
			if b >= 0 && b <= 0x1f {
				return true
			}
			if b == 0x7f {
				return true
			}

			return false
		})
}

func (auth_req *HttpAuthAgent) Auth() (int, error) {
	cli, err := new_http_client(auth_req)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	u := &url.URL{}
	*u = *auth_req.AuthAddr.Url
	if auth_req.Username != "" || auth_req.Password != "" {
		if !verify_user(auth_req.Username) || !verify_pass(auth_req.Password) {
			return http.StatusUnauthorized, nil
		}
		u.User = url.UserPassword(auth_req.Username, auth_req.Password)
	}

	ctx, cancel := context.WithTimeout(context.Background(), auth_req.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	for n, v := range auth_req.SetHeader {
		if v != "" {
			req.Header.Set(n, v)
		} else {
			req.Header.Del(n)
		}
	}

	res, err := cli.Do(req)
	if err != nil {
		if os.IsTimeout(err) {
			return http.StatusGatewayTimeout, err
		}
		return http.StatusBadGateway, err
	}
	defer func() {
		io.Copy(io.Discard, res.Body)
	}()

	return res.StatusCode, nil
}

func new_tls_config(cafiles []string, skip_verify bool) (*tls.Config, error) {
	ca_pool := x509.NewCertPool()
	if len(cafiles) > 0 {
		for _, fn := range cafiles {
			ca_pem, e := os.ReadFile(fn)
			if e != nil {
				return nil, e
			}
			ca_pool.AppendCertsFromPEM(ca_pem)
		}
	} else {
		var e error
		ca_pool, e = x509.SystemCertPool()
		if e != nil {
			return nil, e
		}
	}

	tls_cfg := &tls.Config{
		InsecureSkipVerify: skip_verify,
		RootCAs:            ca_pool,
	}

	return tls_cfg, nil
}

func new_transport(cfg *HttpAuthAgent) (http.RoundTripper, error) {
	tls_cfg, err := new_tls_config(cfg.RootCaFiles, cfg.SkipCertVerify)
	if err != nil {
		return nil, err
	}

	half_tout := cfg.Timeout / 2
	dialer := &net.Dialer{
		Timeout:   half_tout,
		KeepAlive: half_tout,
	}

	dial_ctx := dialer.DialContext
	if cfg.AuthAddr.Unix != nil {
		dial_ctx = func(ctx context.Context, _, _ string) (net.Conn, error) {
			return dialer.DialContext(ctx, "unix", cfg.AuthAddr.Unix.Socket)
		}
	}

	trans := http.DefaultTransport.(*http.Transport).Clone()
	trans.Proxy = nil
	trans.DialContext = dial_ctx
	trans.ForceAttemptHTTP2 = true
	trans.MaxIdleConns = 10
	trans.IdleConnTimeout = half_tout
	trans.TLSHandshakeTimeout = half_tout
	trans.ExpectContinueTimeout = 1 * time.Second
	trans.TLSClientConfig = tls_cfg

	return trans, nil
}

func new_http_client(cfg *HttpAuthAgent) (*http.Client, error) {
	tls_cfg, err := new_transport(cfg)
	if err != nil {
		return nil, err
	}

	return &http.Client{Transport: tls_cfg}, nil
}
