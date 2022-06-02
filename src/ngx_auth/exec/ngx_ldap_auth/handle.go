package main

import (
	"io"
	"net/http"
	"strings"

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

func TestAuthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Context-Type", "text/plain; charset=utf-8")

	user, pass, ok := r.BasicAuth()
	if !ok {
		http_not_auth(w, r)
		return
	}

	if !auth(user, pass) {
		http_not_auth(w, r)
		return
	}

	io.WriteString(w, "Authorized\n")
}
