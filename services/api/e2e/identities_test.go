//go:build e2e

package e2e

import (
	"net/http"
	"strings"
	"testing"
)

// The admin (seeded by infra/e2e/seed.sql) logs in with email and password and, with
// that token, sees an event still in creation that nobody else can see.
func TestAdminIdentitySeesEventsInCreation(t *testing.T) {
	t.Run("[EVE-13][EXP-02] the admin sees an event still in creation that others cannot", testAdminIdentitySeesEventsInCreation)
}

func testAdminIdentitySeesEventsInCreation(t *testing.T) {
	organizer := signUp(t, "Organizador")
	eventID := createEvent(t, organizer)
	other := signUp(t, "Curioso")

	r := do(t, http.MethodGet, "/api/v1/events/"+eventID, other.Token, nil)
	expect(t, r, http.StatusNotFound, "event in creation seen by another user")

	admin := logIn(t, adminEmail, seededPassword)
	r = do(t, http.MethodGet, "/api/v1/events/"+eventID, admin.Token, nil)
	expect(t, r, http.StatusOK, "event in creation seen by the admin")
}

// A user with the global role organizer (seeded) gets through the organizer check on
// an event that is not theirs: the request reaches the stage rule, not a 403.
func TestOrganizerRoleIdentityActsOnAnyEvent(t *testing.T) {
	t.Run("[VOT-13] a global organizer gets through the permission check of any event", testOrganizerRoleIdentityActsOnAnyEvent)
}

func testOrganizerRoleIdentityActsOnAnyEvent(t *testing.T) {
	author := signUp(t, "Autora")
	eventID := createEvent(t, author)

	organizer := logIn(t, organizerEmail, seededPassword)
	r := do(t, http.MethodPost, "/api/v1/events/"+eventID+"/voting-config", organizer.Token, map[string]any{
		"attachments_per_evaluator": 2,
	})
	if r.Status == http.StatusForbidden || r.Status == http.StatusUnauthorized {
		t.Fatalf("organizer role rejected by the permission check: %d %s", r.Status, r.Raw)
	}
	if str(r.Body, "code") != "INVALID_EVENT_STAGE" {
		t.Fatalf("organizer role: want the stage rule (INVALID_EVENT_STAGE), got %d %s", r.Status, r.Raw)
	}
}

// Sign-in with Google, against the fake Google of docker-compose.test.yml: a new Google
// account chooses its name and gets a session; the next time it is an existing user.
func TestGoogleSignIn(t *testing.T) {
	t.Run("[ACC-03][ACC-07][ACC-E9] sign-in with Google creates the account and then logs in", testGoogleSignIn)
}

func testGoogleSignIn(t *testing.T) {
	sub := unique("google-sub")
	email := strings.ToLower(unique("google")) + "@example.com"
	token := "e2e:" + sub + ":" + email + ":Usuaria Google"

	r := do(t, http.MethodPost, "/api/v1/auth/google/verify", "", map[string]any{"token": token})
	expect(t, r, http.StatusOK, "google verify (new user)")
	if str(r.Body, "status") != "new_user" {
		t.Fatalf("google verify: want new_user, got %s", r.Raw)
	}

	r = do(t, http.MethodPost, "/api/v1/auth/google/register", "", map[string]any{
		"token": token, "username": unique("Usuaria Google"),
	})
	expect(t, r, http.StatusCreated, "google register")
	if str(r.Body, "token") == "" {
		t.Fatalf("google register: no session token: %s", r.Raw)
	}

	r = do(t, http.MethodPost, "/api/v1/auth/google/verify", "", map[string]any{"token": token})
	expect(t, r, http.StatusOK, "google verify (existing user)")
	if str(r.Body, "status") != "existing_user" || str(r.Body, "token") == "" {
		t.Fatalf("google verify: want existing_user with a token, got %s", r.Raw)
	}

	r = do(t, http.MethodPost, "/api/v1/auth/google/verify", "", map[string]any{"token": "not-a-google-token"})
	expect(t, r, http.StatusUnauthorized, "google verify with an invalid token")
}

// Password recovery through the email (Mailpit): the link in the email sets a new
// password, and the user logs in with it.
func TestPasswordRecoveryByEmail(t *testing.T) {
	t.Run("[ACC-05][ACC-06][NOT-10] password recovery through the emailed link", testPasswordRecoveryByEmail)
}

func testPasswordRecoveryByEmail(t *testing.T) {
	u := signUp(t, "Olvidadiza")

	r := do(t, http.MethodPost, "/api/v1/users/forgot-password", "", map[string]any{"email": u.Email})
	expect(t, r, http.StatusOK, "forgot password")

	body := mailTo(t, u.Email, "Recuperación de contraseña")
	m := resetToken.FindStringSubmatch(body)
	if m == nil {
		t.Fatalf("no reset link in the email: %q", body)
	}

	r = do(t, http.MethodPost, "/api/v1/users/reset-password", "", map[string]any{
		"token": m[1], "password": "nueva-clave-456",
	})
	expect(t, r, http.StatusOK, "reset password")
	logIn(t, u.Email, "nueva-clave-456")
}
