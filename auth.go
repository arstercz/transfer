// ref https://gist.github.com/elithrar/9146306
package main

import (
	"encoding/base64"
	"log"
	"net/http"
	"strings"
)

func auth(h http.HandlerFunc, middleware ...func(http.HandlerFunc) http.HandlerFunc) http.HandlerFunc {
	for _, m := range middleware {
		h = m(h)
	}
	return h
}

//https://developer.mozilla.org/en-US/docs/Web/HTTP/Authentication
func basicAuth(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
		s := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
		if len(s) != 2 {
			http.Error(w, "Not authorized", 401)
			return
		}

		b, err := base64.StdEncoding.DecodeString(s[1])
		if err != nil {
			http.Error(w, err.Error(), 401)
			return
		}

		pair := strings.SplitN(string(b), ":", 2)
		if len(pair) != 2 {
			http.Error(w, "Not authorized", 401)
			return
		}

		if !VerifyOk(pair[0], pair[1]) {
			log.Printf("Verify user error, username: %s, pass: %s", pair[0], pair[1])
			http.Error(w, "Authorized error!", 401)
			return
		}
		h.ServeHTTP(w, r)
	}
}
