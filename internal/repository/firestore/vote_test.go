package firestore_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/GolangUA/telegram-butler/internal/entity"
	"github.com/GolangUA/telegram-butler/internal/repository/firestore"
)

const (
	testProject = "telegram-butler-test"
	testChatID  = int64(-100)
)

// TestMain starts a fresh Firestore emulator container for the whole
// package, points the SDK at it via FIRESTORE_EMULATOR_HOST, then tears
// it down after the tests. Requires Docker (OrbStack / Docker Desktop)
// to be running. No manual gcloud / docker setup needed.
//
// Uses panic for setup errors and a bare m.Run() at the end — explicit
// log.Fatalf / os.Exit are flagged by revive's redundant-test-main-exit
// (Go 1.15+ test runner exits with m.Run()'s code automatically).
func TestMain(m *testing.M) {
	ctx := context.Background()

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "mtlynch/firestore-emulator:latest",
			ExposedPorts: []string{"8080/tcp"},
			WaitingFor:   wait.ForLog("Dev App Server is now running."),
		},
		Started: true,
	})
	if err != nil {
		panic(fmt.Errorf("start firestore emulator: %w", err))
	}

	defer func() { _ = container.Terminate(ctx) }()

	port, err := container.MappedPort(ctx, "8080")
	if err != nil {
		panic(fmt.Errorf("get mapped port: %w", err))
	}

	err = os.Setenv("FIRESTORE_EMULATOR_HOST", "localhost:"+port.Port())
	if err != nil {
		panic(fmt.Errorf("set FIRESTORE_EMULATOR_HOST: %w", err))
	}

	m.Run()
}

// newTestRepo returns a fresh VoteRepository scoped to a unique collection
// per test (so tests don't pollute each other).
func newTestRepo(t *testing.T) *firestore.VoteRepository {
	t.Helper()

	ctx := context.Background()

	client, err := firestore.NewClient(ctx, testProject)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	t.Cleanup(func() { _ = client.Close() })

	collection := fmt.Sprintf("test_votes_%d_%s", time.Now().UnixNano(), t.Name())

	return firestore.NewVoteRepository(client, collection)
}

func sampleVote(targetID int64, now time.Time) *entity.Vote {
	return &entity.Vote{
		ChatID:       testChatID,
		TargetUserID: targetID,
		TargetName:   "Target User",
		ReporterID:   100,
		MessageID:    42,
		ThreadID:     0,
		Voters: []entity.Voter{
			{ID: 100, FirstName: "Reporter", Username: "reporter"},
		},
		Status:    entity.VoteStatusActive,
		CreatedAt: now,
		ExpiresAt: now.Add(3 * time.Minute),
	}
}

func TestVoteRepository_Create_Active(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	// Firestore Timestamp precision is microseconds; truncate to avoid
	// nano-precision drift if a later test asserts equality.
	now := time.Now().UTC().Truncate(time.Microsecond)
	vote := sampleVote(200, now)

	err := repo.Create(ctx, vote)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.Active(ctx, entity.VoteKey{ChatID: testChatID, TargetUserID: 200})
	if err != nil {
		t.Fatalf("Active: %v", err)
	}

	if got.ChatID != -100 || got.TargetUserID != 200 {
		t.Errorf("ChatID/TargetUserID = %d/%d, want -100/200", got.ChatID, got.TargetUserID)
	}

	if got.TargetName != "Target User" {
		t.Errorf("TargetName = %q, want %q", got.TargetName, "Target User")
	}

	if len(got.Voters) != 1 || got.Voters[0].ID != 100 {
		t.Errorf("Voters = %+v, want 1 voter with ID 100", got.Voters)
	}

	if got.Status != entity.VoteStatusActive {
		t.Errorf("Status = %q, want %q", got.Status, entity.VoteStatusActive)
	}
}

func TestVoteRepository_Active_NotFound(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	_, err := repo.Active(ctx, entity.VoteKey{ChatID: testChatID, TargetUserID: 888})
	if !errors.Is(err, entity.ErrVoteNotFound) {
		t.Errorf("err = %v, want ErrVoteNotFound", err)
	}
}

func TestVoteRepository_AddVoter(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Microsecond)
	vote := sampleVote(200, now)

	err := repo.Create(ctx, vote)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	updated, err := repo.AddVoter(ctx, entity.VoteKey{ChatID: testChatID, TargetUserID: 200}, entity.Voter{
		ID: 300, FirstName: "Second", Username: "second_voter",
	})
	if err != nil {
		t.Fatalf("AddVoter: %v", err)
	}

	if len(updated.Voters) != 2 {
		t.Fatalf("Voters len = %d, want 2", len(updated.Voters))
	}

	// Verify persistence: re-read independently.
	got, err := repo.Active(ctx, entity.VoteKey{ChatID: testChatID, TargetUserID: 200})
	if err != nil {
		t.Fatalf("Active after AddVoter: %v", err)
	}

	if len(got.Voters) != 2 {
		t.Fatalf("persisted Voters len = %d, want 2", len(got.Voters))
	}

	if got.Voters[1].ID != 300 {
		t.Errorf("second voter ID = %d, want 300", got.Voters[1].ID)
	}
}

func TestVoteRepository_SetStatus_RemovesFromActive(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Microsecond)
	vote := sampleVote(200, now)

	err := repo.Create(ctx, vote)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	err = repo.SetStatus(ctx, entity.VoteKey{ChatID: testChatID, TargetUserID: 200}, entity.VoteStatusMuted)
	if err != nil {
		t.Fatalf("SetStatus: %v", err)
	}

	_, err = repo.Active(ctx, entity.VoteKey{ChatID: testChatID, TargetUserID: 200})
	if !errors.Is(err, entity.ErrVoteNotFound) {
		t.Errorf("Active after SetStatus = %v, want ErrVoteNotFound", err)
	}
}

func TestVoteRepository_ListActive(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Microsecond)

	// Two active votes for different targets + one that will be muted.
	active1 := sampleVote(200, now)
	active2 := sampleVote(201, now.Add(time.Microsecond))
	muted := sampleVote(202, now.Add(2*time.Microsecond))

	for _, v := range []*entity.Vote{active1, active2, muted} {
		err := repo.Create(ctx, v)
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	err := repo.SetStatus(ctx, entity.VoteKey{ChatID: testChatID, TargetUserID: 202}, entity.VoteStatusMuted)
	if err != nil {
		t.Fatalf("SetStatus: %v", err)
	}

	list, err := repo.ListActive(ctx)
	if err != nil {
		t.Fatalf("ListActive: %v", err)
	}

	if len(list) != 2 {
		t.Errorf("ListActive len = %d, want 2", len(list))
	}
}
