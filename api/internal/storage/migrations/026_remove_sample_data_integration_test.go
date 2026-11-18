//go:build integration
// +build integration

package migrations

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// insertOldSampleData recreates what migration 006 used to insert, as databases that ran it have.
func insertOldSampleData(t *testing.T, tx *gorm.DB) {
	t.Helper()
	for i, id := range sampleUserIDs {
		require.NoError(t, tx.Exec(`INSERT INTO users (id, name, email, password_hash) VALUES (?, 'Sample', ?, '')`,
			id, uuid.NewString()+"@sample.test").Error, "user %d", i)
	}
	require.NoError(t, tx.Exec(
		`INSERT INTO events (id, name, description, author_id, start_date, end_date, stage, shareable_link)
		 VALUES (?, 'Distributed Time Allocation', 'Sample', ?, NOW(), NOW() + INTERVAL '6 months', 'participation', ?)`,
		sampleEventID, sampleUserIDs[0], sampleEventID).Error)
	for _, id := range sampleUserIDs[1:] {
		require.NoError(t, tx.Exec(`INSERT INTO event_participants (event_id, user_id) VALUES (?, ?)`, sampleEventID, id).Error)
	}
	require.NoError(t, tx.Exec(
		`INSERT INTO voting_configurations (event_id, attachments_per_evaluator, min_evaluations_per_file,
		   quality_good_threshold, quality_bad_threshold, adjustment_magnitude)
		 VALUES (?, 8, 4, 0.65, 0.35, 3)`, sampleEventID).Error)
}

func count(t *testing.T, tx *gorm.DB, query string, args ...interface{}) int64 {
	t.Helper()
	var n int64
	require.NoError(t, tx.Raw(query, args...).Scan(&n).Error)
	return n
}

func TestMigration026_RemovesTheSampleEventAndUsers(t *testing.T) {
	tx := migratedTx(t)
	insertOldSampleData(t, tx)

	require.NoError(t, migration026Up(tx))

	assert.Zero(t, count(t, tx, `SELECT count(*) FROM events WHERE id = ?`, sampleEventID))
	assert.Zero(t, count(t, tx, `SELECT count(*) FROM event_participants WHERE event_id = ?`, sampleEventID))
	assert.Zero(t, count(t, tx, `SELECT count(*) FROM voting_configurations WHERE event_id = ?`, sampleEventID))
	assert.Zero(t, count(t, tx, `SELECT count(*) FROM users WHERE id IN ?`, sampleUserIDs))
}

func TestMigration026_KeepsSampleUsersThatTookPartInRealEvents(t *testing.T) {
	tx := migratedTx(t)
	insertOldSampleData(t, tx)
	realEvent := insertEvent(t, tx, "1 day")
	require.NoError(t, tx.Exec(`INSERT INTO event_participants (event_id, user_id) VALUES (?, ?)`,
		realEvent, sampleUserIDs[3]).Error)

	require.NoError(t, migration026Up(tx))

	assert.Equal(t, int64(1), count(t, tx, `SELECT count(*) FROM users WHERE id IN ?`, sampleUserIDs))
	assert.Equal(t, int64(1), count(t, tx, `SELECT count(*) FROM users WHERE id = ?`, sampleUserIDs[3]))
	assert.Equal(t, int64(1), count(t, tx, `SELECT count(*) FROM events WHERE id = ?`, realEvent))
}

func TestMigration026_DoesNothingWithoutSampleData(t *testing.T) {
	tx := migratedTx(t)
	realEvent := insertEvent(t, tx, "1 day")
	users := count(t, tx, `SELECT count(*) FROM users`)

	require.NoError(t, migration026Up(tx))

	assert.Equal(t, users, count(t, tx, `SELECT count(*) FROM users`))
	assert.Equal(t, int64(1), count(t, tx, `SELECT count(*) FROM events WHERE id = ?`, realEvent))
}

func TestMigration006_NoLongerInsertsSampleData(t *testing.T) {
	tx := migratedTx(t)

	require.NoError(t, migration006Up(tx))

	assert.Zero(t, count(t, tx, `SELECT count(*) FROM events WHERE id = ?`, sampleEventID))
	assert.Zero(t, count(t, tx, `SELECT count(*) FROM users WHERE id IN ?`, sampleUserIDs))
}
