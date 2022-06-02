package main

import (
	"io"
	"net/http"
	"strings"
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
