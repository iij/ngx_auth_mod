package main

import (
	"io"
	"net/http"
	"strings"

	"ngx_auth/ldap_auth"
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

func auth_path(user string, pass string, rpath string) (bool, bool) {
	la, err := ldap_auth.NewLdapAuth(LdapAuthConfig)
	if err != nil {
		return false, false
	}
	defer la.Close()

	ok_auth, ok_authz, err := la.Authenticate(user, pass)
	if err != nil {
		return false, false
	}
	if !ok_auth || !ok_authz {
		return ok_auth, ok_authz
	}

	if !get_path_right(rpath, user) {
		return true, false
	}

	return true, true
}

func http_not_auth(w http.ResponseWriter, r *http.Request) {
	realm := strings.Replace(AuthRealm, `"`, `\"`, -1)
	w.Header().Add("WWW-Authenticate", `Basic realm="`+realm+`"`)
	http.Error(w, "Not authorized", http.StatusUnauthorized)
}

func TestAuthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Context-Type", "text/plain; charset=utf-8")

	rpath := r.Header.Get(PathHeader)
	if rpath == "" {
		http.Error(w, "No path header", http.StatusForbidden)
		return
	}

	user, pass, ok := r.BasicAuth()
	if !ok {
		http_not_auth(w, r)
		return
	}

	ok_auth, ok_authz := auth_path(user, pass, rpath)
	if !ok_auth {
		http_not_auth(w, r)
		return
	}
	if !ok_authz {
		http.Error(w, "Forbidden", http.StatusForbidden)
	}

	io.WriteString(w, "Authorized\n")
}
