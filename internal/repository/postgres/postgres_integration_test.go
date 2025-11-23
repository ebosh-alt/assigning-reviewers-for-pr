package postgres

import (
	"context"
	"database/sql"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"assigning-reviewers-for-pr/config"
	"assigning-reviewers-for-pr/internal/entities"

	_ "github.com/lib/pq"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestRepositoryIntegration(t *testing.T) {
	ctx := context.Background()

	cfg, cleanup := setupPostgres(t)
	t.Cleanup(cleanup)

	repo := New(ctx, testLogger(t), cfg)
	require.NoError(t, repo.OnStart(ctx))
	t.Cleanup(func() { _ = repo.OnStop(ctx) })

	team := entities.Team{Name: "backend", Members: []entities.User{
		{ID: "u1", Username: "Alice", IsActive: true},
		{ID: "u2", Username: "Bob", IsActive: true},
		{ID: "u3", Username: "Charlie", IsActive: true},
		{ID: "u4", Username: "Dana", IsActive: true},
	}}

	createdTeam, err := repo.CreateTeam(ctx, team)
	require.NoError(t, err)
	require.Len(t, createdTeam.Members, 4)

	fetched, err := repo.GetTeam(ctx, team.Name)
	require.NoError(t, err)
	require.Equal(t, team.Name, fetched.Name)

	pr, err := repo.CreatePR(ctx, entities.PullRequest{ID: "pr1", Name: "Init", AuthorID: "u1"})
	require.NoError(t, err)
	require.Equal(t, entities.StatusOpen, pr.Status)
	require.Len(t, pr.Reviewers, 2)
	require.NotContains(t, pr.Reviewers, "u1")

	old := pr.Reviewers[0]
	reassigned, repl, err := repo.ReassignReviewer(ctx, pr.ID, old)
	require.NoError(t, err)
	require.NotEqual(t, old, repl)
	require.Contains(t, reassigned.Reviewers, repl)
	require.NotContains(t, reassigned.Reviewers, old)

	prs, err := repo.GetUserReviews(ctx, repl)
	require.NoError(t, err)
	require.NotEmpty(t, prs)

	merged, err := repo.MergePR(ctx, pr.ID)
	require.NoError(t, err)
	require.Equal(t, entities.StatusMerged, merged.Status)
	require.NotNil(t, merged.MergedAt)

	merged2, err := repo.MergePR(ctx, pr.ID)
	require.NoError(t, err)
	require.Equal(t, merged.MergedAt, merged2.MergedAt)

	_, _, err = repo.ReassignReviewer(ctx, pr.ID, repl)
	require.ErrorIs(t, err, entities.ErrPRMerged)

	prStats, err := repo.PRStats(ctx, pr.ID)
	require.NoError(t, err)
	require.Equal(t, int64(1), prStats.TransferCount)
	require.NotEmpty(t, prStats.Reassignments)
	require.Equal(t, old, prStats.Reassignments[0].OldReviewerID)
	require.NotNil(t, prStats.Reassignments[0].NewReviewerID)

	updated, err := repo.SetUserActive(ctx, "u2", false)
	require.NoError(t, err)
	require.False(t, updated.IsActive)
}

func setupPostgres(t *testing.T) (*config.Config, func()) {
	t.Helper()

	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "16-alpine",
		Env: []string{
			"POSTGRES_PASSWORD=postgres",
			"POSTGRES_USER=postgres",
			"POSTGRES_DB=assigning_reviewers_for_pr_db",
		},
	}, func(hc *docker.HostConfig) {
		hc.AutoRemove = true
	})
	require.NoError(t, err)

	hostPort := resource.GetPort("5432/tcp")

	port, err := strconv.Atoi(hostPort)
	require.NoError(t, err)
	migrationsDir, err := filepath.Abs(filepath.Join("..", "..", "..", "db", "migrations"))
	require.NoError(t, err)
	require.DirExists(t, migrationsDir)

	cfg := &config.Config{
		Server: config.ServerConfig{Host: "0.0.0.0", Port: 8080, ShutdownTimeout: 5 * time.Second},
		HTTP:   config.HTTPConfig{RequestTimeout: 5 * time.Second},
		Postgres: config.PostgresConfig{
			Host:           "localhost",
			Port:           port,
			User:           "postgres",
			Password:       "postgres",
			DBName:         "assigning_reviewers_for_pr_db",
			SSLMode:        "disable",
			MigrationsDir:  migrationsDir,
			QueryTimeout:   10 * time.Second,
			MigrateTimeout: 20 * time.Second,
			MaxConns:       4,
			MinConns:       1,
		},
	}

	require.NoError(t, pool.Retry(func() error {
		db, err := sql.Open("postgres", "host=localhost port="+hostPort+" user=postgres password=postgres dbname=assigning_reviewers_for_pr_db sslmode=disable")
		if err != nil {
			return err
		}
		defer func() { _ = db.Close() }()
		return db.Ping()
	}))

	cleanup := func() {
		_ = pool.Purge(resource)
	}

	return cfg, cleanup
}

func testLogger(t *testing.T) *zap.SugaredLogger {
	t.Helper()

	l, _ := zap.NewDevelopment()
	t.Cleanup(func() { _ = l.Sync() })
	return l.Sugar()
}

func TestRepositoryStatsIntegration(t *testing.T) {
	ctx := context.Background()

	cfg, cleanup := setupPostgres(t)
	t.Cleanup(cleanup)

	repo := New(ctx, testLogger(t), cfg)
	require.NoError(t, repo.OnStart(ctx))
	t.Cleanup(func() { _ = repo.OnStop(ctx) })

	team := entities.Team{Name: "backend", Members: []entities.User{
		{ID: "u1", Username: "Alice", IsActive: true},
		{ID: "u2", Username: "Bob", IsActive: true},
		{ID: "u3", Username: "Charlie", IsActive: true},
		{ID: "u4", Username: "Dana", IsActive: true},
	}}

	_, err := repo.CreateTeam(ctx, team)
	require.NoError(t, err)

	pr1, err := repo.CreatePR(ctx, entities.PullRequest{ID: "pr1", Name: "Init", AuthorID: "u1"})
	require.NoError(t, err)
	pr2, err := repo.CreatePR(ctx, entities.PullRequest{ID: "pr2", Name: "Feature", AuthorID: "u1"})
	require.NoError(t, err)

	stats, err := repo.Stats(ctx)
	require.NoError(t, err)

	totalAssignments := len(pr1.Reviewers) + len(pr2.Reviewers)

	prCounts := map[string]int64{}
	for _, s := range stats.ByPR {
		prCounts[s.PRID] = s.AssignCnt
	}

	require.Equal(t, int64(len(pr1.Reviewers)), prCounts[pr1.ID])
	require.Equal(t, int64(len(pr2.Reviewers)), prCounts[pr2.ID])

	var userSum int64
	for _, s := range stats.ByUser {
		userSum += s.AssignCnt
	}
	require.Equal(t, int64(totalAssignments), userSum)

	statusCounts := map[entities.PullRequestStatus]int64{}
	for _, s := range stats.ByStatus {
		statusCounts[s.Status] = s.PRCount
	}
	require.Equal(t, int64(2), statusCounts[entities.StatusOpen])

	teamCounts := map[string]int64{}
	for _, s := range stats.ByTeam {
		teamCounts[s.TeamName] = s.AssignCnt
	}
	require.Equal(t, int64(totalAssignments), teamCounts[team.Name])
}

func TestRepositoryStatsSummary(t *testing.T) {
	ctx := context.Background()

	cfg, cleanup := setupPostgres(t)
	t.Cleanup(cleanup)

	repo := New(ctx, testLogger(t), cfg)
	require.NoError(t, repo.OnStart(ctx))
	t.Cleanup(func() { _ = repo.OnStop(ctx) })

	team := entities.Team{Name: "backend", Members: []entities.User{
		{ID: "u1", Username: "Alice", IsActive: true},
		{ID: "u2", Username: "Bob", IsActive: true},
		{ID: "u3", Username: "Charlie", IsActive: true},
	}}

	_, err := repo.CreateTeam(ctx, team)
	require.NoError(t, err)

	_, err = repo.CreatePR(ctx, entities.PullRequest{ID: "pr1", Name: "Init", AuthorID: "u1"})
	require.NoError(t, err)
	_, err = repo.CreatePR(ctx, entities.PullRequest{ID: "pr2", Name: "Feature", AuthorID: "u1"})
	require.NoError(t, err)

	summary, err := repo.StatsSummary(ctx, entities.StatsFilter{Limit: 5})
	require.NoError(t, err)
	require.NotEmpty(t, summary.TopReviewers)
	require.NotEmpty(t, summary.TeamAssignments)

	var statusSum int64
	for _, s := range summary.PRStatusCounts {
		statusSum += s.PRCount
	}
	require.Equal(t, int64(2), statusSum)

	reviewerID := summary.TopReviewers[0].UserID
	reviewerStats, err := repo.ReviewerStats(ctx, reviewerID, 3)
	require.NoError(t, err)
	require.Equal(t, reviewerID, reviewerStats.UserID)
	require.NotEmpty(t, reviewerStats.RecentPRs)
}

func TestMergeIdempotentIntegration(t *testing.T) {
	ctx := context.Background()

	cfg, cleanup := setupPostgres(t)
	t.Cleanup(cleanup)

	repo := New(ctx, testLogger(t), cfg)
	require.NoError(t, repo.OnStart(ctx))
	t.Cleanup(func() { _ = repo.OnStop(ctx) })

	team := entities.Team{Name: "backend", Members: []entities.User{{ID: "u1", Username: "Alice", IsActive: true}}}
	_, err := repo.CreateTeam(ctx, team)
	require.NoError(t, err)

	pr, err := repo.CreatePR(ctx, entities.PullRequest{ID: "pr-merge", Name: "Merge", AuthorID: "u1"})
	require.NoError(t, err)
	require.Equal(t, entities.StatusOpen, pr.Status)

	m1, err := repo.MergePR(ctx, pr.ID)
	require.NoError(t, err)
	require.Equal(t, entities.StatusMerged, m1.Status)
	require.NotNil(t, m1.MergedAt)

	m2, err := repo.MergePR(ctx, pr.ID)
	require.NoError(t, err)
	require.Equal(t, entities.StatusMerged, m2.Status)
	require.Equal(t, m1.MergedAt, m2.MergedAt)
}

func TestDeactivateTeamReassignIntegration(t *testing.T) {
	ctx := context.Background()

	cfg, cleanup := setupPostgres(t)
	t.Cleanup(cleanup)

	repo := New(ctx, testLogger(t), cfg)
	require.NoError(t, repo.OnStart(ctx))
	t.Cleanup(func() { _ = repo.OnStop(ctx) })

	teamA := entities.Team{Name: "backend", Members: []entities.User{
		{ID: "u1", Username: "Alice", IsActive: true},
		{ID: "u2", Username: "Bob", IsActive: true},
	}}
	teamB := entities.Team{Name: "frontend", Members: []entities.User{
		{ID: "u3", Username: "Charlie", IsActive: true},
		{ID: "u4", Username: "Dana", IsActive: true},
	}}

	_, err := repo.CreateTeam(ctx, teamA)
	require.NoError(t, err)
	_, err = repo.CreateTeam(ctx, teamB)
	require.NoError(t, err)

	pr, err := repo.CreatePR(ctx, entities.PullRequest{ID: "pr-deact", Name: "Test", AuthorID: "u1"})
	require.NoError(t, err)
	initialCount := len(pr.Reviewers)
	require.NotZero(t, initialCount)

	result, err := repo.DeactivateTeam(ctx, "backend")
	require.NoError(t, err)
	require.Equal(t, 2, result.DeactivatedUsers)
	require.Equal(t, initialCount, result.Reassigned+result.Removed)

	updated, err := repo.MergePR(ctx, pr.ID) // to read current reviewers
	require.NoError(t, err)
	for _, r := range updated.Reviewers {
		require.NotContains(t, []string{"u1", "u2"}, r)
	}
	require.Len(t, updated.Reviewers, initialCount-result.Removed)
}
