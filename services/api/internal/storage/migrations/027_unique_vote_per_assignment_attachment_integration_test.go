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

// seedVoteFixture inserts an event in voting with one reviewer, one assignment over two proposals
// and returns the ids needed to insert votes by hand.
func seedVoteFixture(t *testing.T, tx *gorm.DB) (eventID, reviewer, assignmentID uuid.UUID, proposals [2]uuid.UUID) {
	t.Helper()
	eventID, reviewer, author, assignmentID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	for _, id := range []uuid.UUID{reviewer, author} {
		require.NoError(t, tx.Exec(`INSERT INTO users (id, name, email) VALUES (?, 'Test', ?)`, id, id.String()+"@example.com").Error)
	}
	require.NoError(t, tx.Exec(
		`INSERT INTO events (id, name, description, author_id, start_date, end_date, stage, shareable_link)
		 VALUES (?, 'Test event', 'Test', ?, NOW(), NOW() + INTERVAL '7 days', 'voting', ?)`,
		eventID, author, eventID.String()).Error)
	for i := range proposals {
		proposals[i] = uuid.New()
		owner := uuid.New() // one proposal per participant (INV-13)
		require.NoError(t, tx.Exec(`INSERT INTO users (id, name, email) VALUES (?, 'Test', ?)`, owner, owner.String()+"@example.com").Error)
		require.NoError(t, tx.Exec(
			`INSERT INTO attachments (id, event_id, participant_id, filename, original_name, file_path, file_size, mime_type)
			 VALUES (?, ?, ?, 'file.pdf', 'file.pdf', 'events/file.pdf', 100, 'application/pdf')`,
			proposals[i], eventID, owner).Error)
	}
	require.NoError(t, tx.Exec(`INSERT INTO voting_configurations (event_id, attachments_per_evaluator) VALUES (?, 2)`, eventID).Error)
	require.NoError(t, tx.Exec(
		`INSERT INTO assignments (id, event_id, participant_id, attachment_ids) VALUES (?, ?, ?, ARRAY[?, ?]::uuid[])`,
		assignmentID, eventID, reviewer, proposals[0].String(), proposals[1].String()).Error)
	return
}

func insertVote(tx *gorm.DB, eventID, reviewer, assignmentID, proposal uuid.UUID, rank int) error {
	return tx.Exec(
		`INSERT INTO votes (event_id, assignment_id, voter_id, attachment_id, rank_position) VALUES (?, ?, ?, ?, ?)`,
		eventID, assignmentID, reviewer, proposal, rank).Error
}

func TestMigration027_RejectsSecondVoteForSameProposal(t *testing.T) {
	tx := migratedTx(t)
	eventID, reviewer, assignmentID, proposals := seedVoteFixture(t, tx)

	require.NoError(t, insertVote(tx, eventID, reviewer, assignmentID, proposals[0], 1))

	tx.SavePoint("dup")
	assert.Error(t, insertVote(tx, eventID, reviewer, assignmentID, proposals[0], 2))
	tx.RollbackTo("dup")
}

func TestMigration027_DeletesRankingsWithRepeatedProposals(t *testing.T) {
	tx := migratedTx(t)
	eventID, reviewer, assignmentID, proposals := seedVoteFixture(t, tx)

	// Recreate a database from before 027: drop the index and store a repeated ranking.
	require.NoError(t, migration027Down(tx))
	require.NoError(t, insertVote(tx, eventID, reviewer, assignmentID, proposals[1], 1))
	require.NoError(t, insertVote(tx, eventID, reviewer, assignmentID, proposals[1], 2))

	require.NoError(t, migration027Up(tx))

	assert.Zero(t, count(t, tx, `SELECT count(*) FROM votes WHERE assignment_id = ?`, assignmentID))
	assert.Equal(t, int64(1), count(t, tx,
		`SELECT count(*) FROM pg_indexes WHERE indexname = 'uq_votes_assignment_attachment'`))
}
