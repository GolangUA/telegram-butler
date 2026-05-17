package firestore

import (
	"context"
	"errors"
	"fmt"
	"time"

	fsdk "cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"

	"github.com/GolangUA/telegram-butler/internal/entity"
)

// DefaultCollection is the Firestore collection that hosts vote documents.
const DefaultCollection = "report_votes"

// firestoreOpTimeout bounds every Firestore RPC. Without this, a single
// /report (or coordinator goroutine vote) would hang for ~10 minutes when
// the backend is unreachable, blocking the handler and producing a silent
// UX. With it, requests fail fast and the caller sees errReportInfra.
const firestoreOpTimeout = 5 * time.Second

// VoteRepository persists community-vote state in Firestore.
type VoteRepository struct {
	col *fsdk.CollectionRef
}

// NewVoteRepository builds a repository scoped to a collection.
// Pass DefaultCollection unless tests override it.
func NewVoteRepository(client *Client, collection string) *VoteRepository {
	return &VoteRepository{col: client.Collection(collection)}
}

// Create persists a fresh vote. The doc ID is composite so history is
// preserved across re-reports of the same target. Active-vote uniqueness
// is enforced at the handler layer (validateReport).
func (r *VoteRepository) Create(ctx context.Context, vote *entity.Vote) error {
	ctx, cancel := context.WithTimeout(ctx, firestoreOpTimeout)
	defer cancel()

	id := buildDocID(vote.ChatID, vote.TargetUserID, vote.CreatedAt)

	_, err := r.col.Doc(id).Set(ctx, toDoc(vote))
	if err != nil {
		return fmt.Errorf("firestore: Create vote %s: %w", id, err)
	}

	return nil
}

// GetActive returns the single active vote for a target, or
// entity.ErrVoteNotFound. The query relies on a composite index on
// (chat_id, target_user_id, status).
func (r *VoteRepository) GetActive(ctx context.Context, key entity.VoteKey) (*entity.Vote, error) {
	ctx, cancel := context.WithTimeout(ctx, firestoreOpTimeout)
	defer cancel()

	snap, err := r.findActive(ctx, key)
	if err != nil {
		return nil, err
	}

	return decode(snap)
}

// AddVoter appends a voter to the active vote and returns the updated state.
// Per-vote serialization lives in the coordinator goroutine, so no
// transaction is needed here.
func (r *VoteRepository) AddVoter(ctx context.Context, key entity.VoteKey, voter entity.Voter) (*entity.Vote, error) {
	ctx, cancel := context.WithTimeout(ctx, firestoreOpTimeout)
	defer cancel()

	snap, err := r.findActive(ctx, key)
	if err != nil {
		return nil, err
	}

	var doc voteDoc

	err = snap.DataTo(&doc)
	if err != nil {
		return nil, fmt.Errorf("firestore: AddVoter decode: %w", err)
	}

	doc.Voters = append(doc.Voters, toVoterDoc(voter))

	_, err = snap.Ref.Update(ctx, []fsdk.Update{
		{Path: "voters", Value: doc.Voters},
	})
	if err != nil {
		return nil, fmt.Errorf("firestore: AddVoter update: %w", err)
	}

	return fromDoc(doc), nil
}

// SetStatus moves the active vote to a terminal state (muted / expired).
func (r *VoteRepository) SetStatus(ctx context.Context, key entity.VoteKey, status string) error {
	ctx, cancel := context.WithTimeout(ctx, firestoreOpTimeout)
	defer cancel()

	snap, err := r.findActive(ctx, key)
	if err != nil {
		return err
	}

	_, err = snap.Ref.Update(ctx, []fsdk.Update{
		{Path: "status", Value: status},
	})
	if err != nil {
		return fmt.Errorf("firestore: SetStatus: %w", err)
	}

	return nil
}

// ListActive returns every vote in the active state. Used at startup to
// reconcile in-flight votes after a bot restart.
func (r *VoteRepository) ListActive(ctx context.Context) ([]*entity.Vote, error) {
	ctx, cancel := context.WithTimeout(ctx, firestoreOpTimeout)
	defer cancel()

	iter := r.col.
		Where("status", "==", entity.VoteStatusActive).
		Documents(ctx)
	defer iter.Stop()

	var out []*entity.Vote

	for {
		snap, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("firestore: ListActive: %w", err)
		}

		vote, err := decode(snap)
		if err != nil {
			return nil, err
		}

		out = append(out, vote)
	}

	return out, nil
}

func (r *VoteRepository) findActive(ctx context.Context, key entity.VoteKey) (*fsdk.DocumentSnapshot, error) {
	iter := r.col.
		Where("chat_id", "==", key.ChatID).
		Where("target_user_id", "==", key.TargetUserID).
		Where("status", "==", entity.VoteStatusActive).
		Limit(1).
		Documents(ctx)
	defer iter.Stop()

	snap, err := iter.Next()
	if errors.Is(err, iterator.Done) {
		return nil, entity.ErrVoteNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("firestore: findActive: %w", err)
	}

	return snap, nil
}

func buildDocID(chatID, targetUserID int64, createdAt time.Time) string {
	return fmt.Sprintf("%d_%d_%d", chatID, targetUserID, createdAt.UnixNano())
}

func decode(snap *fsdk.DocumentSnapshot) (*entity.Vote, error) {
	var doc voteDoc

	err := snap.DataTo(&doc)
	if err != nil {
		return nil, fmt.Errorf("firestore: decode vote: %w", err)
	}

	return fromDoc(doc), nil
}

type voteDoc struct {
	ChatID       int64      `firestore:"chat_id"`
	TargetUserID int64      `firestore:"target_user_id"`
	TargetName   string     `firestore:"target_name"`
	ReporterID   int64      `firestore:"reporter_id"`
	MessageID    int        `firestore:"message_id"`
	ThreadID     int        `firestore:"thread_id"`
	Voters       []voterDoc `firestore:"voters"`
	Status       string     `firestore:"status"`
	CreatedAt    time.Time  `firestore:"created_at"`
	ExpiresAt    time.Time  `firestore:"expires_at"`
}

type voterDoc struct {
	ID        int64  `firestore:"id"`
	FirstName string `firestore:"first_name"`
	Username  string `firestore:"username"`
}

func toDoc(v *entity.Vote) voteDoc {
	voters := make([]voterDoc, 0, len(v.Voters))
	for _, voter := range v.Voters {
		voters = append(voters, toVoterDoc(voter))
	}

	return voteDoc{
		ChatID:       v.ChatID,
		TargetUserID: v.TargetUserID,
		TargetName:   v.TargetName,
		ReporterID:   v.ReporterID,
		MessageID:    v.MessageID,
		ThreadID:     v.ThreadID,
		Voters:       voters,
		Status:       v.Status,
		CreatedAt:    v.CreatedAt,
		ExpiresAt:    v.ExpiresAt,
	}
}

func fromDoc(d voteDoc) *entity.Vote {
	voters := make([]entity.Voter, 0, len(d.Voters))
	for _, voter := range d.Voters {
		voters = append(voters, entity.Voter{
			ID:        voter.ID,
			FirstName: voter.FirstName,
			Username:  voter.Username,
		})
	}

	return &entity.Vote{
		ChatID:       d.ChatID,
		TargetUserID: d.TargetUserID,
		TargetName:   d.TargetName,
		ReporterID:   d.ReporterID,
		MessageID:    d.MessageID,
		ThreadID:     d.ThreadID,
		Voters:       voters,
		Status:       d.Status,
		CreatedAt:    d.CreatedAt,
		ExpiresAt:    d.ExpiresAt,
	}
}

func toVoterDoc(v entity.Voter) voterDoc {
	return voterDoc{
		ID:        v.ID,
		FirstName: v.FirstName,
		Username:  v.Username,
	}
}
