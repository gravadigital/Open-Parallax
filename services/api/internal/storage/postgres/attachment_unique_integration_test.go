//go:build integration
// +build integration

package postgres

import (
	"errors"
	"os"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/gravadigital/openparallax/api/internal/config"
	attachmentDomain "github.com/gravadigital/openparallax/api/internal/domain/attachment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Two uploads of the same participant reach Create at the same time, each on its own connection,
// as two requests do: the unique index lets exactly one through.
func TestAttachmentCreate_ConcurrentSecondProposal(t *testing.T) {
	t.Run("[GAR-16][INV-13][INS-E15] of two simultaneous proposals of the same participant only one is saved, the other ends in ErrDuplicateAttachment", func(t *testing.T) {
		cfg := config.Load()
		if testDB := os.Getenv("TEST_DB_NAME"); testDB != "" {
			cfg.DB.Name = testDB
		}
		db, err := Connect(cfg)
		require.NoError(t, err)
		require.NoError(t, AutoMigrate(db, nil))

		eventID, author, p := uuid.New(), uuid.New(), uuid.New()
		for _, id := range []uuid.UUID{author, p} {
			require.NoError(t, db.Exec(`INSERT INTO users (id, name, email) VALUES (?, 'Test', ?)`, id, id.String()+"@example.com").Error)
		}
		require.NoError(t, db.Exec(
			`INSERT INTO events (id, name, description, author_id, start_date, end_date, stage, shareable_link)
			 VALUES (?, 'Test event', 'Test', ?, NOW(), NOW() + INTERVAL '7 days', 'participation', ?)`,
			eventID, author, eventID.String()).Error)
		t.Cleanup(func() {
			db.Exec(`DELETE FROM attachments WHERE event_id = ?`, eventID)
			db.Exec(`DELETE FROM events WHERE id = ?`, eventID)
			db.Exec(`DELETE FROM users WHERE id IN (?, ?)`, author, p)
			if sqlDB, err := db.DB(); err == nil {
				sqlDB.Close()
			}
		})

		repo := NewPostgresAttachmentRepository(db)
		const uploads = 2
		errs := make([]error, uploads)
		var start, done sync.WaitGroup
		start.Add(1)
		for i := 0; i < uploads; i++ {
			done.Add(1)
			go func(i int) {
				defer done.Done()
				start.Wait()
				key := uuid.NewString()
				errs[i] = repo.Create(attachmentDomain.NewAttachment(eventID, p, key, "file.pdf", key, "application/pdf", 100, ""))
			}(i)
		}
		start.Done()
		done.Wait()

		var saved, duplicates int
		for _, err := range errs {
			switch {
			case err == nil:
				saved++
			case errors.Is(err, ErrDuplicateAttachment):
				duplicates++
			default:
				t.Fatalf("unexpected error: %v", err)
			}
		}
		assert.Equal(t, 1, saved)
		assert.Equal(t, 1, duplicates)
		var n int
		require.NoError(t, db.Raw(`SELECT count(*) FROM attachments WHERE event_id = ? AND participant_id = ?`, eventID, p).Scan(&n).Error)
		assert.Equal(t, 1, n)
	})
}
