package main

import (
	"crypto/sha512"
	"encoding/binary"
	"fmt"
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

func makeEtag(ms int64, user, pass string, rw *http_auth.RewriteAgent) string {
	tm := make([]byte, 8)
	set_int64bin(tm, ms)

	hash := sha512.New()
	rw.WriteSum(hash)
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

	tag := makeEtag(StartTimeMS, user, pass, RewriteAgent)
	w.Header().Set("Etag", tag)

	if UseEtag {
		if !isModified(header, tag) {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

	ha_agt, err := RewriteAgent.NewHttpAuthAgent(user, pass, header)
	if err != nil {
		SystemHttpResponse.InvalidSetting.Error(w)
		return
	}
	if ha_agt == nil {
		http_not_auth(SystemHttpResponse, w, r)
		return
	}

	if UseSerializedAuth {
		userMtx.Lock(user)
		defer userMtx.Unlock(user)
	}

	rcode, err := ha_agt.Auth()
	if err != nil {
		if os.IsTimeout(err) {
			SystemHttpResponse.Timeout.Error(w)
			return
		}
		SystemHttpResponse.BadGateway.Error(w)
		return
	}

	switch {
	default:
		SystemHttpResponse.BadRequest.Error(w)
		return

	case rcode == http.StatusUnauthorized:
		http_not_auth(SystemHttpResponse, w, r)
		return

	case rcode == http.StatusForbidden:
		SystemHttpResponse.Forbidden.Error(w)
		return

	case rcode >= 200 && rcode < 300:
	}

	if CacheSeconds > 0 {
		w.Header().Set("Cache-Control",
			fmt.Sprintf("max-age=%d, must-revalidate", CacheSeconds))
	}
	SystemHttpResponse.Ok.Error(w)
}
