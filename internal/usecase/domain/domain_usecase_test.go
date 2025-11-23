package domain

import (
	"context"
	"testing"
	"time"

	"assigning-reviewers-for-pr/internal/entities"
	"assigning-reviewers-for-pr/internal/repository"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type repoMock struct{ mock.Mock }

var _ repository.Repository = (*repoMock)(nil)

func (m *repoMock) OnStart(_ context.Context) error { return nil }
func (m *repoMock) OnStop(_ context.Context) error  { return nil }

func (m *repoMock) CreatePR(ctx context.Context, pr entities.PullRequest) (*entities.PullRequest, error) {
	args := m.Called(ctx, pr)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.PullRequest), args.Error(1)
}

func (m *repoMock) MergePR(ctx context.Context, prID string) (*entities.PullRequest, error) {
	args := m.Called(ctx, prID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.PullRequest), args.Error(1)
}

func (m *repoMock) ReassignReviewer(ctx context.Context, prID, oldUserID string) (*entities.PullRequest, string, error) {
	args := m.Called(ctx, prID, oldUserID)
	var pr *entities.PullRequest
	if args.Get(0) != nil {
		pr = args.Get(0).(*entities.PullRequest)
	}
	repl := args.String(1)
	return pr, repl, args.Error(2)
}

func (m *repoMock) SetUserActive(ctx context.Context, userID string, isActive bool) (*entities.User, error) {
	args := m.Called(ctx, userID, isActive)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *repoMock) GetUserReviews(ctx context.Context, userID string) ([]entities.PullRequestShort, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entities.PullRequestShort), args.Error(1)
}

func (m *repoMock) CreateTeam(ctx context.Context, team entities.Team) (*entities.Team, error) {
	args := m.Called(ctx, team)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Team), args.Error(1)
}

func (m *repoMock) GetTeam(ctx context.Context, name string) (*entities.Team, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Team), args.Error(1)
}

func (m *repoMock) Stats(ctx context.Context) (entities.Stats, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return entities.Stats{}, args.Error(1)
	}
	return args.Get(0).(entities.Stats), args.Error(1)
}

func (m *repoMock) StatsSummary(ctx context.Context, filter entities.StatsFilter) (entities.StatsSummary, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return entities.StatsSummary{}, args.Error(1)
	}
	return args.Get(0).(entities.StatsSummary), args.Error(1)
}

func (m *repoMock) ReviewerStats(ctx context.Context, userID string, limit int) (entities.ReviewerStats, error) {
	args := m.Called(ctx, userID, limit)
	if args.Get(0) == nil {
		return entities.ReviewerStats{}, args.Error(1)
	}
	return args.Get(0).(entities.ReviewerStats), args.Error(1)
}

func (m *repoMock) PRStats(ctx context.Context, prID string) (entities.PRStats, error) {
	args := m.Called(ctx, prID)
	if args.Get(0) == nil {
		return entities.PRStats{}, args.Error(1)
	}
	return args.Get(0).(entities.PRStats), args.Error(1)
}

func (m *repoMock) DeactivateTeam(ctx context.Context, teamName string) (entities.DeactivateResult, error) {
	args := m.Called(ctx, teamName)
	if args.Get(0) == nil {
		return entities.DeactivateResult{}, args.Error(1)
	}
	return args.Get(0).(entities.DeactivateResult), args.Error(1)
}

func TestUsecase_CreatePullRequestValidation(t *testing.T) {
	repo := &repoMock{}
	uc := New(zap.NewNop().Sugar(), context.Background(), repo, time.Second)

	_, err := uc.CreatePullRequest(context.Background(), entities.PullRequest{})
	require.ErrorIs(t, err, entities.ErrInvalidArgument)
	repo.AssertNotCalled(t, "CreatePR", mock.Anything, mock.Anything)
}

func TestUsecase_CreatePullRequestDelegates(t *testing.T) {
	repo := &repoMock{}
	uc := New(zap.NewNop().Sugar(), context.Background(), repo, time.Second)

	expected := &entities.PullRequest{ID: "1", Name: "demo", AuthorID: "a1"}
	repo.On("CreatePR", mock.Anything, mock.MatchedBy(func(pr entities.PullRequest) bool {
		return pr.ID == expected.ID
	})).Return(expected, nil)

	pr, err := uc.CreatePullRequest(context.Background(), entities.PullRequest{ID: "1", Name: "demo", AuthorID: "a1"})
	require.NoError(t, err)
	require.Equal(t, expected, pr)
	repo.AssertExpectations(t)
}

func TestUsecase_SetActiveUserValidation(t *testing.T) {
	repo := &repoMock{}
	uc := New(zap.NewNop().Sugar(), context.Background(), repo, time.Second)

	_, err := uc.SetActiveUser(context.Background(), "", true)
	require.ErrorIs(t, err, entities.ErrInvalidArgument)
}

func TestUsecase_TeamGetValidation(t *testing.T) {
	repo := &repoMock{}
	uc := New(zap.NewNop().Sugar(), context.Background(), repo, time.Second)

	_, err := uc.Team(context.Background(), "")
	require.ErrorIs(t, err, entities.ErrInvalidArgument)
}

func TestUsecase_ReviewerStatsValidation(t *testing.T) {
	repo := &repoMock{}
	uc := New(zap.NewNop().Sugar(), context.Background(), repo, time.Second)

	_, err := uc.ReviewerStats(context.Background(), "", 0)
	require.ErrorIs(t, err, entities.ErrInvalidArgument)
}

func TestUsecase_DeactivateValidation(t *testing.T) {
	repo := &repoMock{}
	uc := New(zap.NewNop().Sugar(), context.Background(), repo, time.Second)

	_, err := uc.DeactivateTeam(context.Background(), "")
	require.ErrorIs(t, err, entities.ErrInvalidArgument)
}
