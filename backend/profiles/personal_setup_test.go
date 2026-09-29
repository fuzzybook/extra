package profiles

import (
	"context"
	"strings"
	"testing"

	authsvc "encore.app/auth"
	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/et"
	"encore.dev/types/uuid"
)

func TestPersonalSetupValidation(t *testing.T) {
	for _, p := range []*PersonalSetupParams{
		nil,
		{CanBroadcast: true},
		{CanBroadcast: true, RoomName: "   "},
		{CanBroadcast: true, RoomName: "a/b"},
		{CanBroadcast: true, RoomName: "a\x00b"},
		{CanBroadcast: true, RoomName: strings.Repeat("a", 129)},
	} {
		if err := p.Validate(); errs.Code(err) != errs.InvalidArgument {
			t.Errorf("expected invalid setup, got %v", err)
		}
	}
	p := &PersonalSetupParams{CanBroadcast: true, RoomName: " demo "}
	if err := p.Validate(); err != nil || p.RoomName != "demo" {
		t.Fatalf("expected normalized room name: %+v, %v", p, err)
	}
	p.CanBroadcast = false
	if err := p.Validate(); err != nil || p.RoomName != "" {
		t.Fatalf("revoking broadcasting must clear the room: %+v, %v", p, err)
	}
}

func TestPersonalSetupPermissions(t *testing.T) {
	defer et.OverrideAuthInfo("", nil)
	ctx := context.Background()
	et.OverrideAuthInfo("", nil)
	if _, err := GetPersonalSetup(ctx); errs.Code(err) != errs.Unauthenticated {
		t.Fatalf("expected unauthenticated, got %v", err)
	}
	if _, err := UpsertPersonalSetup(ctx, uuid.Nil, &PersonalSetupParams{}); errs.Code(err) != errs.Unauthenticated {
		t.Fatalf("expected unauthenticated, got %v", err)
	}
	if _, err := GetUserPersonalSetup(ctx, uuid.Nil); errs.Code(err) != errs.Unauthenticated {
		t.Fatalf("anonymous admin read: %v", err)
	}
	et.OverrideAuthInfo("user", &authsvc.AuthData{Role: "user"})
	if _, err := GetUserPersonalSetup(ctx, uuid.Nil); errs.Code(err) != errs.PermissionDenied {
		t.Fatalf("non-admin read: %v", err)
	}
	if _, err := UpsertPersonalSetup(ctx, uuid.Nil, &PersonalSetupParams{CanBroadcast: true, RoomName: "demo"}); errs.Code(err) != errs.PermissionDenied {
		t.Fatalf("users must not grant themselves broadcasting: %v", err)
	}
}

func TestPersonalSetupPersistence(t *testing.T) {
	ctx := context.Background()
	testDB, err := et.NewTestDatabase(ctx, "profiles")
	if err != nil {
		t.Fatal(err)
	}
	originalDB := db
	db = testDB
	defer func() { db = originalDB; et.OverrideAuthInfo("", nil) }()
	userID, err := uuid.NewV4()
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(ctx, `INSERT INTO user_profiles (user_id, display_name, email) VALUES ($1, 'Setup Test', 'setup@example.invalid')`, userID)
	if err != nil {
		t.Fatal(err)
	}
	et.OverrideAuthInfo(auth.UID(userID.String()), &authsvc.AuthData{Role: "user"})
	setup, err := GetPersonalSetup(ctx)
	if err != nil || setup.CanBroadcast || setup.RoomName != "" {
		t.Fatalf("expected disabled default setup: %+v, %v", setup, err)
	}
	et.OverrideAuthInfo("admin", &authsvc.AuthData{Role: "admin"})
	if _, err := UpsertPersonalSetup(ctx, userID, &PersonalSetupParams{CanBroadcast: true, RoomName: "demo"}); err != nil {
		t.Fatal(err)
	}
	setup, err = GetUserPersonalSetup(ctx, userID)
	if err != nil || setup.UserID != userID || !setup.CanBroadcast || setup.RoomName != "demo" {
		t.Fatalf("admin must read selected user's settings: %+v, %v", setup, err)
	}
	if _, err := GetUserPersonalSetup(ctx, uuid.Nil); errs.Code(err) != errs.NotFound {
		t.Fatalf("missing profile read: %v", err)
	}
	et.OverrideAuthInfo(auth.UID(userID.String()), &authsvc.AuthData{Role: "user"})
	setup, err = GetPersonalSetup(ctx)
	if err != nil || !setup.CanBroadcast || setup.RoomName != "demo" || setup.UserID != userID {
		t.Fatalf("expected persisted setup: %+v, %v", setup, err)
	}
	et.OverrideAuthInfo("admin", &authsvc.AuthData{Role: "admin"})
	setup, err = UpsertPersonalSetup(ctx, userID, &PersonalSetupParams{CanBroadcast: false, RoomName: "demo"})
	if err != nil || setup.CanBroadcast || setup.RoomName != "" {
		t.Fatalf("expected revoked setup: %+v, %v", setup, err)
	}
	if _, err := UpsertPersonalSetup(ctx, uuid.Nil, &PersonalSetupParams{}); errs.Code(err) != errs.NotFound {
		t.Fatalf("expected missing profile: %v", err)
	}
	if _, err := db.Exec(ctx, `DELETE FROM user_profiles WHERE user_id = $1`, userID); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.QueryRow(ctx, `SELECT COUNT(*) FROM personal_setup WHERE user_id = $1`, userID).Scan(&count); err != nil || count != 0 {
		t.Fatalf("setup must be deleted with profile: count=%d, err=%v", count, err)
	}
	et.OverrideAuthInfo(auth.UID(userID.String()), &authsvc.AuthData{Role: "user"})
	if _, err := GetPersonalSetup(ctx); errs.Code(err) != errs.NotFound {
		t.Fatalf("expected missing profile: %v", err)
	}
}
