package main

import (
	"io"
	"net/http"
	"strings"

	"ngx_auth/ldap_auth"
)

func get_path_filter(rpath string) (bool, string) {
	pathid, ok := check_path(rpath)
	if !ok {
		if BanNomatch {
			return false, ""
		}
		return true, NomatchFilter
	}

	filter, has := PathFilter[pathid]
	if has {
		return true, filter
	}
	if BanDefault {
		return false, ""
	}
	return true, DefaultFilter
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

func http_not_auth(w http.ResponseWriter, r *http.Request) {
	realm := strings.Replace(AuthRealm, `"`, `\"`, -1)
	w.Header().Add("WWW-Authenticate", `Basic realm="`+realm+`"`)
	http.Error(w, "Not authorized", http.StatusUnauthorized)
}

func auth_path(user string, pass string, path string) (bool, bool) {
	ldap_cfg := *LdapAuthConfig
	ok_path, path_filter := get_path_filter(path)
	if !ok_path {
		path_filter = ""
	}

	ldap_cfg.AuthzFilter = path_filter
	la, lerr := ldap_auth.NewLdapAuth(&ldap_cfg)
	if lerr != nil {
		return false, false
	}
	defer la.Close()

	ok_auth, ok_authz, err := la.Authenticate(user, pass)
	if err != nil {
		return false, false
	}
	if !ok_path {
		ok_authz = false
	}

	return ok_auth, ok_authz
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
		return
	}

	io.WriteString(w, "Authorized\n")
}
