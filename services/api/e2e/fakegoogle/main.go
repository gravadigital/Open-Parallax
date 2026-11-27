// Command fakegoogle replaces Google's tokeninfo and userinfo endpoints in the
// end-to-end tests (docker-compose.test.yml), so the Google sign-in can be
// exercised with no real Google account.
//
// An access token is "e2e:{sub}:{email}:{name}". Anything else is rejected the
// way Google rejects an invalid token. The audience it reports is FAKE_GOOGLE_CLIENT_ID,
// which must equal the api's GOOGLE_CLIENT_ID.
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
)

type profile struct {
	sub, email, name string
}

func parse(token string) (profile, bool) {
	parts := strings.SplitN(token, ":", 4)
	if len(parts) != 4 || parts[0] != "e2e" || parts[1] == "" || parts[2] == "" {
		return profile{}, false
	}
	return profile{sub: parts[1], email: parts[2], name: parts[3]}, true
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func main() {
	clientID := os.Getenv("FAKE_GOOGLE_CLIENT_ID")
	addr := os.Getenv("FAKE_GOOGLE_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /tokeninfo", func(w http.ResponseWriter, r *http.Request) {
		p, ok := parse(r.URL.Query().Get("access_token"))
		if !ok {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_token"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{
			"aud": clientID, "sub": p.sub, "email": p.email, "email_verified": "true",
		})
	})
	mux.HandleFunc("GET /userinfo", func(w http.ResponseWriter, r *http.Request) {
		p, ok := parse(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid_token"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"sub": p.sub, "email": p.email, "name": p.name})
	})
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "up"})
	})

	log.Printf("fake google listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
