package main

import (
	"io"
	"net/http"
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

func TestAuthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Context-Type", "text/plain; charset=utf-8")

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

	if !get_path_right(rpath, user) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	io.WriteString(w, "Authorized\n")
}
