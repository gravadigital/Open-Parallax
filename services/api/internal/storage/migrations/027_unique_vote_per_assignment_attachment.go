package migrations

import "gorm.io/gorm"

// migration027Up allows a single vote per (assignment, proposal). Without it a ranking that
// repeated a proposal gave it two votes, and two concurrent submissions of the same ranking
// could leave both sets of votes behind.
//
// A ranking that already repeats a proposal is invalid as a whole, so all the votes of those
// assignments are deleted before the index is created; the completion trigger then marks the
// assignment as not completed and the participant has to submit the ranking again.
func migration027Up(db *gorm.DB) error {
	stmts := []string{
		`DELETE FROM votes WHERE assignment_id IN (
			SELECT assignment_id FROM votes
			GROUP BY assignment_id, attachment_id
			HAVING COUNT(*) > 1
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_votes_assignment_attachment
			ON votes (assignment_id, attachment_id)`,
	}
	for _, s := range stmts {
		if err := db.Exec(s).Error; err != nil {
			return err
		}
	}
	return nil
}

// migration027Down drops the index; the deleted votes do not come back.
func migration027Down(db *gorm.DB) error {
	return db.Exec(`DROP INDEX IF EXISTS uq_votes_assignment_attachment`).Error
}
