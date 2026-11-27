//go:build e2e

package e2e

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"testing"
)

// A participant has at most one proposal per event: a second upload, by the participant or by the
// organizer on their behalf, and the losing one of two simultaneous uploads, answer 409 with the
// existing proposal, whose file stays as it was.
func TestOneProposalPerParticipant(t *testing.T) {
	t.Run("[INS-05][INS-E15][INV-13][GAR-16] a participant cannot have a second proposal in the event, not even with two simultaneous uploads", testOneProposalPerParticipant)
}

func testOneProposalPerParticipant(t *testing.T) {
	organizer := signUp(t, "Organizadora")
	eventID := createEvent(t, organizer)
	r := do(t, http.MethodPatch, "/api/v1/events/"+eventID+"/stage", organizer.Token, map[string]any{
		"stage": "participation", "estimated_end_date": date(7),
	})
	expect(t, r, http.StatusOK, "open registration")

	register := func(name string) user {
		p := signUp(t, name)
		r := do(t, http.MethodPost, "/api/v1/events/"+eventID+"/register", "", map[string]any{
			"participant_name": name, "participant_email": p.Email,
		})
		expect(t, r, http.StatusCreated, "register "+p.Email)
		return p
	}

	// A second upload after the first one.
	first := register("ParticipanteUno")
	existing := upload(t, eventID, first, "primera.txt", "Primera propuesta")

	again := send(t, uploadRequest(t, eventID, first, "segunda.txt", "Segunda propuesta"), first.Token)
	expect(t, again, http.StatusConflict, "second upload")
	if code := str(again.Body, "code"); code != "DUPLICATE_ATTACHMENT" {
		t.Fatalf("second upload: code %q, want DUPLICATE_ATTACHMENT", code)
	}
	if id := str(again.Body, "existing_attachment_id"); id != existing {
		t.Fatalf("second upload: existing_attachment_id %q, want %q", id, existing)
	}

	// The organizer uploading on the participant's behalf (INS-06) meets the same rule.
	onBehalf := send(t, uploadRequest(t, eventID, first, "en-nombre.txt", "En nombre"), organizer.Token)
	expect(t, onBehalf, http.StatusConflict, "upload on behalf of a participant who has one")

	expectProposals(t, eventID, organizer, map[string]string{existing: "Primera propuesta"})

	// Two simultaneous uploads of another participant: exactly one is saved.
	second := register("ParticipanteDos")
	const uploads = 4
	type result struct {
		status int
		body   map[string]any
	}
	results := make([]result, uploads)
	var start, done sync.WaitGroup
	start.Add(1)
	for i := 0; i < uploads; i++ {
		req := uploadRequest(t, eventID, second, fmt.Sprintf("simultanea-%d.txt", i), fmt.Sprintf("Simultánea %d", i))
		req.Header.Set("Authorization", "Bearer "+second.Token)
		done.Add(1)
		go func(i int, req *http.Request) {
			defer done.Done()
			start.Wait()
			res, err := httpClient.Do(req)
			if err != nil {
				results[i] = result{status: -1}
				return
			}
			defer res.Body.Close()
			raw, _ := io.ReadAll(res.Body)
			results[i].status = res.StatusCode
			_ = json.Unmarshal(raw, &results[i].body)
		}(i, req)
	}
	start.Done()
	done.Wait()

	var accepted, acceptedContent string
	var rejected []result
	for i, r := range results {
		switch r.status {
		case http.StatusCreated:
			if accepted != "" {
				t.Fatalf("two simultaneous uploads were saved: %v", results)
			}
			accepted, acceptedContent = str(r.body, "data", "id"), fmt.Sprintf("Simultánea %d", i)
		case http.StatusConflict:
			rejected = append(rejected, r)
		default:
			t.Fatalf("simultaneous upload %d: status %d: %v", i, r.status, r.body)
		}
	}
	if accepted == "" || len(rejected) != uploads-1 {
		t.Fatalf("simultaneous uploads: want one 201 and %d 409, got %v", uploads-1, results)
	}
	for _, r := range rejected {
		if str(r.body, "code") != "DUPLICATE_ATTACHMENT" || str(r.body, "existing_attachment_id") != accepted {
			t.Fatalf("rejected simultaneous upload: %v, want DUPLICATE_ATTACHMENT with %s", r.body, accepted)
		}
	}

	expectProposals(t, eventID, organizer, map[string]string{
		existing: "Primera propuesta",
		accepted: acceptedContent,
	})
}

// expectProposals checks, as the organizer, that the event has exactly these proposals and that
// each one downloads with its content: the file of the accepted upload was not touched.
func expectProposals(t *testing.T, eventID string, organizer user, want map[string]string) {
	t.Helper()
	r := do(t, http.MethodGet, "/api/v1/events/"+eventID+"/attachments", organizer.Token, nil)
	expect(t, r, http.StatusOK, "list proposals")
	list, _ := field(r.Body, "data").([]any)
	if len(list) != len(want) {
		t.Fatalf("event has %d proposals, want %d: %s", len(list), len(want), r.Raw)
	}
	for id, content := range want {
		dl := do(t, http.MethodGet, "/api/v1/attachments/"+id+"/download", organizer.Token, nil)
		expect(t, dl, http.StatusOK, "download "+id)
		if string(dl.Raw) != content {
			t.Fatalf("proposal %s downloads %q, want %q", id, dl.Raw, content)
		}
	}
}
