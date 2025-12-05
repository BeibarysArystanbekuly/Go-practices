package api

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"polling-system/internal/domain/poll"
	"polling-system/internal/domain/user"
	"polling-system/internal/domain/vote"
	jwtpkg "polling-system/internal/platform/jwt"
	"polling-system/internal/worker"
)

type testUserRepo struct {
	mu     sync.Mutex
	users  map[int64]*user.User
	byMail map[string]int64
	nextID int64
}

func newTestUserRepo() *testUserRepo {
	return &testUserRepo{
		users:  make(map[int64]*user.User),
		byMail: make(map[string]int64),
		nextID: 1,
	}
}

func (r *testUserRepo) seed(u *user.User) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if u.ID == 0 {
		u.ID = r.nextID
		r.nextID++
	}
	if u.CreatedAt.IsZero() {
		u.CreatedAt = time.Now()
	}
	copyUser := *u
	r.users[u.ID] = &copyUser
	r.byMail[u.Email] = u.ID
}

func (r *testUserRepo) Create(ctx context.Context, u *user.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	u.ID = r.nextID
	r.nextID++
	u.CreatedAt = time.Now()
	copyUser := *u
	r.users[u.ID] = &copyUser
	r.byMail[u.Email] = u.ID
	return nil
}

func (r *testUserRepo) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	id, ok := r.byMail[email]
	if !ok {
		return nil, sql.ErrNoRows
	}
	copyUser := *r.users[id]
	return &copyUser, nil
}

func (r *testUserRepo) GetByID(ctx context.Context, id int64) (*user.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	u, ok := r.users[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	copyUser := *u
	return &copyUser, nil
}

func (r *testUserRepo) List(ctx context.Context) ([]user.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	res := make([]user.User, 0, len(r.users))
	for _, u := range r.users {
		res = append(res, *u)
	}
	return res, nil
}

func (r *testUserRepo) UpdateRole(ctx context.Context, id int64, role string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	u, ok := r.users[id]
	if !ok {
		return sql.ErrNoRows
	}
	u.Role = role
	return nil
}

type testPollRepo struct {
	mu     sync.Mutex
	polls  map[int64]*poll.Poll
	opts   map[int64][]poll.Option
	nextID int64
}

func newTestPollRepo() *testPollRepo {
	return &testPollRepo{
		polls:  make(map[int64]*poll.Poll),
		opts:   make(map[int64][]poll.Option),
		nextID: 1,
	}
}

func (r *testPollRepo) Create(ctx context.Context, p *poll.Poll, options []poll.Option) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	p.ID = r.nextID
	r.nextID++
	p.CreatedAt = time.Now()
	p.UpdatedAt = p.CreatedAt
	copyPoll := *p
	r.polls[p.ID] = &copyPoll

	cloned := make([]poll.Option, len(options))
	for i, opt := range options {
		opt.ID = int64(i + 1)
		opt.PollID = p.ID
		opt.CreatedAt = time.Now()
		cloned[i] = opt
	}
	r.opts[p.ID] = cloned
	return p.ID, nil
}

func (r *testPollRepo) GetByID(ctx context.Context, id int64) (*poll.Poll, []poll.Option, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.polls[id]
	if !ok {
		return nil, nil, sql.ErrNoRows
	}
	opts := r.opts[id]
	copyPoll := *p
	copiedOpts := make([]poll.Option, len(opts))
	copy(copiedOpts, opts)
	return &copyPoll, copiedOpts, nil
}

func (r *testPollRepo) List(ctx context.Context, status *string) ([]poll.Poll, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	res := []poll.Poll{}
	for _, p := range r.polls {
		if status != nil && p.Status != *status {
			continue
		}
		res = append(res, *p)
	}
	return res, nil
}

func (r *testPollRepo) UpdateStatus(ctx context.Context, id int64, status string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.polls[id]
	if !ok {
		return sql.ErrNoRows
	}
	p.Status = status
	p.UpdatedAt = time.Now()
	return nil
}

type noopVoteRepo struct{}

func (noopVoteRepo) Create(ctx context.Context, v *vote.Vote) error {
	return nil
}

func (noopVoteRepo) CountByPoll(ctx context.Context, pollID int64) (map[int64]int64, int64, error) {
	return map[int64]int64{}, 0, nil
}

func (noopVoteRepo) AggregatedByPoll(ctx context.Context, pollID int64) (map[int64]int64, int64, error) {
	return map[int64]int64{}, 0, nil
}

func (noopVoteRepo) IncrementAggregated(ctx context.Context, pollID, optionID int64) error {
	return nil
}

func TestLoginAndCreatePoll(t *testing.T) {
	userRepo := newTestUserRepo()
	pollRepo := newTestPollRepo()
	voteRepo := noopVoteRepo{}

	pass := "adminpass"
	hash, _ := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.MinCost)
	userRepo.seed(&user.User{Email: "admin@test.com", PasswordHash: string(hash), Role: "admin"})

	userSvc := user.NewService(userRepo)
	pollSvc := poll.NewService(pollRepo)
	voteSvc := vote.NewService(voteRepo)
	jwtMgr := jwtpkg.NewManager("secret")
	voteCh := make(chan worker.VoteEvent, 10)

	server := httptest.NewServer(NewRouter(userSvc, pollSvc, voteSvc, jwtMgr, voteCh))
	defer server.Close()

	loginReq := authRequest{Email: "admin@test.com", Password: pass}
	body, _ := json.Marshal(loginReq)
	resp, err := http.Post(server.URL+"/api/v1/auth/login", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("login request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 login, got %d", resp.StatusCode)
	}
	var loginResp map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		t.Fatalf("decode login response: %v", err)
	}
	token, ok := loginResp["token"].(string)
	if !ok || token == "" {
		t.Fatalf("token missing in response")
	}

	createReq := createPollRequest{
		Title:   "Campus Survey",
		Options: []string{"Yes", "No"},
	}
	payload, _ := json.Marshal(createReq)
	req, _ := http.NewRequest(http.MethodPost, server.URL+"/api/v1/polls", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	createResp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("create poll request failed: %v", err)
	}
	defer createResp.Body.Close()
	if createResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 create poll, got %d", createResp.StatusCode)
	}

	var respBody map[string]int64
	if err := json.NewDecoder(createResp.Body).Decode(&respBody); err != nil {
		t.Fatalf("decode create poll response: %v", err)
	}
	if respBody["id"] == 0 {
		t.Fatalf("expected poll id in response")
	}
}
