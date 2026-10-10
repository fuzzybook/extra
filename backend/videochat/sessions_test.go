package videochat

import (
	"context"
	"testing"
)

func TestBroadcastSessionLifecycle(t *testing.T) {
	ctx := context.Background()

	// 1. Log OPEN
	err := logBroadcastSession(ctx, "test-owner", "test-sid-1", "test-room", "OPEN")
	if err != nil {
		t.Fatalf("failed to log OPEN: %v", err)
	}

	// 2. Duplicate OPEN should be ignored (idempotent)
	err = logBroadcastSession(ctx, "test-owner", "test-sid-1", "test-room", "OPEN")
	if err != nil {
		t.Fatalf("failed on duplicate OPEN: %v", err)
	}

	var openCount int
	err = db.QueryRow(ctx, "SELECT count(*) FROM broadcast_sessions WHERE sid = $1 AND action = 'OPEN'", "test-sid-1").Scan(&openCount)
	if err != nil || openCount != 1 {
		t.Fatalf("expected exactly 1 OPEN record, got %d (err: %v)", openCount, err)
	}

	// 3. Log CLOSE
	err = logBroadcastSession(ctx, "test-owner", "test-sid-1", "test-room", "CLOSE")
	if err != nil {
		t.Fatalf("failed to log CLOSE: %v", err)
	}

	// 4. Duplicate CLOSE should be ignored
	err = logBroadcastSession(ctx, "test-owner", "test-sid-1", "test-room", "CLOSE")
	if err != nil {
		t.Fatalf("failed on duplicate CLOSE: %v", err)
	}

	var closeCount int
	err = db.QueryRow(ctx, "SELECT count(*) FROM broadcast_sessions WHERE sid = $1 AND action = 'CLOSE'", "test-sid-1").Scan(&closeCount)
	if err != nil || closeCount != 1 {
		t.Fatalf("expected exactly 1 CLOSE record, got %d (err: %v)", closeCount, err)
	}

	// 5. Test syncClosedSessions
	err = logBroadcastSession(ctx, "test-owner-2", "test-sid-orphan", "test-room-2", "OPEN")
	if err != nil {
		t.Fatalf("failed to log orphan OPEN: %v", err)
	}

	// Active SIDs does not contain test-sid-orphan -> should be automatically closed
	syncClosedSessions(ctx, map[string]bool{"other-sid": true})

	var orphanCloseCount int
	err = db.QueryRow(ctx, "SELECT count(*) FROM broadcast_sessions WHERE sid = $1 AND action = 'CLOSE'", "test-sid-orphan").Scan(&orphanCloseCount)
	if err != nil || orphanCloseCount != 1 {
		t.Fatalf("expected syncClosedSessions to create CLOSE record, got %d (err: %v)", orphanCloseCount, err)
	}
}
