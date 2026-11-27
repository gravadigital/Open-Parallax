package vote

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func defaultConfig(eventID uuid.UUID, m int) *VotingConfiguration {
	return &VotingConfiguration{
		ID:                      uuid.New(),
		EventID:                 eventID,
		AttachmentsPerEvaluator: m,
		QualityGoodThreshold:    0.6,
		QualityBadThreshold:     0.3,
		AdjustmentMagnitude:     3,
		MinEvaluationsPerFile:   1,
		NonEvaluatorSanction:    SanctionExclusion,
		NonEvaluatorPenalty:     DefaultNonEvaluatorPenalty,
	}
}

func newTestService() (*VotingService, *fakeVoteRepository, *fakeAttachmentRepository, *fakeUserRepository) {
	voteRepo := newFakeVoteRepository()
	attachmentRepo := newFakeAttachmentRepository()
	userRepo := newFakeUserRepository()
	return NewVotingService(voteRepo, attachmentRepo, userRepo), voteRepo, attachmentRepo, userRepo
}

// ---------------------------------------------------------------------------
// GenerateAssignments
// ---------------------------------------------------------------------------

func TestGenerateAssignments_ValidConfiguration(t *testing.T) {
	vs, _, attachmentRepo, _ := newTestService()
	eventID := uuid.New()

	participants := []uuid.UUID{uuid.New(), uuid.New(), uuid.New(), uuid.New()}
	attachments := make([]uuid.UUID, 4)
	for i, p := range participants {
		attachments[i] = uuid.New()
		attachmentRepo.add(eventID, &fakeAttachment{id: attachments[i], participantID: p, originalName: "photo.jpg"})
	}

	config := defaultConfig(eventID, 2) // recommendedM for k=4 with 60% flexibility should allow m=2
	config.MinEvaluationsPerFile = 1

	assignments, err := vs.GenerateAssignments(eventID, participants, attachments, config)
	require.NoError(t, err)
	require.Len(t, assignments, len(participants))

	for i, a := range assignments {
		assert.Equal(t, participants[i], a.ParticipantID)
		assert.LessOrEqual(t, len(a.AttachmentIDs), config.AttachmentsPerEvaluator)
		// A participant must never be assigned their own attachment.
		for _, assignedID := range a.GetAttachmentUUIDs() {
			assert.NotEqual(t, attachments[i], assignedID, "participant should not evaluate their own attachment")
		}
	}
}

// MinEvaluationsPerFile no longer drives the assignment (every proposal gets exactly m), so even
// an extreme value must not push anyone past m. Kept from the earlier two-phase algorithm, which
// once did.
func TestGenerateAssignments_NeverExceedsMEvenWithHighMinEvaluationsPerFile(t *testing.T) {
	vs, _, attachmentRepo, _ := newTestService()
	eventID := uuid.New()

	const n = 4
	const k = 4
	participants := make([]uuid.UUID, n)
	attachments := make([]uuid.UUID, k)
	for i := range participants {
		participants[i] = uuid.New()
		attachments[i] = uuid.New()
		attachmentRepo.add(eventID, &fakeAttachment{id: attachments[i], participantID: participants[i]})
	}

	config := defaultConfig(eventID, 2)
	config.MinEvaluationsPerFile = n

	for attempt := 0; attempt < 50; attempt++ {
		assignments, err := vs.GenerateAssignments(eventID, participants, attachments, config)
		require.NoError(t, err)

		for _, a := range assignments {
			assert.LessOrEqualf(t, len(a.AttachmentIDs), config.AttachmentsPerEvaluator,
				"attempt %d: participant assigned more than m=%d attachments", attempt, config.AttachmentsPerEvaluator)
		}
	}
}

// ownersSetup registers k proposals, one per participant, and returns both lists aligned.
func ownersSetup(t *testing.T, k int) (*VotingService, uuid.UUID, []uuid.UUID, []uuid.UUID, map[uuid.UUID]uuid.UUID) {
	t.Helper()
	vs, _, attachmentRepo, _ := newTestService()
	eventID := uuid.New()
	participants := make([]uuid.UUID, k)
	attachments := make([]uuid.UUID, k)
	owner := make(map[uuid.UUID]uuid.UUID, k)
	for i := range participants {
		participants[i], attachments[i] = uuid.New(), uuid.New()
		owner[attachments[i]] = participants[i]
		attachmentRepo.add(eventID, &fakeAttachment{id: attachments[i], participantID: participants[i]})
	}
	return vs, eventID, participants, attachments, owner
}

// Audit finding C4: every proposal must get exactly m reviews, not "between 4 and 29".
func TestGenerateAssignments_IsRegular(t *testing.T) {
	for _, k := range []int{3, 4, 5, 10, 20, 100} {
		_, maxM, recommended := VotingBounds(k)
		for _, m := range []int{recommended, maxM} {
			vs, eventID, participants, attachments, owner := ownersSetup(t, k)
			for run := 0; run < 20; run++ {
				assignments, err := vs.GenerateAssignments(eventID, participants, attachments, defaultConfig(eventID, m))
				require.NoError(t, err, "k=%d m=%d", k, m)
				require.Len(t, assignments, k)

				received := make(map[uuid.UUID]int, k)
				for i, a := range assignments {
					require.Equal(t, participants[i], a.ParticipantID, "assignments keep the order of the evaluators")
					ids := a.GetAttachmentUUIDs()
					require.Len(t, ids, m, "k=%d m=%d: each evaluator reviews exactly m", k, m)
					seen := map[uuid.UUID]bool{}
					for _, id := range ids {
						require.NotEqual(t, a.ParticipantID, owner[id], "k=%d m=%d: own proposal assigned", k, m)
						require.False(t, seen[id], "k=%d m=%d: proposal repeated in one assignment", k, m)
						seen[id] = true
						received[id]++
					}
				}
				for _, id := range attachments {
					require.Equal(t, m, received[id], "k=%d m=%d: each proposal is reviewed exactly m times", k, m)
				}
			}
		}
	}
}

// The circle is only the starting point: the switches must leave assignments that differ
// between runs and that do not always pair the same reviewers.
func TestGenerateAssignments_IsRandomized(t *testing.T) {
	const k, m = 20, 9
	vs, eventID, participants, attachments, _ := ownersSetup(t, k)

	distinct := map[string]bool{}
	pairCount := map[[2]uuid.UUID]int{}
	const runs = 30
	for run := 0; run < runs; run++ {
		assignments, err := vs.GenerateAssignments(eventID, participants, attachments, defaultConfig(eventID, m))
		require.NoError(t, err)
		key := ""
		for _, a := range assignments {
			ids := a.GetAttachmentUUIDs()
			key += fmt.Sprint(ids)
			for _, id := range ids {
				pairCount[[2]uuid.UUID{a.ParticipantID, id}]++
			}
		}
		distinct[key] = true
	}
	assert.Equal(t, runs, len(distinct), "every run should produce a different assignment")
	for pair, c := range pairCount {
		assert.Less(t, c, runs, "pair %v appeared in every run", pair)
	}
}

func TestGenerateAssignments_RejectsEvaluatorsThatAreNotTheOwners(t *testing.T) {
	vs, eventID, participants, attachments, _ := ownersSetup(t, 4)

	_, err := vs.GenerateAssignments(eventID, participants[:3], attachments, defaultConfig(eventID, 2))
	assert.ErrorContains(t, err, "one evaluator per proposal")

	stranger := append([]uuid.UUID{uuid.New()}, participants[1:]...)
	_, err = vs.GenerateAssignments(eventID, stranger, attachments, defaultConfig(eventID, 2))
	assert.ErrorContains(t, err, "not among the evaluators")
}

func TestRegularAssignment_SmallestCase(t *testing.T) {
	// k=3, m=2: the only valid assignment is "everyone reviews the other two".
	own := []int{0, 1, 2}
	reviews := regularAssignment(own, 2, rand.New(rand.NewSource(1)))
	for e, list := range reviews {
		assert.ElementsMatch(t, []int{(e + 1) % 3, (e + 2) % 3}, list)
	}
}

func TestGenerateAssignments_RejectsMExceedingConflictOfInterestBound(t *testing.T) {
	vs, _, attachmentRepo, _ := newTestService()
	eventID := uuid.New()

	participants := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
	attachments := make([]uuid.UUID, 3)
	for i, p := range participants {
		attachments[i] = uuid.New()
		attachmentRepo.add(eventID, &fakeAttachment{id: attachments[i], participantID: p})
	}

	// k=3, so maxPossibleM = k-1 = 2. Requesting m=3 must fail.
	config := defaultConfig(eventID, 3)

	_, err := vs.GenerateAssignments(eventID, participants, attachments, config)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot exceed")
}

func TestGenerateAssignments_RejectsMBelowRecommendedMinimum(t *testing.T) {
	vs, _, attachmentRepo, _ := newTestService()
	eventID := uuid.New()

	// Large k so the recommended minimum m (2*log2(k)) is meaningfully > 1.
	const k = 64
	participants := make([]uuid.UUID, k)
	attachments := make([]uuid.UUID, k)
	for i := range participants {
		participants[i] = uuid.New()
		attachments[i] = uuid.New()
		attachmentRepo.add(eventID, &fakeAttachment{id: attachments[i], participantID: participants[i]})
	}

	config := defaultConfig(eventID, 1)

	_, err := vs.GenerateAssignments(eventID, participants, attachments, config)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "recommended minimum")
}

func TestGenerateAssignments_NoParticipantEvaluatesOwnAttachment(t *testing.T) {
	vs, _, attachmentRepo, _ := newTestService()
	eventID := uuid.New()

	participants := []uuid.UUID{uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()}
	attachments := make([]uuid.UUID, 5)
	for i, p := range participants {
		attachments[i] = uuid.New()
		attachmentRepo.add(eventID, &fakeAttachment{id: attachments[i], participantID: p})
	}

	// k=5 -> maxPossibleM=4; with k<=10 the lenient recommendation is
	// ceil(maxPossibleM*0.6)=3, so m must be at least 3 here.
	config := defaultConfig(eventID, 3)
	config.MinEvaluationsPerFile = 2

	assignments, err := vs.GenerateAssignments(eventID, participants, attachments, config)
	require.NoError(t, err)

	for i, a := range assignments {
		for _, assignedID := range a.GetAttachmentUUIDs() {
			assert.NotEqual(t, attachments[i], assignedID)
		}
	}
}

// ---------------------------------------------------------------------------
// CalculateModifiedBordaCount - validation and error paths
// ---------------------------------------------------------------------------

func TestCalculateModifiedBordaCount_NilConfig(t *testing.T) {
	vs, _, _, _ := newTestService()
	_, err := vs.CalculateModifiedBordaCount(uuid.New(), nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "voting configuration is required")
}

func TestCalculateModifiedBordaCount_MTooSmall(t *testing.T) {
	vs, _, _, _ := newTestService()
	eventID := uuid.New()
	config := defaultConfig(eventID, 1)
	_, err := vs.CalculateModifiedBordaCount(eventID, config)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "at least 2")
}

func TestCalculateModifiedBordaCount_NoVotes(t *testing.T) {
	vs, _, _, _ := newTestService()
	eventID := uuid.New()
	config := defaultConfig(eventID, 2)

	_, err := vs.CalculateModifiedBordaCount(eventID, config)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no votes found")
}

func TestCalculateModifiedBordaCount_NoAttachments(t *testing.T) {
	vs, voteRepo, _, _ := newTestService()
	eventID := uuid.New()
	config := defaultConfig(eventID, 2)

	voteRepo.votes = append(voteRepo.votes, &Vote{
		ID: uuid.New(), EventID: eventID, VoterID: uuid.New(),
		AttachmentID: uuid.New(), RankPosition: 1,
	})

	_, err := vs.CalculateModifiedBordaCount(eventID, config)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no attachments found")
}

// ---------------------------------------------------------------------------
// CalculateModifiedBordaCount - core scoring behaviour
// ---------------------------------------------------------------------------

// setupThreeWayVoting builds a scenario with 3 participants, each submitting one
// attachment and evaluating the other two (m=2), fully symmetric and complete.
func setupThreeWayVoting(t *testing.T) (vs *VotingService, voteRepo *fakeVoteRepository, eventID uuid.UUID, participants, attachments []uuid.UUID, config *VotingConfiguration) {
	t.Helper()
	vs, voteRepo, attachmentRepo, _ := newTestService()
	eventID = uuid.New()

	participants = []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
	attachments = make([]uuid.UUID, 3)
	for i, p := range participants {
		attachments[i] = uuid.New()
		attachmentRepo.add(eventID, &fakeAttachment{id: attachments[i], participantID: p, originalName: "a.jpg"})
	}
	config = defaultConfig(eventID, 2)
	config.MinEvaluationsPerFile = 1
	return
}

func addAssignment(voteRepo *fakeVoteRepository, eventID, participantID uuid.UUID, attachmentIDs []uuid.UUID, completed bool) *Assignment {
	a := NewAssignment(eventID, participantID, attachmentIDs)
	if completed {
		a.MarkCompleted()
	}
	voteRepo.assignments = append(voteRepo.assignments, a)
	return a
}

func TestCalculateModifiedBordaCount_UnanimousRankingProducesExpectedOrder(t *testing.T) {
	vs, voteRepo, eventID, participants, attachments, config := setupThreeWayVoting(t)

	// participants[0] evaluates attachments[1] (best) and attachments[2] (worst)
	// participants[1] evaluates attachments[0] (best) and attachments[2] (worst)
	// participants[2] evaluates attachments[0] (best) and attachments[1] (worst)
	// => attachments[0] gets two 1st-place votes, attachments[1] one 1st + one 2nd,
	//    attachments[2] gets two 2nd-place (last) votes.
	votes := []*Vote{
		{EventID: eventID, VoterID: participants[0], AttachmentID: attachments[1], RankPosition: 1},
		{EventID: eventID, VoterID: participants[0], AttachmentID: attachments[2], RankPosition: 2},
		{EventID: eventID, VoterID: participants[1], AttachmentID: attachments[0], RankPosition: 1},
		{EventID: eventID, VoterID: participants[1], AttachmentID: attachments[2], RankPosition: 2},
		{EventID: eventID, VoterID: participants[2], AttachmentID: attachments[0], RankPosition: 1},
		{EventID: eventID, VoterID: participants[2], AttachmentID: attachments[1], RankPosition: 2},
	}
	for _, v := range votes {
		v.ID = uuid.New()
		voteRepo.votes = append(voteRepo.votes, v)
	}

	addAssignment(voteRepo, eventID, participants[0], []uuid.UUID{attachments[1], attachments[2]}, true)
	addAssignment(voteRepo, eventID, participants[1], []uuid.UUID{attachments[0], attachments[2]}, true)
	addAssignment(voteRepo, eventID, participants[2], []uuid.UUID{attachments[0], attachments[1]}, true)

	results, err := vs.CalculateModifiedBordaCount(eventID, config)
	require.NoError(t, err)

	require.Len(t, results.GlobalRanking, 3)
	assert.Equal(t, attachments[0], results.GlobalRanking[0].AttachmentID, "attachment with two 1st-place votes should rank first")
	assert.Equal(t, 1, results.GlobalRanking[0].GlobalRank)
	assert.InDelta(t, 1.0, results.GlobalRanking[0].MBCScore, 1e-9, "two 1st-place votes out of m(m-1)=2 should normalize to 1.0")

	assert.Equal(t, attachments[2], results.GlobalRanking[2].AttachmentID, "attachment with two last-place votes should rank last")
	assert.InDelta(t, 0.0, results.GlobalRanking[2].MBCScore, 1e-9)

	assert.Equal(t, 6, results.TotalVotes)
	assert.Equal(t, 3, results.TotalParticipants, "quality map should include all assignments, completed or not")
}

func TestCalculateModifiedBordaCount_TieBreaksByID(t *testing.T) {
	vs, voteRepo, eventID, participants, attachments, config := setupThreeWayVoting(t)

	votes := []*Vote{
		{EventID: eventID, VoterID: participants[2], AttachmentID: attachments[0], RankPosition: 1},
		{EventID: eventID, VoterID: participants[2], AttachmentID: attachments[1], RankPosition: 2},
		{EventID: eventID, VoterID: participants[1], AttachmentID: attachments[0], RankPosition: 2},
		{EventID: eventID, VoterID: participants[1], AttachmentID: attachments[2], RankPosition: 1},
		{EventID: eventID, VoterID: participants[0], AttachmentID: attachments[1], RankPosition: 1},
		{EventID: eventID, VoterID: participants[0], AttachmentID: attachments[2], RankPosition: 2},
	}
	for _, v := range votes {
		v.ID = uuid.New()
		voteRepo.votes = append(voteRepo.votes, v)
	}
	addAssignment(voteRepo, eventID, participants[0], []uuid.UUID{attachments[1], attachments[2]}, true)
	addAssignment(voteRepo, eventID, participants[1], []uuid.UUID{attachments[0], attachments[2]}, true)
	addAssignment(voteRepo, eventID, participants[2], []uuid.UUID{attachments[0], attachments[1]}, true)

	results, err := vs.CalculateModifiedBordaCount(eventID, config)
	require.NoError(t, err)

	// attachments[0]: ranks {1,2} -> mbcSum = (2-1)+(2-2) = 1 -> score 0.5, voteCount 2
	// attachments[1]: ranks {2,1} -> mbcSum = (2-2)+(2-1) = 1 -> score 0.5, voteCount 2
	// attachments[2]: ranks {1,2} -> mbcSum = (2-1)+(2-2) = 1 -> score 0.5, voteCount 2
	// All three tie on score and vote count, so order must fall back to attachment ID string ordering.
	assert.InDelta(t, results.GlobalRanking[0].MBCScore, results.GlobalRanking[1].MBCScore, 1e-9)
	assert.InDelta(t, results.GlobalRanking[1].MBCScore, results.GlobalRanking[2].MBCScore, 1e-9)
	assert.True(t, results.GlobalRanking[0].AttachmentID.String() < results.GlobalRanking[1].AttachmentID.String())
	assert.True(t, results.GlobalRanking[1].AttachmentID.String() < results.GlobalRanking[2].AttachmentID.String())
}

// Four participants, m=2. participants[3] never votes, so attachments[0] gets a single
// 1st-place vote while attachments[1] gets two. Both were ranked first every time they were
// evaluated, so they must score the same: the MBC averages over votes received instead of
// rewarding the attachment that happened to get more evaluators.
func TestCalculateModifiedBordaCount_ScoreDoesNotDependOnVoteCount(t *testing.T) {
	vs, voteRepo, attachmentRepo, _ := newTestService()
	eventID := uuid.New()
	participants := []uuid.UUID{uuid.New(), uuid.New(), uuid.New(), uuid.New()}
	attachments := make([]uuid.UUID, 4)
	for i, p := range participants {
		attachments[i] = uuid.New()
		attachmentRepo.add(eventID, &fakeAttachment{id: attachments[i], participantID: p, originalName: "a.jpg"})
	}
	config := defaultConfig(eventID, 2)

	votes := []*Vote{
		{EventID: eventID, VoterID: participants[0], AttachmentID: attachments[1], RankPosition: 1},
		{EventID: eventID, VoterID: participants[0], AttachmentID: attachments[2], RankPosition: 2},
		{EventID: eventID, VoterID: participants[1], AttachmentID: attachments[1], RankPosition: 1},
		{EventID: eventID, VoterID: participants[1], AttachmentID: attachments[3], RankPosition: 2},
		{EventID: eventID, VoterID: participants[2], AttachmentID: attachments[0], RankPosition: 1},
		{EventID: eventID, VoterID: participants[2], AttachmentID: attachments[3], RankPosition: 2},
	}
	for _, v := range votes {
		v.ID = uuid.New()
		voteRepo.votes = append(voteRepo.votes, v)
	}
	addAssignment(voteRepo, eventID, participants[0], []uuid.UUID{attachments[1], attachments[2]}, true)
	addAssignment(voteRepo, eventID, participants[1], []uuid.UUID{attachments[1], attachments[3]}, true)
	addAssignment(voteRepo, eventID, participants[2], []uuid.UUID{attachments[0], attachments[3]}, true)
	addAssignment(voteRepo, eventID, participants[3], []uuid.UUID{attachments[0], attachments[2]}, false)

	results, err := vs.CalculateModifiedBordaCount(eventID, config)
	require.NoError(t, err)

	byID := map[uuid.UUID]AttachmentResult{}
	for _, r := range results.GlobalRanking {
		byID[r.AttachmentID] = r
	}
	assert.Equal(t, 1, byID[attachments[0]].VoteCount)
	assert.Equal(t, 2, byID[attachments[1]].VoteCount)
	assert.InDelta(t, 1.0, byID[attachments[0]].MBCScore, 1e-9)
	assert.InDelta(t, 1.0, byID[attachments[1]].MBCScore, 1e-9)

	// The tie between them is broken by ID, not by who got more votes.
	top := results.GlobalRanking[:2]
	assert.True(t, top[0].AttachmentID.String() < top[1].AttachmentID.String())
}

func TestCalculateModifiedBordaCount_AttachmentWithNoVotesHasNoPosition(t *testing.T) {
	vs, voteRepo, eventID, participants, attachments, config := setupThreeWayVoting(t)
	// Penalty keeps the non-evaluators' proposals in G', so this test only sees "unevaluated".
	config.NonEvaluatorSanction = SanctionPenalty

	// Only participants[0] votes, and only on attachments[1]. attachments[0] and
	// attachments[2] receive zero votes each.
	voteRepo.votes = append(voteRepo.votes, &Vote{
		ID: uuid.New(), EventID: eventID, VoterID: participants[0],
		AttachmentID: attachments[1], RankPosition: 1,
	})
	addAssignment(voteRepo, eventID, participants[0], []uuid.UUID{attachments[1]}, true)
	addAssignment(voteRepo, eventID, participants[1], []uuid.UUID{attachments[0]}, false)
	addAssignment(voteRepo, eventID, participants[2], []uuid.UUID{attachments[2]}, false)

	results, err := vs.CalculateModifiedBordaCount(eventID, config)
	require.NoError(t, err)

	byID := map[uuid.UUID]AttachmentResult{}
	for _, r := range results.GlobalRanking {
		byID[r.AttachmentID] = r
	}

	// attachments[0] and attachments[2] got no votes: no score to rank by, so no position.
	for _, id := range []uuid.UUID{attachments[0], attachments[2]} {
		assert.Equal(t, 0, byID[id].VoteCount)
		assert.True(t, byID[id].Unevaluated)
		assert.Equal(t, 0, byID[id].GlobalRank)
	}

	// attachments[1] got the only vote and is the only ranked one.
	assert.False(t, byID[attachments[1]].Unevaluated)
	assert.Equal(t, 1, byID[attachments[1]].GlobalRank)

	// Both rankings list the ranked attachment first, then the unevaluated ones in ID order.
	for name, ranking := range map[string][]AttachmentResult{"global": results.GlobalRanking, "adjusted": results.AdjustedRanking} {
		require.Len(t, ranking, 3, name)
		assert.Equal(t, attachments[1], ranking[0].AttachmentID, name)
		assert.True(t, ranking[1].Unevaluated && ranking[2].Unevaluated, name)
		assert.True(t, ranking[1].AttachmentID.String() < ranking[2].AttachmentID.String(), name)
	}
	for _, r := range results.AdjustedRanking[1:] {
		assert.Equal(t, 0, r.AdjustedRank, "the incentive must not give an unevaluated attachment a position")
	}
}

// ---------------------------------------------------------------------------
// Participant who does not complete their ranking (Merrifield & Saari 2009:
// the proposal of whoever does not referee is removed from the ranking)
// ---------------------------------------------------------------------------

// setupOneNonEvaluator: participants[0] and participants[1] complete their rankings;
// participants[2] only ranks one of their two proposals (a partial, incomplete ranking).
func setupOneNonEvaluator(t *testing.T) (*VotingService, uuid.UUID, []uuid.UUID, []uuid.UUID, *VotingConfiguration) {
	t.Helper()
	vs, voteRepo, eventID, participants, attachments, config := setupThreeWayVoting(t)
	votes := []*Vote{
		{EventID: eventID, VoterID: participants[0], AttachmentID: attachments[1], RankPosition: 1},
		{EventID: eventID, VoterID: participants[0], AttachmentID: attachments[2], RankPosition: 2},
		{EventID: eventID, VoterID: participants[1], AttachmentID: attachments[2], RankPosition: 1},
		{EventID: eventID, VoterID: participants[1], AttachmentID: attachments[0], RankPosition: 2},
		// Partial vote of the non-evaluator: must be discarded.
		{EventID: eventID, VoterID: participants[2], AttachmentID: attachments[0], RankPosition: 1},
	}
	for _, v := range votes {
		v.ID = uuid.New()
		voteRepo.votes = append(voteRepo.votes, v)
	}
	addAssignment(voteRepo, eventID, participants[0], []uuid.UUID{attachments[1], attachments[2]}, true)
	addAssignment(voteRepo, eventID, participants[1], []uuid.UUID{attachments[2], attachments[0]}, true)
	addAssignment(voteRepo, eventID, participants[2], []uuid.UUID{attachments[0], attachments[1]}, false)
	return vs, eventID, participants, attachments, config
}

func byAttachment(ranking []AttachmentResult) map[uuid.UUID]AttachmentResult {
	out := map[uuid.UUID]AttachmentResult{}
	for _, r := range ranking {
		out[r.AttachmentID] = r
	}
	return out
}

func TestCalculateModifiedBordaCount_DiscardsPartialVotesOfNonEvaluator(t *testing.T) {
	vs, eventID, _, attachments, config := setupOneNonEvaluator(t)

	results, err := vs.CalculateModifiedBordaCount(eventID, config)
	require.NoError(t, err)

	global := byAttachment(results.GlobalRanking)
	// attachments[0] only counts the vote of participants[1] (last of 2 → 0 points); the
	// non-evaluator's first place is ignored.
	assert.Equal(t, 1, global[attachments[0]].VoteCount)
	assert.InDelta(t, 0.0, global[attachments[0]].MBCScore, 1e-9)
	assert.Equal(t, 4, results.TotalVotes)
}

func TestCalculateModifiedBordaCount_NonEvaluatorHasNoQuality(t *testing.T) {
	vs, eventID, participants, _, config := setupOneNonEvaluator(t)

	results, err := vs.CalculateModifiedBordaCount(eventID, config)
	require.NoError(t, err)

	assert.NotContains(t, results.ParticipantQualities, participants[2].String())
	assert.Contains(t, results.ParticipantQualities, participants[0].String())
	assert.Equal(t, 3, results.TotalParticipants)
}

func TestCalculateModifiedBordaCount_ExclusionRemovesNonEvaluatorFromAdjustedRanking(t *testing.T) {
	vs, eventID, participants, attachments, config := setupOneNonEvaluator(t)
	config.NonEvaluatorSanction = SanctionExclusion

	results, err := vs.CalculateModifiedBordaCount(eventID, config)
	require.NoError(t, err)

	// G keeps it: the votes it received are valid.
	global := byAttachment(results.GlobalRanking)
	require.Len(t, results.GlobalRanking, 3)
	assert.NotZero(t, global[attachments[2]].GlobalRank)
	assert.True(t, global[attachments[2]].NonEvaluator)

	// G' lists it last, without a position.
	require.Len(t, results.AdjustedRanking, 3)
	last := results.AdjustedRanking[2]
	assert.Equal(t, participants[2], last.ParticipantID)
	assert.True(t, last.Excluded)
	assert.Equal(t, 0, last.AdjustedRank)
	assert.False(t, last.HasAdjustedPosition())
	for i, r := range results.AdjustedRanking[:2] {
		assert.False(t, r.Excluded)
		assert.Equal(t, i+1, r.AdjustedRank, "the others keep consecutive positions")
	}
}

func TestCalculateModifiedBordaCount_EmptySanctionMeansExclusion(t *testing.T) {
	vs, eventID, participants, _, config := setupOneNonEvaluator(t)
	config.NonEvaluatorSanction = "" // configuration stored before the sanction existed

	results, err := vs.CalculateModifiedBordaCount(eventID, config)
	require.NoError(t, err)

	last := results.AdjustedRanking[len(results.AdjustedRanking)-1]
	assert.Equal(t, participants[2], last.ParticipantID)
	assert.True(t, last.Excluded)
}

func TestCalculateModifiedBordaCount_PenaltyMovesNonEvaluatorDown(t *testing.T) {
	vs, eventID, participants, _, config := setupOneNonEvaluator(t)
	config.NonEvaluatorSanction = SanctionPenalty
	config.NonEvaluatorPenalty = 10 // larger than the table: it ends last

	results, err := vs.CalculateModifiedBordaCount(eventID, config)
	require.NoError(t, err)

	require.Len(t, results.AdjustedRanking, 3)
	for _, r := range results.AdjustedRanking {
		assert.False(t, r.Excluded)
		assert.NotZero(t, r.AdjustedRank)
	}
	last := results.AdjustedRanking[2]
	assert.Equal(t, participants[2], last.ParticipantID)
	assert.True(t, last.NonEvaluator)
	assert.Equal(t, 3, last.AdjustedRank)
}

func TestCalculateModifiedBordaCount_AllParticipantsAbsentExceptOneStillProducesResult(t *testing.T) {
	vs, voteRepo, eventID, participants, attachments, config := setupThreeWayVoting(t)

	// Only one vote exists in the whole event; everyone else's assignments are incomplete.
	voteRepo.votes = append(voteRepo.votes, &Vote{
		ID: uuid.New(), EventID: eventID, VoterID: participants[0],
		AttachmentID: attachments[1], RankPosition: 1,
	})
	addAssignment(voteRepo, eventID, participants[0], []uuid.UUID{attachments[1]}, true)
	addAssignment(voteRepo, eventID, participants[1], []uuid.UUID{attachments[0]}, false)
	addAssignment(voteRepo, eventID, participants[2], []uuid.UUID{attachments[2]}, false)

	results, err := vs.CalculateModifiedBordaCount(eventID, config)
	require.NoError(t, err, "a single completed vote must be enough to calculate results without error")

	assert.NotContains(t, results.ParticipantQualities, participants[1].String())
	assert.NotContains(t, results.ParticipantQualities, participants[2].String())
	// Only participants[0]'s proposal stays in G' (it got no votes, so it is unevaluated);
	// the other two are excluded.
	excluded := 0
	for _, r := range results.AdjustedRanking {
		if r.Excluded {
			excluded++
		}
	}
	assert.Equal(t, 2, excluded)
}

// ---------------------------------------------------------------------------
// calculateParticipantQualities - Q_i normalization (Merrifield & Saari 2009)
// ---------------------------------------------------------------------------

func TestMaxRankDeviation(t *testing.T) {
	// ⌊m²/2⌋: the Σ|R_i - RelativeRank| of a ranking in exactly reverse order.
	for m, want := range map[int]int{2: 2, 3: 4, 4: 8, 5: 12} {
		assert.Equal(t, want, maxRankDeviation(m), "m=%d", m)
	}
}

func TestCalculateParticipantQualities_SpansZeroToOne(t *testing.T) {
	// Global order a0 > a1 > a2; one evaluator per way of ranking that same subset.
	cases := []struct {
		name  string
		ranks [3]int // positions given to a0, a1, a2
		want  float64
	}{
		{"same order as the consensus", [3]int{1, 2, 3}, 1.0},
		{"one adjacent swap", [3]int{2, 1, 3}, 0.5},
		{"exact reverse", [3]int{3, 2, 1}, 0.0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			vs, voteRepo, _, _ := newTestService()
			eventID := uuid.New()
			attachments := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
			global := make([]AttachmentResult, 3)
			for i, a := range attachments {
				global[i] = AttachmentResult{AttachmentID: a, GlobalRank: i + 1}
			}

			evaluator := uuid.New()
			for i, a := range attachments {
				voteRepo.votes = append(voteRepo.votes, &Vote{
					ID: uuid.New(), EventID: eventID, VoterID: evaluator,
					AttachmentID: a, RankPosition: tc.ranks[i],
				})
			}
			addAssignment(voteRepo, eventID, evaluator, attachments, true)

			qualities, err := vs.calculateParticipantQualities(eventID, voteRepo.assignments, global, 3)
			require.NoError(t, err)
			assert.InDelta(t, tc.want, qualities[evaluator.String()], 1e-9)
		})
	}
}

// ---------------------------------------------------------------------------
// applyIncentiveSystem behaviour (via CalculateModifiedBordaCount)
// ---------------------------------------------------------------------------

func TestCalculateModifiedBordaCount_GoodEvaluatorBonusImprovesOwnRank(t *testing.T) {
	vs, voteRepo, eventID, participants, attachments, config := setupThreeWayVoting(t)
	config.AdjustmentMagnitude = 1
	config.QualityGoodThreshold = 0.5

	// All three vote in perfect agreement with the eventual global ranking,
	// so every participant should be a "good" evaluator (quality >= threshold).
	votes := []*Vote{
		{EventID: eventID, VoterID: participants[0], AttachmentID: attachments[1], RankPosition: 1},
		{EventID: eventID, VoterID: participants[0], AttachmentID: attachments[2], RankPosition: 2},
		{EventID: eventID, VoterID: participants[1], AttachmentID: attachments[0], RankPosition: 1},
		{EventID: eventID, VoterID: participants[1], AttachmentID: attachments[2], RankPosition: 2},
		{EventID: eventID, VoterID: participants[2], AttachmentID: attachments[0], RankPosition: 1},
		{EventID: eventID, VoterID: participants[2], AttachmentID: attachments[1], RankPosition: 2},
	}
	for _, v := range votes {
		v.ID = uuid.New()
		voteRepo.votes = append(voteRepo.votes, v)
	}
	addAssignment(voteRepo, eventID, participants[0], []uuid.UUID{attachments[1], attachments[2]}, true)
	addAssignment(voteRepo, eventID, participants[1], []uuid.UUID{attachments[0], attachments[2]}, true)
	addAssignment(voteRepo, eventID, participants[2], []uuid.UUID{attachments[0], attachments[1]}, true)

	results, err := vs.CalculateModifiedBordaCount(eventID, config)
	require.NoError(t, err)

	// Adjusted ranks must stay within valid bounds [1, len(results)] regardless of adjustment.
	for _, r := range results.AdjustedRanking {
		assert.GreaterOrEqual(t, r.AdjustedRank, 1)
		assert.LessOrEqual(t, r.AdjustedRank, len(results.AdjustedRanking))
	}

	// Adjusted ranking must still be a permutation of 1..N (consecutive, no gaps/dupes).
	seen := map[int]bool{}
	for _, r := range results.AdjustedRanking {
		assert.False(t, seen[r.AdjustedRank], "adjusted ranks must be unique after reassignment")
		seen[r.AdjustedRank] = true
	}
	assert.Len(t, seen, 3)
}

// ---------------------------------------------------------------------------
// ValidateVotingConfiguration
// ---------------------------------------------------------------------------

func TestValidateVotingConfiguration_MNonPositive(t *testing.T) {
	vs, _, _, _ := newTestService()
	config := defaultConfig(uuid.New(), 0)
	err := vs.ValidateVotingConfiguration(config, 10, 10)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must be positive")
}

func TestValidateVotingConfiguration_MExceedsConflictBound(t *testing.T) {
	vs, _, _, _ := newTestService()
	config := defaultConfig(uuid.New(), 5)
	err := vs.ValidateVotingConfiguration(config, 5, 5) // maxPossibleM = k-1 = 4
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot exceed")
}

func TestValidateVotingConfiguration_BadThresholdOrdering(t *testing.T) {
	vs, _, _, _ := newTestService()
	config := defaultConfig(uuid.New(), 2)
	config.QualityGoodThreshold = 0.3
	config.QualityBadThreshold = 0.6
	err := vs.ValidateVotingConfiguration(config, 10, 10)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "higher than")
}

func TestValidateVotingConfiguration_ThresholdsOutOfRange(t *testing.T) {
	vs, _, _, _ := newTestService()
	config := defaultConfig(uuid.New(), 2)
	config.QualityGoodThreshold = 1.5
	err := vs.ValidateVotingConfiguration(config, 10, 10)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "[0, 1] range")
}

func TestValidateVotingConfiguration_InsufficientTotalEvaluations(t *testing.T) {
	vs, _, _, _ := newTestService()
	// k=10, n=3: maxPossibleM=9, recommendedM=ceil(2*log2(10))=7, capped to
	// ceil(9*0.6)=6 since k<=10. m=6 satisfies both bounds.
	config := defaultConfig(uuid.New(), 6)
	config.MinEvaluationsPerFile = 10
	// n*m = 3*6 = 18, need k*MinEvaluationsPerFile = 10*10 = 100
	err := vs.ValidateVotingConfiguration(config, 10, 3)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient total evaluations")
}

func TestValidateVotingConfiguration_ValidConfigPasses(t *testing.T) {
	vs, _, _, _ := newTestService()
	config := defaultConfig(uuid.New(), 2)
	config.MinEvaluationsPerFile = 1
	err := vs.ValidateVotingConfiguration(config, 4, 4)
	assert.NoError(t, err)
}

// ---------------------------------------------------------------------------
// Vote / Assignment / VotingConfiguration domain validation (vote.go)
// ---------------------------------------------------------------------------

func TestVoteValidate(t *testing.T) {
	valid := NewVote(uuid.New(), uuid.New(), uuid.New(), 1)
	assert.NoError(t, valid.Validate())

	missingEvent := NewVote(uuid.Nil, uuid.New(), uuid.New(), 1)
	assert.Error(t, missingEvent.Validate())

	zeroRank := NewVote(uuid.New(), uuid.New(), uuid.New(), 0)
	assert.Error(t, zeroRank.Validate())

	negativeRank := NewVote(uuid.New(), uuid.New(), uuid.New(), -1)
	assert.Error(t, negativeRank.Validate())
}

func TestAssignmentValidateAndCompletion(t *testing.T) {
	a := NewAssignment(uuid.New(), uuid.New(), []uuid.UUID{uuid.New()})
	assert.NoError(t, a.Validate())
	assert.False(t, a.IsCompleted)
	assert.Nil(t, a.CompletedAt)

	a.MarkCompleted()
	assert.True(t, a.IsCompleted)
	assert.NotNil(t, a.CompletedAt)

	empty := NewAssignment(uuid.New(), uuid.New(), nil)
	assert.Error(t, empty.Validate())
}

func TestVotingConfigurationValidate(t *testing.T) {
	config := NewVotingConfiguration(uuid.New(), 3)
	assert.NoError(t, config.Validate())

	config.AttachmentsPerEvaluator = 0
	assert.Error(t, config.Validate())
}

// ---------------------------------------------------------------------------
// GenerateAssignments with n = k (evaluators are the participants with a proposal)
// ---------------------------------------------------------------------------

func TestGenerateAssignments_EvaluatorsAreProposalOwners(t *testing.T) {
	for _, k := range []int{3, 4, 10} {
		vs, _, attachmentRepo, _ := newTestService()
		eventID := uuid.New()

		participants := make([]uuid.UUID, k)
		attachments := make([]uuid.UUID, k)
		ownerOf := map[string]uuid.UUID{}
		for i := range participants {
			participants[i] = uuid.New()
			attachments[i] = uuid.New()
			ownerOf[attachments[i].String()] = participants[i]
			attachmentRepo.add(eventID, &fakeAttachment{id: attachments[i], participantID: participants[i]})
		}

		_, _, m := VotingBounds(k)
		config := defaultConfig(eventID, m)
		config.MinEvaluationsPerFile = min(3, m)

		assignments, err := vs.GenerateAssignments(eventID, participants, attachments, config)
		require.NoError(t, err, "k=%d", k)
		require.Len(t, assignments, k)

		total := 0
		for _, a := range assignments {
			assert.Len(t, a.AttachmentIDs, m, "k=%d", k)
			total += len(a.AttachmentIDs)
			for _, id := range a.AttachmentIDs {
				assert.NotEqual(t, a.ParticipantID, ownerOf[id], "k=%d: own proposal assigned", k)
			}
		}
		assert.Equal(t, k*m, total, "k=%d", k)
	}
}

// ---------------------------------------------------------------------------
// applyIncentiveSystem: exact shift of n places relative to neutral evaluators
// (audit finding C5: the old version moved n−1 places or fewer)
// ---------------------------------------------------------------------------

// incentiveTable builds a G of size k where every owner is a neutral evaluator (Q=0.45), and
// returns it with the qualities map so a test can turn some owners good or bad.
func incentiveTable(k int) ([]AttachmentResult, map[string]float64) {
	results := make([]AttachmentResult, k)
	qualities := make(map[string]float64, k)
	for i := range results {
		owner := uuid.New()
		results[i] = AttachmentResult{AttachmentID: uuid.New(), ParticipantID: owner, GlobalRank: i + 1}
		qualities[owner.String()] = 0.45
	}
	return results, qualities
}

func positionOf(adjusted []AttachmentResult, id uuid.UUID) int {
	for _, r := range adjusted {
		if r.AttachmentID == id {
			return r.AdjustedRank
		}
	}
	return 0
}

func TestApplyIncentiveSystem_GoodEvaluatorMovesUpExactlyN(t *testing.T) {
	vs, _, _, _ := newTestService()
	config := defaultConfig(uuid.New(), 2)
	for _, n := range []int{1, 2, 3, 5} {
		config.AdjustmentMagnitude = n
		results, qualities := incentiveTable(10)
		qualities[results[6].ParticipantID.String()] = 0.9 // 7th in G

		adjusted := vs.applyIncentiveSystem(results, qualities, nil, config)

		assert.Equal(t, 7-n, positionOf(adjusted, results[6].AttachmentID), "n=%d", n)
		assert.Equal(t, n, adjusted[7-n-1].Adjustment)
		assert.Equal(t, AdjustmentGoodEvaluator, adjusted[7-n-1].AdjustmentReason)
	}
}

func TestApplyIncentiveSystem_BadEvaluatorMovesDownExactlyN(t *testing.T) {
	vs, _, _, _ := newTestService()
	config := defaultConfig(uuid.New(), 2)
	for _, n := range []int{1, 2, 3, 5} {
		config.AdjustmentMagnitude = n
		results, qualities := incentiveTable(10)
		qualities[results[1].ParticipantID.String()] = 0.1 // 2nd in G

		adjusted := vs.applyIncentiveSystem(results, qualities, nil, config)

		assert.Equal(t, 2+n, positionOf(adjusted, results[1].AttachmentID), "n=%d", n)
		assert.Equal(t, -n, adjusted[2+n-1].Adjustment)
		assert.Equal(t, AdjustmentBadEvaluator, adjusted[2+n-1].AdjustmentReason)
	}
}

func TestApplyIncentiveSystem_NeutralOrderIsKept(t *testing.T) {
	vs, _, _, _ := newTestService()
	config := defaultConfig(uuid.New(), 2)
	results, qualities := incentiveTable(10)
	qualities[results[6].ParticipantID.String()] = 0.9
	qualities[results[1].ParticipantID.String()] = 0.1

	adjusted := vs.applyIncentiveSystem(results, qualities, nil, config)

	// The neutral proposals keep their relative order: they are the frame.
	last := 0
	for _, r := range adjusted {
		if r.Adjustment != 0 {
			continue
		}
		assert.Greater(t, r.GlobalRank, last, "neutral proposals must not swap among themselves")
		last = r.GlobalRank
		assert.Empty(t, r.AdjustmentReason)
	}
}

func TestApplyIncentiveSystem_OnlyTheEndsCutTheShiftShort(t *testing.T) {
	vs, _, _, _ := newTestService()
	config := defaultConfig(uuid.New(), 2) // n = 3
	results, qualities := incentiveTable(6)
	qualities[results[1].ParticipantID.String()] = 0.9 // 2nd, good: can only reach 1st
	qualities[results[4].ParticipantID.String()] = 0.1 // 5th, bad: can only reach 6th

	adjusted := vs.applyIncentiveSystem(results, qualities, nil, config)

	assert.Equal(t, 1, positionOf(adjusted, results[1].AttachmentID))
	assert.Equal(t, 6, positionOf(adjusted, results[4].AttachmentID))
}

func TestApplyIncentiveSystem_AllGoodKeepsTheOrder(t *testing.T) {
	vs, _, _, _ := newTestService()
	config := defaultConfig(uuid.New(), 2)
	results, qualities := incentiveTable(8)
	for owner := range qualities {
		qualities[owner] = 0.9
	}

	adjusted := vs.applyIncentiveSystem(results, qualities, nil, config)

	for i, r := range adjusted {
		assert.Equal(t, i+1, r.GlobalRank, "everyone moving up by n is no change relative to each other")
	}
}

func TestApplyIncentiveSystem_PenaltyUsesItsOwnMagnitude(t *testing.T) {
	vs, _, _, _ := newTestService()
	config := defaultConfig(uuid.New(), 2)
	config.NonEvaluatorSanction = SanctionPenalty
	config.NonEvaluatorPenalty = 4
	results, qualities := incentiveTable(10)
	nonEvaluator := results[0].ParticipantID
	delete(qualities, nonEvaluator.String())

	adjusted := vs.applyIncentiveSystem(results, qualities, map[uuid.UUID]bool{nonEvaluator: true}, config)

	assert.Equal(t, 5, positionOf(adjusted, results[0].AttachmentID))
	assert.Equal(t, -4, adjusted[4].Adjustment)
	assert.Equal(t, AdjustmentNonEvaluator, adjusted[4].AdjustmentReason)
}

func TestAttachmentResult_WithoutEvaluatorDetail(t *testing.T) {
	r := AttachmentResult{Adjustment: 3, AdjustmentReason: AdjustmentGoodEvaluator, NonEvaluator: true, Excluded: true, AdjustedRank: 2}
	clean := r.WithoutEvaluatorDetail()
	assert.Zero(t, clean.Adjustment)
	assert.Empty(t, clean.AdjustmentReason)
	assert.False(t, clean.NonEvaluator)
	assert.True(t, clean.Excluded, "exclusion is public: the proposal is out of the ranking")
	assert.Equal(t, 2, clean.AdjustedRank)
}
