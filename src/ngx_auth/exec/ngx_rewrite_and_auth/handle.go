package main

import (
	"crypto/sha512"
	"encoding/binary"
	"fmt"
	"iter"
	"net/http"
	"os"
	"strings"

	"github.com/l4go/var_mtx"

	"ngx_auth/etag"
	"ngx_auth/htstat"
	"ngx_auth/http_auth"
)

var userMtx = var_mtx.NewVarMutex()

func http_not_auth(ht_res htstat.HttpStatusTbl, w http.ResponseWriter, _ *http.Request) {
	realm := strings.ReplaceAll(AuthRealm, `"`, `\"`)
	w.Header().Add("WWW-Authenticate", `Basic realm="`+realm+`"`)
	ht_res.Unauth.Error(w)
}

func set_int64bin(bin []byte, v int64) {
	binary.LittleEndian.PutUint64(bin, uint64(v))
}

func makeEtag(ms int64, user, pass string, rw_it iter.Seq[*http_auth.RewriteAgent]) string {
	tm := make([]byte, 8)
	set_int64bin(tm, ms)

	hash := sha512.New()
	for rw := range rw_it {
		rw.WriteSum(hash)
	}
	hd_sum := hash.Sum(nil)

	return etag.Make(tm,
		etag.Crypt(tm, []byte(user)),
		etag.Hmac([]byte(user), []byte(pass)),
		hd_sum)
}

func isModified(hd http.Header, org_tag string) bool {
	if_nmatch := hd.Get("If-None-Match")

	if if_nmatch != "" {
		return !isEtagMatch(if_nmatch, org_tag)
	}

	return true
}

func isEtagMatch(tag_str string, org_tag string) bool {
	tags, _ := etag.Split(tag_str)
	for _, tag := range tags {
		if tag == org_tag {
			return true
		}
	}

	return false
}

func TestAuthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")

	user, pass, ok := r.BasicAuth()
	if !ok {
		user = ""
		pass = ""
	}
	header := r.Header

	if NegCacheSeconds > 0 {
		w.Header().Set("Cache-Control",
			fmt.Sprintf("max-age=%d, must-revalidate", NegCacheSeconds))
	}

	rw_iter := func(yield func(*http_auth.RewriteAgent) bool) {
		for _, and_p := range AndParamList {
			if !yield(and_p.RewriteAgent) {
				return
			}
		}
	}
	tag := makeEtag(StartTimeMS, user, pass, rw_iter)
	w.Header().Set("Etag", tag)

	if UseEtag {
		if !isModified(header, tag) {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

	for _, and_p := range AndParamList {
		if !and_p.RewriteAgent.Match(user) {
			http_not_auth(and_p.HttpResponse, w, r)
			return
		}
	}

	if UseSerializedAuth {
		userMtx.Lock(user)
		defer userMtx.Unlock(user)
	}

	for _, and_p := range AndParamList {
		ha_agt, e := and_p.RewriteAgent.NewHttpAuthAgent(user, pass, header)
		if e != nil {
			and_p.HttpResponse.InvalidSetting.Error(w)
			return
		}
		if ha_agt == nil {
			http_not_auth(and_p.HttpResponse, w, r)
			return
		}

		rcode, err := ha_agt.Auth()
		if err != nil {
			if os.IsTimeout(err) {
				and_p.HttpResponse.Timeout.Error(w)
				return
			}
			and_p.HttpResponse.BadGateway.Error(w)
			return
		}

		switch {
		default:
			and_p.HttpResponse.BadRequest.Error(w)
			return

		case rcode == http.StatusUnauthorized:
			http_not_auth(and_p.HttpResponse, w, r)
			return

		case rcode == http.StatusForbidden:
			and_p.HttpResponse.Forbidden.Error(w)
			return

		case rcode >= 200 && rcode < 300:
		}
	}

	if CacheSeconds > 0 {
		w.Header().Set("Cache-Control",
			fmt.Sprintf("max-age=%d, must-revalidate", CacheSeconds))
	}

	SystemHttpResponse.Ok.Error(w)
}
