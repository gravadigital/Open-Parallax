//go:build integration
// +build integration

package handlers

import (
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gravadigital/openparallax/api/internal/config"
	"github.com/gravadigital/openparallax/api/internal/domain/notification"
	"github.com/gravadigital/openparallax/api/internal/domain/participant"
	"github.com/gravadigital/openparallax/api/internal/email"
	"github.com/gravadigital/openparallax/api/internal/storage/migrations"
	"github.com/gravadigital/openparallax/api/internal/storage/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// migratedTx opens the test database, applies the migrations and returns a transaction that is
// rolled back when the test ends.
func migratedTx(t *testing.T) *gorm.DB {
	t.Helper()
	cfg := config.Load()
	if testDB := os.Getenv("TEST_DB_NAME"); testDB != "" {
		cfg.DB.Name = testDB
	}
	db, err := postgres.Connect(cfg)
	require.NoError(t, err)
	require.NoError(t, postgres.AutoMigrate(db, nil))
	tx := db.Begin()
	t.Cleanup(func() {
		tx.Rollback()
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
	})
	return tx
}

// serializedParticipants guards the participants read with mu: the stage-change email goroutine
// makes it on the test transaction, which takes one query at a time, so the test holds mu while it
// queries.
type serializedParticipants struct {
	postgres.UserRepository
	mu *sync.Mutex
}

func (r *serializedParticipants) GetEventParticipants(eventID string) ([]*participant.UserWithEventRole, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.UserRepository.GetEventParticipants(eventID)
}

// eventHandlerOnDB builds the EventHandler over the real repositories of tx; email is disabled and
// notifications go to the in-memory repository.
func eventHandlerOnDB(tx *gorm.DB, users *serializedParticipants) *testEventHandlerSet {
	setupRepo := postgres.NewPostgresVotingSetupRepository(tx)
	handler := NewEventHandler(
		postgres.NewPostgresEventRepository(tx),
		users,
		postgres.NewPostgresAttachmentRepository(tx),
		postgres.NewPostgresVoteRepository(tx),
		setupRepo,
		email.NewEmailService(&config.Config{}), nil, &config.Config{},
		notification.NewService(newMockNotificationRepository()),
	)
	return &testEventHandlerSet{handler: handler}
}

func migration029(t *testing.T) migrations.Migration {
	t.Helper()
	for _, m := range migrations.GetMigrations() {
		if m.ID == "029" {
			return m
		}
	}
	t.Fatal("migration 029 is not registered")
	return migrations.Migration{}
}

func TestOpenVotingAfterDuplicateCleanup(t *testing.T) {
	t.Run("[662-AC-02] an event where a participant had two proposals opens voting after the deployment, and that participant gets a single assignment", func(t *testing.T) {
		// Registered before the transaction, so it runs after the rollback: a late email goroutine
		// then fails on the closed transaction instead of racing the test's queries.
		var mu sync.Mutex
		holding := false
		t.Cleanup(func() {
			if holding {
				mu.Unlock()
			}
		})
		tx := migratedTx(t)
		m := migration029(t)
		require.NoError(t, m.Down(tx)) // a database from before 029, where duplicates were accepted

		eventID, author := uuid.New(), uuid.New()
		participants := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
		for _, id := range append([]uuid.UUID{author}, participants...) {
			require.NoError(t, tx.Exec(`INSERT INTO users (id, name, email) VALUES (?, 'Test', ?)`, id, id.String()+"@example.com").Error)
		}
		require.NoError(t, tx.Exec(
			`INSERT INTO events (id, name, description, author_id, start_date, end_date, stage, shareable_link)
			 VALUES (?, 'Test event', 'Test', ?, NOW(), NOW() + INTERVAL '30 days', 'participation', ?)`,
			eventID, author, eventID.String()).Error)
		require.NoError(t, tx.Exec(`INSERT INTO event_participants (event_id, user_id, role) VALUES (?, ?, 'creator')`, eventID, author).Error)
		upload := func(p uuid.UUID, at time.Time) {
			key := uuid.NewString()
			require.NoError(t, tx.Exec(
				`INSERT INTO attachments (event_id, participant_id, filename, original_name, file_path, file_size, mime_type, uploaded_at)
				 VALUES (?, ?, ?, 'file.pdf', ?, 100, 'application/pdf', ?)`,
				eventID, p, key, key, at).Error)
		}
		for _, p := range participants {
			require.NoError(t, tx.Exec(`INSERT INTO event_participants (event_id, user_id, role) VALUES (?, ?, 'participant')`, eventID, p).Error)
			upload(p, time.Now().Add(-time.Hour))
		}
		upload(participants[0], time.Now()) // the second proposal of the same participant

		users := &serializedParticipants{UserRepository: postgres.NewPostgresUserRepository(tx), mu: &mu}
		s := eventHandlerOnDB(tx, users)
		body := votingBody(dateStr(time.Now().AddDate(0, 0, 14)), map[string]interface{}{"attachments_per_evaluator": 2})

		blocked := patchStage(t, s, eventID, body)
		require.Equal(t, http.StatusInternalServerError, blocked.Code, "before the deployment the event cannot open voting: %s", blocked.Body.String())

		_, err := m.UpRemovingFiles(tx)
		require.NoError(t, err)

		w := patchStage(t, s, eventID, body)

		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		mu.Lock()
		holding = true
		var stage string
		require.NoError(t, tx.Raw(`SELECT stage FROM events WHERE id = ?`, eventID).Scan(&stage).Error)
		assert.Equal(t, "voting", stage)
		var assignments, ofDuplicated int
		require.NoError(t, tx.Raw(`SELECT count(*) FROM assignments WHERE event_id = ?`, eventID).Scan(&assignments).Error)
		require.NoError(t, tx.Raw(`SELECT count(*) FROM assignments WHERE event_id = ? AND participant_id = ?`, eventID, participants[0]).Scan(&ofDuplicated).Error)
		assert.Equal(t, 3, assignments)
		assert.Equal(t, 1, ofDuplicated)
	})
}
