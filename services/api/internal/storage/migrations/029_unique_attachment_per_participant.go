package migrations

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/gravadigital/openparallax/api/internal/logger"
)

// migration029Up enforces one proposal per participant and event (INV-13, GAR-16).
//
// The Go check never worked, so a participant could have several proposals in an event; those
// events cannot open voting. Before the unique index is created, each participant keeps only the
// most recent proposal of each event (DEC-10, ties broken by id) and the others are deleted. Those
// events are all in participation, so no assignment or vote refers to them. The storage keys of the
// deleted records are returned, once each and never one a remaining record still uses, so their
// files are deleted after the migration commits.
func migration029Up(db *gorm.DB) ([]string, error) {
	var removed []struct {
		ID            uuid.UUID
		EventID       uuid.UUID
		ParticipantID uuid.UUID
		FilePath      string
	}
	if err := db.Raw(`
		DELETE FROM attachments a
		USING (
			SELECT id, ROW_NUMBER() OVER (
				PARTITION BY event_id, participant_id
				ORDER BY uploaded_at DESC NULLS LAST, id DESC
			) AS position
			FROM attachments
		) ranked
		WHERE a.id = ranked.id AND ranked.position > 1
		RETURNING a.id, a.event_id, a.participant_id, a.file_path`).Scan(&removed).Error; err != nil {
		return nil, err
	}

	// Under the old naming two uploads in the same second shared one key, so a removed record can
	// point at the file of the proposal that stays: that file is kept.
	var stillUsed []string
	if len(removed) > 0 {
		paths := make([]string, 0, len(removed))
		for _, r := range removed {
			paths = append(paths, r.FilePath)
		}
		if err := db.Raw(`SELECT DISTINCT file_path FROM attachments WHERE file_path IN ?`, paths).Scan(&stillUsed).Error; err != nil {
			return nil, err
		}
	}
	skip := make(map[string]bool, len(stillUsed))
	for _, k := range stillUsed {
		skip[k] = true
	}

	log := logger.Migration()
	keys := make([]string, 0, len(removed))
	for _, r := range removed {
		log.Info("removed a duplicate proposal", "attachment_id", r.ID, "event_id", r.EventID, "participant_id", r.ParticipantID, "key", r.FilePath)
		if skip[r.FilePath] {
			log.Info("file kept: the remaining proposal uses it", "key", r.FilePath)
			continue
		}
		skip[r.FilePath] = true // each key is returned once
		keys = append(keys, r.FilePath)
	}

	if err := db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uq_attachments_event_participant
		ON attachments (event_id, participant_id)`).Error; err != nil {
		return nil, err
	}
	return keys, nil
}

// migration029Down drops the index; the deleted proposals and their files do not come back.
func migration029Down(db *gorm.DB) error {
	return db.Exec(`DROP INDEX IF EXISTS uq_attachments_event_participant`).Error
}
