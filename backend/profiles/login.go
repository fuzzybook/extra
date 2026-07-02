package profiles

import (
	"context"
	"errors"
	"sync"
	"time"

	"encore.app/auth"
	"encore.dev/beta/errs"
	"encore.dev/pubsub"
	"encore.dev/storage/sqldb"
	"encore.dev/types/uuid"
)

const loginTimeout = 10 * time.Second

type LoginParams struct {
	UserEmail string `json:"user_email"`
	Password  string `json:"password" encore:"sensitive"`
}

type LoginResponse struct {
	Token     uuid.UUID `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

var loginRequests = pubsub.TopicRef[pubsub.Publisher[*auth.LoginRequested]](auth.LoginRequests)

// Login is the public entry point for authenticating a user. It first
// resolves the given email to a user ID via the local profiles database.
// The actual credential check lives in the auth service; rather than calling
// it directly, this hands the request off over Pub/Sub (publishing to
// auth.LoginRequests) and waits for the matching auth.LoginCompleted reply
// on auth.LoginResults before responding to the caller.
//
// NOTE: the wait is implemented with an in-process registry keyed by request
// ID, so the instance that publishes the request must be the one that
// receives the reply. That holds for local development (a single instance)
// but not for a horizontally scaled deployment, which would need a shared
// mechanism instead (e.g. a results table polled with retry_until).
//
//encore:api public method=POST path=/auth/login
func Login(ctx context.Context, p *LoginParams) (*LoginResponse, error) {
	var userID uuid.UUID
	err := db.QueryRow(ctx, `
		SELECT user_id FROM user_profiles WHERE email = $1
	`, p.UserEmail).Scan(&userID)
	if errors.Is(err, sqldb.ErrNoRows) {
		return nil, &errs.Error{Code: errs.Unauthenticated, Message: "invalid credentials"}
	}
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to look up user")
	}

	requestID, err := uuid.NewV4()
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to generate request id")
	}

	wait := loginWaiters.register(requestID)
	defer loginWaiters.cancel(requestID)

	if _, err := loginRequests.Publish(ctx, &auth.LoginRequested{
		RequestID: requestID,
		UserID:    userID,
		Password:  p.Password,
	}); err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to publish login request")
	}

	select {
	case result := <-wait:
		if !result.OK {
			return nil, &errs.Error{Code: errs.Unauthenticated, Message: "invalid credentials"}
		}
		return &LoginResponse{Token: result.Token, ExpiresAt: result.ExpiresAt}, nil
	case <-time.After(loginTimeout):
		return nil, &errs.Error{Code: errs.DeadlineExceeded, Message: "login timed out"}
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

var _ = pubsub.NewSubscription(
	auth.LoginResults, "deliver-to-waiting-login",
	pubsub.SubscriptionConfig[*auth.LoginCompleted]{
		Handler: func(ctx context.Context, ev *auth.LoginCompleted) error {
			loginWaiters.deliver(ev)
			return nil
		},
	},
)

// loginWaiters lets the synchronous Login handler above wait for the
// asynchronous LoginCompleted reply that matches its RequestID.
var loginWaiters = newWaiterRegistry()

type waiterRegistry struct {
	mu      sync.Mutex
	waiting map[uuid.UUID]chan *auth.LoginCompleted
}

func newWaiterRegistry() *waiterRegistry {
	return &waiterRegistry{waiting: make(map[uuid.UUID]chan *auth.LoginCompleted)}
}

func (r *waiterRegistry) register(id uuid.UUID) <-chan *auth.LoginCompleted {
	ch := make(chan *auth.LoginCompleted, 1)
	r.mu.Lock()
	r.waiting[id] = ch
	r.mu.Unlock()
	return ch
}

func (r *waiterRegistry) cancel(id uuid.UUID) {
	r.mu.Lock()
	delete(r.waiting, id)
	r.mu.Unlock()
}

func (r *waiterRegistry) deliver(ev *auth.LoginCompleted) {
	r.mu.Lock()
	ch, ok := r.waiting[ev.RequestID]
	if ok {
		delete(r.waiting, ev.RequestID)
	}
	r.mu.Unlock()
	if ok {
		ch <- ev
	}
}
