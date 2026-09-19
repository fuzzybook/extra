package profiles

import (
	"context"
	"testing"
	"time"

	"encore.app/auth"
	"encore.dev/beta/errs"
	"encore.dev/storage/sqldb"
	"encore.dev/types/uuid"
)

func TestPollLoginResultReturnsValueWhenReady(t *testing.T) {
	requestID, err := uuid.NewV4()
	if err != nil {
		t.Fatalf("uuid.NewV4() failed: %v", err)
	}
	token, err := uuid.NewV4()
	if err != nil {
		t.Fatalf("uuid.NewV4() failed: %v", err)
	}

	calls := 0
	result, err := pollLoginResult(context.Background(), func(context.Context) (*auth.LoginCompleted, error) {
		calls++
		if calls < 3 {
			return nil, sqldb.ErrNoRows
		}
		return &auth.LoginCompleted{
			RequestID: requestID,
			OK:        true,
			Token:     token,
			ExpiresAt: time.Now().Add(time.Minute),
		}, nil
	}, 200*time.Millisecond, 10*time.Millisecond)
	if err != nil {
		t.Fatalf("pollLoginResult() returned error: %v", err)
	}
	if !result.OK {
		t.Fatal("expected successful login result")
	}
	if result.RequestID != requestID {
		t.Fatalf("request ID mismatch: got %v, want %v", result.RequestID, requestID)
	}
	if result.Token != token {
		t.Fatalf("token mismatch: got %v, want %v", result.Token, token)
	}
}

func TestPollLoginResultTimesOut(t *testing.T) {
	_, err := pollLoginResult(context.Background(), func(context.Context) (*auth.LoginCompleted, error) {
		return nil, sqldb.ErrNoRows
	}, 50*time.Millisecond, 10*time.Millisecond)
	if err == nil {
		t.Fatal("expected timeout error")
	}
	e, ok := err.(*errs.Error)
	if !ok {
		t.Fatalf("expected *errs.Error, got %T", err)
	}
	if e.Code != errs.DeadlineExceeded {
		t.Fatalf("expected deadline exceeded error, got %v", e.Code)
	}
}
