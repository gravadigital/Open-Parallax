//go:build integration
// +build integration

package migrations

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/charmbracelet/log"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// seedParticipationEvent inserts an event in participation with its author and returns their ids.
func seedParticipationEvent(t *testing.T, tx *gorm.DB) (eventID uuid.UUID) {
	t.Helper()
	eventID, author := uuid.New(), uuid.New()
	insertUser(t, tx, author)
	require.NoError(t, tx.Exec(
		`INSERT INTO events (id, name, description, author_id, start_date, end_date, stage, shareable_link)
		 VALUES (?, 'Test event', 'Test', ?, NOW(), NOW() + INTERVAL '7 days', 'participation', ?)`,
		eventID, author, eventID.String()).Error)
	return eventID
}

func insertUser(t *testing.T, tx *gorm.DB, id uuid.UUID) {
	t.Helper()
	require.NoError(t, tx.Exec(`INSERT INTO users (id, name, email) VALUES (?, 'Test', ?)`, id, id.String()+"@example.com").Error)
}

func insertAttachment(tx *gorm.DB, eventID, participantID uuid.UUID, key string, uploadedAt time.Time) (uuid.UUID, error) {
	id := uuid.New()
	err := tx.Exec(
		`INSERT INTO attachments (id, event_id, participant_id, filename, original_name, file_path, file_size, mime_type, uploaded_at)
		 VALUES (?, ?, ?, ?, 'file.pdf', ?, 100, 'application/pdf', ?)`,
		id, eventID, participantID, key, key, uploadedAt).Error
	return id, err
}

func migration029(t *testing.T) Migration {
	t.Helper()
	for _, m := range GetMigrations() {
		if m.ID == "029" {
			return m
		}
	}
	t.Fatal("migration 029 is not registered")
	return Migration{}
}

func TestMigration029_KeepsOnlyTheLatestProposalOfEachParticipant(t *testing.T) {
	t.Run("[662-AC-01] the most recent proposal of each participant stays and the others are deleted, returning their files", func(t *testing.T) {
		tx := migratedTx(t)
		m := migration029(t)
		require.NoError(t, m.Down(tx)) // a database from before 029

		eventID := seedParticipationEvent(t, tx)
		dup, single := uuid.New(), uuid.New()
		insertUser(t, tx, dup)
		insertUser(t, tx, single)
		now := time.Now()
		_, err := insertAttachment(tx, eventID, dup, "oldest-"+dup.String(), now.Add(-2*time.Hour))
		require.NoError(t, err)
		latest, err := insertAttachment(tx, eventID, dup, "latest-"+dup.String(), now)
		require.NoError(t, err)
		_, err = insertAttachment(tx, eventID, dup, "middle-"+dup.String(), now.Add(-time.Hour))
		require.NoError(t, err)
		only, err := insertAttachment(tx, eventID, single, "only-"+single.String(), now.Add(-3*time.Hour))
		require.NoError(t, err)

		keys, err := m.UpRemovingFiles(tx)
		require.NoError(t, err)

		assert.ElementsMatch(t, []string{"oldest-" + dup.String(), "middle-" + dup.String()}, keys)
		var remaining []uuid.UUID
		require.NoError(t, tx.Raw(`SELECT id FROM attachments WHERE event_id = ?`, eventID).Scan(&remaining).Error)
		assert.ElementsMatch(t, []uuid.UUID{latest, only}, remaining)
	})
}

func TestMigration029_RejectsASecondProposalOfTheSameParticipant(t *testing.T) {
	t.Run("[INV-13][GAR-16] the database rejects a second proposal of the same participant in the event, not one in another event", func(t *testing.T) {
		tx := migratedTx(t)
		eventID, otherEvent := seedParticipationEvent(t, tx), seedParticipationEvent(t, tx)
		p := uuid.New()
		insertUser(t, tx, p)

		_, err := insertAttachment(tx, eventID, p, "first-"+p.String(), time.Now())
		require.NoError(t, err)

		tx.SavePoint("dup")
		_, err = insertAttachment(tx, eventID, p, "second-"+p.String(), time.Now())
		assert.ErrorContains(t, err, "uq_attachments_event_participant")
		tx.RollbackTo("dup")

		_, err = insertAttachment(tx, otherEvent, p, "other-"+p.String(), time.Now())
		assert.NoError(t, err)
	})
}

// fakeRemover records deletions and fails for the keys in failing.
type fakeRemover struct {
	deleted []string
	failing map[string]bool
}

func (f *fakeRemover) Delete(_ context.Context, key string) error {
	if f.failing[key] {
		return errors.New("storage unavailable")
	}
	f.deleted = append(f.deleted, key)
	return nil
}

func TestApplyMigration_DeletesTheFilesOfRemovedRecords(t *testing.T) {
	t.Run("[662-AC-01] the files of the removed proposals are deleted after the migration, and one that cannot be deleted is logged without stopping it", func(t *testing.T) {
		tx := migratedTx(t)
		m := migration029(t)
		require.NoError(t, m.Down(tx))
		require.NoError(t, tx.Exec(`DELETE FROM schema_migrations WHERE id = '029'`).Error)

		eventID := seedParticipationEvent(t, tx)
		a, b := uuid.New(), uuid.New()
		insertUser(t, tx, a)
		insertUser(t, tx, b)
		now := time.Now()
		for _, row := range []struct {
			p   uuid.UUID
			key string
			at  time.Time
		}{
			{a, "a-old", now.Add(-time.Hour)}, {a, "a-new", now},
			{b, "b-old", now.Add(-time.Hour)}, {b, "b-new", now},
		} {
			_, err := insertAttachment(tx, eventID, row.p, row.key+"-"+eventID.String(), row.at)
			require.NoError(t, err)
		}
		remover := &fakeRemover{failing: map[string]bool{"b-old-" + eventID.String(): true}}
		var logs bytes.Buffer

		err := applyMigration(tx, m, remover, log.New(&logs))

		require.NoError(t, err, "a file that cannot be deleted does not stop the deployment")
		assert.Equal(t, []string{"a-old-" + eventID.String()}, remover.deleted)
		assert.Contains(t, logs.String(), "b-old-"+eventID.String())
		assert.Contains(t, logs.String(), "storage unavailable")
		assert.Equal(t, int64(1), count(t, tx, `SELECT count(*) FROM schema_migrations WHERE id = '029'`))
		assert.Equal(t, int64(2), count(t, tx, `SELECT count(*) FROM attachments WHERE event_id = ?`, eventID))
	})

	t.Run("[662-AC-01] without a file storage the files are logged as not deleted and the migration still applies", func(t *testing.T) {
		tx := migratedTx(t)
		m := migration029(t)
		require.NoError(t, m.Down(tx))
		require.NoError(t, tx.Exec(`DELETE FROM schema_migrations WHERE id = '029'`).Error)

		eventID := seedParticipationEvent(t, tx)
		p := uuid.New()
		insertUser(t, tx, p)
		_, err := insertAttachment(tx, eventID, p, "old-"+eventID.String(), time.Now().Add(-time.Hour))
		require.NoError(t, err)
		_, err = insertAttachment(tx, eventID, p, "new-"+eventID.String(), time.Now())
		require.NoError(t, err)
		var logs bytes.Buffer

		require.NoError(t, applyMigration(tx, m, nil, log.New(&logs)))

		assert.Contains(t, logs.String(), "old-"+eventID.String())
		assert.Equal(t, int64(1), count(t, tx, `SELECT count(*) FROM attachments WHERE event_id = ?`, eventID))
	})
}
