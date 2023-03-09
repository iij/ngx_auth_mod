package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"net/http"
	"strings"

	"ngx_auth/etag"
	"ngx_auth/ldap_auth"
)

func auth(user string, pass string) bool {
	la, err := ldap_auth.NewLdapAuth(LdapAuthConfig)
	if err != nil {
		return false
	}
	defer la.Close()

	ok_auth, _, err := la.Authenticate(user, pass)
	if err != nil {
		return false
	}

	return ok_auth
}

func http_not_auth(w http.ResponseWriter, r *http.Request) {
	realm := strings.Replace(AuthRealm, `"`, `\"`, -1)
	w.Header().Add("WWW-Authenticate", `Basic realm="`+realm+`"`)
	http.Error(w, "Not authorized", http.StatusUnauthorized)
}

func set_int64bin(bin []byte, v int64) {
	binary.LittleEndian.PutUint64(bin, uint64(v))
}

func makeEtag(ms int64, user, pass string) string {
	tm := make([]byte, 8)
	set_int64bin(tm, ms)

	return etag.Make(tm, etag.Crypt(tm, []byte(user)),
		etag.Hmac([]byte(user), []byte(pass)))
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
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	user, pass, ok := r.BasicAuth()
	if !ok {
		http_not_auth(w, r)
		return
	}

	tag := makeEtag(StartTimeMS, user, pass)
	if UseEtag {
		if !isModified(r.Header, tag) {
			if CacheSeconds > 0 {
				w.Header().Set("Cache-Control",
					fmt.Sprintf("max-age=%d, must-revalidate", CacheSeconds))
			}
			w.Header().Set("Etag", tag)
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

	if !auth(user, pass) {
		http_not_auth(w, r)
		return
	}

	if CacheSeconds > 0 {
		w.Header().Set("Cache-Control",
			fmt.Sprintf("max-age=%d, must-revalidate", CacheSeconds))
	}
	w.Header().Set("Etag", tag)
	io.WriteString(w, "Authorized\n")
}
