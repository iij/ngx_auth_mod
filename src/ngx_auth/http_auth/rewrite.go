package http_auth

import (
	"errors"
	"io"
	"maps"
	"net/http"
	"regexp"
	"slices"
	"time"

	"ngx_auth/htval"
)

var ErrInvalidRewriteFormat = errors.New("Invalid rewrite format")

type RewriteAgentConfig struct {
	UsernameRe *regexp.Regexp

	UserAgent      string
	AuthAddr       HtAuthAddr
	SetUsernameVal htval.HtValue
	SetPasswordVal htval.HtValue
	SetHeadersVal  htval.HtSetHeader

	SkipCertVerify bool
	RootCaFiles    []string

	Timeout time.Duration
}

type RewriteAgent struct {
	RewriteAgentConfig
}

const DEFAULT_TIMEOUT = 30 * time.Second

func NewRewriteAgent(cfg *RewriteAgentConfig) *RewriteAgent {
	rw_agt := &RewriteAgent{}
	rw_agt.RewriteAgentConfig = *cfg

	if rw_agt.Timeout <= 0 {
		rw_agt.Timeout = DEFAULT_TIMEOUT
	}

	return rw_agt
}

func (ra *RewriteAgent) sum_bytes(yield func([]byte) bool) {
	hkey := slices.Sorted(maps.Keys(ra.SetHeadersVal))
	if !yield([]byte(ra.SetUsernameVal)) {
		return
	}
	if !yield([]byte{'\n'}) {
		return
	}
	if !yield([]byte(ra.SetPasswordVal)) {
		return
	}
	if !yield([]byte{'\n'}) {
		return
	}
	for _, k := range hkey {
		if !yield([]byte(ra.SetHeadersVal[k])) {
			return
		}
		if !yield([]byte{'\n'}) {
			return
		}
	}
	if !yield([]byte{0}) {
		return
	}
}

func (ra *RewriteAgent) WriteSum(w io.Writer) (int, error) {
	var nums int = 0
	for b := range ra.sum_bytes {
		n, err := w.Write(b)
		nums += n
		if err != nil {
			return nums, err
		}
	}

	return nums, nil
}

func (ra *RewriteAgent) Match(src_user string) bool {
	if ra.UsernameRe == nil {
		return true
	}

	return ra.UsernameRe.MatchString(src_user)
}

func (ra *RewriteAgent) NewHttpAuthAgent(
	src_user, src_pass string, src_head http.Header) (*HttpAuthAgent, error) {

	var matchs []string = nil
	if ra.UsernameRe != nil {
		m := ra.UsernameRe.FindStringSubmatch(src_user)
		if m == nil {
			return nil, nil
		}

		matchs = m
	}

	ht_args := htval.NewHtValueArgs(src_user, src_pass, src_head, matchs)

	dst_user := ra.SetUsernameVal.ToString(ht_args)
	dst_pass := ra.SetPasswordVal.ToString(ht_args)
	set_head := map[string]string{}

	set_head["User-Agent"] = ra.UserAgent
	set_head["Cache-Control"] = "no-cache"
	for h, vf := range ra.SetHeadersVal {
		set_head[http.CanonicalHeaderKey(string(h))] = vf.ToString(ht_args)
	}

	ha_agt := NewHttpAuthAgent(&HttpAuthConfig{
		AuthAddr:  ra.AuthAddr,
		Username:  dst_user,
		Password:  dst_pass,
		SetHeader: set_head,

		SkipCertVerify: ra.SkipCertVerify,
		RootCaFiles:    ra.RootCaFiles,

		Timeout: ra.Timeout,
	})

	return ha_agt, nil
}
