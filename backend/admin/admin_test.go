package admin

import (
	"context"
	"testing"

	authsvc "encore.app/auth"
	profilessvc "encore.app/profiles"
	encoreauth "encore.dev/beta/auth"
	"encore.dev/et"
	"encore.dev/types/uuid"
)

func TestUpdateProfileMemberTypeBroadcasterLifecycle(t *testing.T) {
	adminID, _ := uuid.NewV4()
	et.OverrideAuthInfo(encoreauth.UID(adminID.String()), &authsvc.AuthData{Role: "admin"})
	defer et.OverrideAuthInfo("", nil)

	ctx := context.Background()

	// 1. Create a test user profile (initially visitor)
	testUserID, err := uuid.NewV4()
	if err != nil {
		t.Fatalf("failed to generate user id: %v", err)
	}

	_, err = profilesDB.Exec(ctx, `
		INSERT INTO user_profiles (user_id, email, display_name, member_type)
		VALUES ($1, $2, $3, 'visitor')
	`, testUserID, "test-broadcaster-lifecycle@example.com", "Lifecycle Artist")
	if err != nil {
		t.Fatalf("failed to insert test profile: %v", err)
	}

	// 2. Change member_type to 'broadcaster' -> should create broadcaster with used=true, enabled=false
	p, err := UpdateProfileMemberType(ctx, testUserID, &UpdateProfileMemberTypeParams{
		MemberType: profilessvc.MemberTypeBroadcaster,
	})
	if err != nil {
		t.Fatalf("UpdateProfileMemberType to broadcaster failed: %v", err)
	}
	if p.MemberType != profilessvc.MemberTypeBroadcaster {
		t.Errorf("expected member_type 'broadcaster', got '%s'", p.MemberType)
	}
	if p.Broadcaster == nil {
		t.Fatalf("expected profile.Broadcaster to be non-nil")
	}

	createdBroadcasterID := *p.Broadcaster

	// Verify in broadcasters DB: used=true, enabled=false
	var enabled, used bool
	err = broadcastersDB.QueryRow(ctx, `SELECT enabled, used FROM broadcasters WHERE id = $1`, createdBroadcasterID).Scan(&enabled, &used)
	if err != nil {
		t.Fatalf("failed to query broadcaster: %v", err)
	}
	if !used {
		t.Errorf("expected broadcaster.used to be true, got false")
	}
	if enabled {
		t.Errorf("expected broadcaster.enabled to be false, got true")
	}

	// 3. Change member_type from 'broadcaster' to 'visitor' -> should disable broadcaster (enabled=false, used=false) and set profile.broadcaster = nil
	p2, err := UpdateProfileMemberType(ctx, testUserID, &UpdateProfileMemberTypeParams{
		MemberType: profilessvc.MemberTypeVisitor,
	})
	if err != nil {
		t.Fatalf("UpdateProfileMemberType to visitor failed: %v", err)
	}
	if p2.MemberType != profilessvc.MemberTypeVisitor {
		t.Errorf("expected member_type 'visitor', got '%s'", p2.MemberType)
	}
	if p2.Broadcaster != nil {
		t.Errorf("expected profile.Broadcaster to be nil, got %v", p2.Broadcaster)
	}

	// Verify in broadcasters DB: used=false, enabled=false
	err = broadcastersDB.QueryRow(ctx, `SELECT enabled, used FROM broadcasters WHERE id = $1`, createdBroadcasterID).Scan(&enabled, &used)
	if err != nil {
		t.Fatalf("failed to query broadcaster: %v", err)
	}
	if used {
		t.Errorf("expected broadcaster.used to be false after type change, got true")
	}
	if enabled {
		t.Errorf("expected broadcaster.enabled to be false after type change, got true")
	}

	// 4. Change back to 'broadcaster' -> should reuse existing broadcaster and set used=true, enabled=false
	p3, err := UpdateProfileMemberType(ctx, testUserID, &UpdateProfileMemberTypeParams{
		MemberType: profilessvc.MemberTypeBroadcaster,
	})
	if err != nil {
		t.Fatalf("UpdateProfileMemberType back to broadcaster failed: %v", err)
	}
	if p3.MemberType != profilessvc.MemberTypeBroadcaster {
		t.Errorf("expected member_type 'broadcaster', got '%s'", p3.MemberType)
	}
	if p3.Broadcaster == nil || *p3.Broadcaster != createdBroadcasterID {
		t.Errorf("expected existing broadcaster ID %v, got %v", createdBroadcasterID, p3.Broadcaster)
	}

	err = broadcastersDB.QueryRow(ctx, `SELECT enabled, used FROM broadcasters WHERE id = $1`, createdBroadcasterID).Scan(&enabled, &used)
	if err != nil {
		t.Fatalf("failed to query broadcaster: %v", err)
	}
	if !used {
		t.Errorf("expected broadcaster.used to be true on reactivation, got false")
	}
	if enabled {
		t.Errorf("expected broadcaster.enabled to be false on reactivation, got true")
	}
}
