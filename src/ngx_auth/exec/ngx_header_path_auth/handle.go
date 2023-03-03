package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"net/http"

	"ngx_auth/etag"
)

func get_path_right(rpath string, user string) bool {
	pathid, ok := check_path(rpath)
	if !ok {
		return UserMap.Authz(NomatchRight, user)
	}

	right_type, has := PathRight[pathid]
	if !has {
		return UserMap.Authz(DefaultRight, user)
	}

	return UserMap.Authz(right_type, user)
}

func check_path(rpath string) (string, bool) {
	if PathPatternReg == nil {
		return "", false
	}
	matchs := PathPatternReg.FindStringSubmatch(rpath)
	if len(matchs) < 1 {
		return "", false
	}
	return matchs[1], true
}

func set_int64bin(bin []byte, v int64) {
	binary.LittleEndian.PutUint64(bin, uint64(v))
}

func makeEtag(ms int64, user, rpath string) string {
	pathid, ok := check_path(rpath)
	if ok {
		pathid = "M" + pathid
	} else {
		pathid = "N"
	}

	tm := make([]byte, 8)
	set_int64bin(tm, ms)

	return etag.Make(tm, etag.Crypt(tm, []byte(user)), []byte(pathid))
}

func TestAuthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	rpath := r.Header.Get(PathHeader)
	if rpath == "" {
		http.Error(w, "No path header", http.StatusForbidden)
		return
	}

	user := r.Header.Get(UserHeader)
	if user == "" {
		http.Error(w, "No user header", http.StatusForbidden)
		return
	}
	tag := makeEtag(StartTimeMS, user, rpath)

	if !get_path_right(rpath, user) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if CacheSeconds > 0 {
		w.Header().Set("Cache-Control",
			fmt.Sprintf("max-age=%d, must-revalidate", CacheSeconds))
	}
	w.Header().Set("Etag", tag)
	io.WriteString(w, "Authorized\n")
}
