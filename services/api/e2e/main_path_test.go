//go:build e2e

package e2e

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
)

// The main path of the product through the HTTP interface: an organizer creates an
// event and opens registration, three participants sign up, register and upload a
// proposal (stored in Garage), the organizer opens voting, each participant downloads
// what was assigned and sends their ranking, and the organizer publishes the results.
// The stage changes reach the participants by email (Mailpit).
func TestEventLifecycleThroughHTTP(t *testing.T) {
	organizer := signUp(t, "Organizadora")
	eventID := createEvent(t, organizer)

	r := do(t, http.MethodPatch, "/api/v1/events/"+eventID+"/stage", organizer.Token, map[string]any{
		"stage": "participation", "estimated_end_date": date(7),
	})
	expect(t, r, http.StatusOK, "open registration")

	participants := make([]user, 3)
	for i := range participants {
		p := signUp(t, fmt.Sprintf("Participante%d", i+1))
		r := do(t, http.MethodPost, "/api/v1/events/"+eventID+"/register", "", map[string]any{
			"participant_name": fmt.Sprintf("Participante %d", i+1), "participant_email": p.Email,
		})
		expect(t, r, http.StatusCreated, "register "+p.Email)
		if got := str(r.Body, "data", "participant_id"); got != p.ID {
			t.Fatalf("register %s: participant_id %q, want the existing user %q", p.Email, got, p.ID)
		}
		upload(t, eventID, p, fmt.Sprintf("propuesta-%d.txt", i+1), fmt.Sprintf("Propuesta número %d", i+1))
		participants[i] = p
	}

	r = do(t, http.MethodGet, "/api/v1/events/"+eventID+"/voting-config/preview", organizer.Token, nil)
	expect(t, r, http.StatusOK, "voting preview")
	if field(r.Body, "data", "can_open_voting") != true {
		t.Fatalf("voting preview: cannot open voting: %s", r.Raw)
	}
	m := field(r.Body, "data", "recommended_m")

	r = do(t, http.MethodPatch, "/api/v1/events/"+eventID+"/stage", organizer.Token, map[string]any{
		"stage": "voting", "estimated_end_date": date(14),
		"voting_config": map[string]any{"attachments_per_evaluator": m},
	})
	expect(t, r, http.StatusOK, "open voting")

	for _, p := range participants {
		body := mailTo(t, p.Email, "avanzó a la etapa Votación")
		if !strings.Contains(body, "Votación") {
			t.Fatalf("stage email to %s does not name the stage: %q", p.Email, body)
		}

		r := do(t, http.MethodGet, "/api/v1/events/"+eventID+"/participants/"+p.ID+"/assignment", p.Token, nil)
		expect(t, r, http.StatusOK, "assignment of "+p.Email)
		assignmentID := str(r.Body, "assignment", "id")
		attachments, _ := field(r.Body, "assignment", "attachments").([]any)
		if len(attachments) == 0 {
			t.Fatalf("assignment of %s is empty: %s", p.Email, r.Raw)
		}

		rankings := make([]map[string]any, 0, len(attachments))
		for i, a := range attachments {
			id := str(a, "id")
			dl := do(t, http.MethodGet, "/api/v1/attachments/"+id+"/download", p.Token, nil)
			expect(t, dl, http.StatusOK, "download assigned proposal")
			if !strings.HasPrefix(string(dl.Raw), "Propuesta número") {
				t.Fatalf("downloaded proposal has unexpected content: %q", dl.Raw)
			}
			rankings = append(rankings, map[string]any{"attachment_id": id, "rank": i + 1})
		}

		r = do(t, http.MethodPost, "/api/v1/events/"+eventID+"/participants/"+p.ID+"/ranking-votes", p.Token,
			map[string]any{"assignment_id": assignmentID, "rankings": rankings})
		expect(t, r, http.StatusCreated, "ranking of "+p.Email)
	}

	r = do(t, http.MethodPatch, "/api/v1/events/"+eventID+"/stage", organizer.Token, map[string]any{"stage": "results"})
	expect(t, r, http.StatusOK, "publish results")

	r = do(t, http.MethodGet, "/api/v1/events/"+eventID+"/distributed-results", "", nil)
	expect(t, r, http.StatusOK, "public results")
	ranking, _ := field(r.Body, "data", "adjusted_ranking").([]any)
	if len(ranking) != len(participants) {
		t.Fatalf("results rank %d proposals, want %d: %s", len(ranking), len(participants), r.Raw)
	}
}
