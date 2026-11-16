package migrations

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

// sampleEventID is the event migration 006 used to insert into every database.
const sampleEventID = "660e8400-e29b-41d4-a716-446655440000"

// sampleUserIDs are the users migration 006 used to insert: an admin with no password and seven
// participants of the sample event. They are matched by id, not by email: with ON CONFLICT (email)
// DO NOTHING, a real account that already had one of those emails kept its own id.
var sampleUserIDs = []string{
	"550e8400-e29b-41d4-a716-446655440000",
	"550e8400-e29b-41d4-a716-446655440001",
	"550e8400-e29b-41d4-a716-446655440002",
	"550e8400-e29b-41d4-a716-446655440003",
	"550e8400-e29b-41d4-a716-446655440004",
	"550e8400-e29b-41d4-a716-446655440005",
	"550e8400-e29b-41d4-a716-446655440006",
	"550e8400-e29b-41d4-a716-446655440007",
}

// foreignKeyViolation is the PostgreSQL error code for a row still referenced by another table.
const foreignKeyViolation = "23503"

// migration026Up removes the sample event and users migration 006 inserted. The event goes with
// everything attached to it. Each user is deleted inside its own savepoint: one that someone
// registered or used for a real event is still referenced, the delete fails with a foreign key
// violation, and that user is kept instead of aborting the migration.
func migration026Up(db *gorm.DB) error {
	eventTables := []string{
		"notifications", "vote_drafts", "votes", "voting_results", "assignments",
		"attachments", "voting_configurations", "event_participants",
	}
	for _, table := range eventTables {
		if err := db.Exec("DELETE FROM "+table+" WHERE event_id = ?", sampleEventID).Error; err != nil {
			return err
		}
	}
	if err := db.Exec("DELETE FROM events WHERE id = ?", sampleEventID).Error; err != nil {
		return err
	}

	for _, id := range sampleUserIDs {
		err := db.Transaction(func(sp *gorm.DB) error {
			return sp.Exec("DELETE FROM users WHERE id = ?", id).Error
		})
		var pgErr *pgconn.PgError
		if err != nil && !(errors.As(err, &pgErr) && pgErr.Code == foreignKeyViolation) {
			return err
		}
	}
	return nil
}

// migration026Down is a no-op: the sample data must not come back.
func migration026Down(db *gorm.DB) error {
	return nil
}
