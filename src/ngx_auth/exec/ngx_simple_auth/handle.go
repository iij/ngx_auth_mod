package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"net/http"
	"strings"

	"ngx_auth/etag"
)

func auth(user string, pass string) bool {
	pw, ok := Password[user]
	return ok && pw == pass
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

func TestAuthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	user, pass, ok := r.BasicAuth()
	if !ok {
		http_not_auth(w, r)
		return
	}

	tag := makeEtag(StartTimeMS, user, pass)

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
