//go:build e2e

// Package e2e holds the end-to-end tests of the api's HTTP interface. They run against
// the product brought up by `make test-e2e` (docker-compose.yml + docker-compose.test.yml)
// and enter as each actor does: an HTTP client with the token the api issues to that
// user. Email is read from Mailpit; Google is the fake of docker-compose.test.yml.
package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"
)

// Seeded by infra/e2e/seed.sql.
const (
	adminEmail     = "admin.e2e@openparallax.test"
	organizerEmail = "organizer.e2e@openparallax.test"
	seededPassword = "e2e-password-123"
)

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return strings.TrimRight(v, "/")
	}
	return def
}

var (
	apiURL     = env("API_URL", "http://localhost:8080")
	mailpitURL = env("MAILPIT_URL", "http://localhost:8025")
	httpClient = &http.Client{Timeout: 30 * time.Second}
)

// unique makes names and emails that do not collide with earlier runs.
func unique(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}

type response struct {
	Status int
	Body   map[string]any
	Raw    []byte
}

func do(t *testing.T, method, path, token string, body any) response {
	t.Helper()
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		reader = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, apiURL+path, reader)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return send(t, req, token)
}

func send(t *testing.T, req *http.Request, token string) response {
	t.Helper()
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	res, err := httpClient.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", req.Method, req.URL.Path, err)
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	out := response{Status: res.StatusCode, Raw: raw}
	_ = json.Unmarshal(raw, &out.Body)
	return out
}

func expect(t *testing.T, r response, status int, what string) {
	t.Helper()
	if r.Status != status {
		t.Fatalf("%s: status %d, want %d: %s", what, r.Status, status, r.Raw)
	}
}

// field walks a decoded JSON body: field(body, "data", "id").
func field(v any, path ...string) any {
	for _, k := range path {
		m, ok := v.(map[string]any)
		if !ok {
			return nil
		}
		v = m[k]
	}
	return v
}

func str(v any, path ...string) string {
	s, _ := field(v, path...).(string)
	return s
}

type user struct {
	ID, Email, Token string
}

// signUp creates an account with email and password, the way the web does.
func signUp(t *testing.T, name string) user {
	t.Helper()
	email := strings.ToLower(unique(name)) + "@example.com"
	r := do(t, http.MethodPost, "/api/v1/users", "", map[string]any{
		"name": name, "email": email, "password": "e2e-password-123",
	})
	expect(t, r, http.StatusCreated, "sign up "+name)
	return user{ID: str(r.Body, "user", "id"), Email: email, Token: str(r.Body, "token")}
}

func logIn(t *testing.T, email, password string) user {
	t.Helper()
	r := do(t, http.MethodPost, "/api/v1/users/authenticate", "", map[string]any{
		"email": email, "password": password,
	})
	expect(t, r, http.StatusOK, "log in "+email)
	return user{ID: str(r.Body, "user", "id"), Email: email, Token: str(r.Body, "token")}
}

func date(days int) string {
	return time.Now().AddDate(0, 0, days).Format("2006-01-02")
}

// createEvent creates an event in the creation stage, as the web's wizard does.
func createEvent(t *testing.T, organizer user) string {
	t.Helper()
	r := do(t, http.MethodPost, "/api/v1/events", organizer.Token, map[string]any{
		"name":             unique("Evento e2e"),
		"description":      "Evento creado por los tests end to end.",
		"start_date":       date(1),
		"end_date":         date(2),
		"max_participants": 10,
	})
	expect(t, r, http.StatusCreated, "create event")
	id := str(r.Body, "event", "id")
	if id == "" {
		t.Fatalf("create event: no id in %s", r.Raw)
	}
	return id
}

func upload(t *testing.T, eventID string, participant user, filename, content string) string {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, err := w.CreatePart(map[string][]string{
		"Content-Disposition": {fmt.Sprintf(`form-data; name="file"; filename=%q`, filename)},
		"Content-Type":        {"text/plain"},
	})
	if err != nil {
		t.Fatal(err)
	}
	_, _ = part.Write([]byte(content))
	_ = w.WriteField("description", "Propuesta de "+participant.Email)
	_ = w.Close()
	req, _ := http.NewRequest(http.MethodPost,
		fmt.Sprintf("%s/api/v1/events/%s/participant/%s/attachment", apiURL, eventID, participant.ID), &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	r := send(t, req, participant.Token)
	expect(t, r, http.StatusCreated, "upload proposal of "+participant.Email)
	return str(r.Body, "data", "id")
}

// mailTo waits for an email to the address whose subject contains subject, and
// returns its plain-text body.
func mailTo(t *testing.T, to, subject string) string {
	t.Helper()
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		q := url.QueryEscape(fmt.Sprintf("to:%q", to))
		res, err := httpClient.Get(mailpitURL + "/api/v1/search?query=" + q)
		if err == nil {
			var list struct {
				Messages []struct {
					ID      string `json:"ID"`
					Subject string `json:"Subject"`
				} `json:"messages"`
			}
			_ = json.NewDecoder(res.Body).Decode(&list)
			res.Body.Close()
			for _, m := range list.Messages {
				if strings.Contains(m.Subject, subject) {
					msg, err := httpClient.Get(mailpitURL + "/api/v1/message/" + m.ID)
					if err != nil {
						t.Fatalf("mailpit: %v", err)
					}
					var body struct {
						Text string `json:"Text"`
					}
					_ = json.NewDecoder(msg.Body).Decode(&body)
					msg.Body.Close()
					return body.Text
				}
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	t.Fatalf("no email to %s with subject %q in Mailpit", to, subject)
	return ""
}

var resetToken = regexp.MustCompile(`reset-password\?token=([0-9a-f]+)`)
