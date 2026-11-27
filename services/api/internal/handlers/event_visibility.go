package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gravadigital/openparallax/api/internal/domain/event"
	"github.com/gravadigital/openparallax/api/internal/domain/participant"
	"github.com/gravadigital/openparallax/api/internal/domain/vote"
	"github.com/gravadigital/openparallax/api/internal/middleware/auth"
)

// eventVisibleTo reports whether the caller may see evt. An event in creation is
// visible only to its author and to admins; to anyone else it must look like a
// nonexistent event, so callers answer 404 EVENT_NOT_FOUND.
func eventVisibleTo(c *gin.Context, evt *event.Event) bool {
	if evt.Stage != event.StageCreation {
		return true
	}
	if role, err := auth.GetUserRoleFromContext(c); err == nil && role == participant.RoleAdmin {
		return true
	}
	userID, err := auth.GetUserIDFromContext(c)
	return err == nil && userID == evt.AuthorID
}

// respondEventNotFound writes the 404 shared by a missing event and a hidden one.
func respondEventNotFound(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{
		"error": "Event not found",
		"code":  "EVENT_NOT_FOUND",
	})
}

// resultsDetailFor reports what a viewer of evt's results may see about how each author
// evaluated (adjustments and their reasons): everything for the event's author and admins,
// only their own proposal for anyone else signed in, and nothing for an anonymous visitor.
func resultsDetailFor(c *gin.Context, evt *event.Event) (seeAll bool, viewer uuid.UUID) {
	if role, err := auth.GetUserRoleFromContext(c); err == nil && role == participant.RoleAdmin {
		return true, uuid.Nil
	}
	userID, err := auth.GetUserIDFromContext(c)
	if err != nil {
		return false, uuid.Nil
	}
	return userID == evt.AuthorID, userID
}

// filterResultsDetail removes from ranking the adjustment details the viewer may not see.
func filterResultsDetail(ranking vote.AttachmentResultSlice, seeAll bool, viewer uuid.UUID) vote.AttachmentResultSlice {
	if seeAll || ranking == nil {
		return ranking
	}
	out := make(vote.AttachmentResultSlice, len(ranking))
	for i, r := range ranking {
		if viewer != uuid.Nil && r.ParticipantID == viewer {
			out[i] = r
			continue
		}
		out[i] = r.WithoutEvaluatorDetail()
	}
	return out
}
