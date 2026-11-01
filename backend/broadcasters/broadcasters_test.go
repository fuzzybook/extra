package broadcasters

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"
	"testing"

	authsvc "encore.app/auth"
	"encore.app/config"
	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/types/uuid"
)

func TestCreateBroadcasterParamsValidate(t *testing.T) {
	tests := []struct {
		name    string
		params  *CreateBroadcasterParams
		wantErr bool
	}{
		{
			name:    "nil params",
			params:  nil,
			wantErr: true,
		},
		{
			name: "empty display name",
			params: &CreateBroadcasterParams{
				DisplayName: "",
			},
			wantErr: true,
		},
		{
			name: "whitespace display name",
			params: &CreateBroadcasterParams{
				DisplayName: "   ",
			},
			wantErr: true,
		},
		{
			name: "valid broadcaster params with defaults applied",
			params: &CreateBroadcasterParams{
				DisplayName: "DJ Rossi",
			},
			wantErr: false,
		},
		{
			name: "valid full broadcaster params",
			params: &CreateBroadcasterParams{
				DisplayName:   "Live Streamer Alpha",
				Category:      "Music",
				CategoryColor: "#ec4899",
				Bio:           "Official channel for live stream sets.",
				Certified:     true,
				Availability:  "live",
				Phone:         "+39 02 1234567",
				Services:      []string{"Live DJ", "Video chat"},
				Languages:     []string{"IT", "EN"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.params.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil {
				if tt.params.Category == "" {
					t.Errorf("expected default Category, got empty string")
				}
				if tt.params.CategoryColor == "" {
					t.Errorf("expected default CategoryColor, got empty string")
				}
				if tt.params.Availability == "" {
					t.Errorf("expected default Availability, got empty string")
				}
			}
		})
	}
}

func TestUpdateBroadcasterParamsValidate(t *testing.T) {
	tests := []struct {
		name    string
		params  *UpdateBroadcasterParams
		wantErr bool
	}{
		{
			name:    "nil update params",
			params:  nil,
			wantErr: true,
		},
		{
			name: "empty display name in update",
			params: &UpdateBroadcasterParams{
				DisplayName: "",
			},
			wantErr: true,
		},
		{
			name: "valid update params",
			params: &UpdateBroadcasterParams{
				DisplayName: "New Artist Name",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.params.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBroadcasterImageConfig(t *testing.T) {
	if BroadcasterImageSize.Width != 600 {
		t.Errorf("expected default image width 600, got %d", BroadcasterImageSize.Width)
	}
	if BroadcasterImageSize.Proportion != "2:3" {
		t.Errorf("expected default image proportion 2:3, got %s", BroadcasterImageSize.Proportion)
	}
}

func TestGetCategoriesFromSiteConfig(t *testing.T) {
	resp, err := GetCategories(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Categories) == 0 {
		t.Errorf("expected at least 1 category from SiteConfig, got 0")
	}
	foundGeneral := false
	for _, c := range resp.Categories {
		if c.Name == "General" {
			foundGeneral = true
			break
		}
	}
	if !foundGeneral {
		t.Errorf("expected to find General category in SiteConfig")
	}
}

func TestValidateHours(t *testing.T) {
	// Valid slots
	valid := []BroadcasterHours{
		{
			Day: "Lunedì",
			Hours: []BroadcasterTimeSlot{
				{From: "09:00", To: "12:00"},
				{From: "14:00", To: "18:00"},
			},
		},
	}
	if err := ValidateHours(valid); err != nil {
		t.Errorf("expected valid hours, got error: %v", err)
	}

	// Invalid: From >= To
	invalidRange := []BroadcasterHours{
		{
			Day: "Martedì",
			Hours: []BroadcasterTimeSlot{
				{From: "18:00", To: "12:00"},
			},
		},
	}
	if err := ValidateHours(invalidRange); err == nil {
		t.Errorf("expected error for from >= to, got nil")
	}

	// Invalid: overlapping slots
	overlapping := []BroadcasterHours{
		{
			Day: "Mercoledì",
			Hours: []BroadcasterTimeSlot{
				{From: "10:00", To: "13:00"},
				{From: "12:00", To: "15:00"},
			},
		},
	}
	if err := ValidateHours(overlapping); err == nil {
		t.Errorf("expected error for overlapping hours, got nil")
	}
}

func TestGetBroadcasterAttributePresetsFromSiteConfig(t *testing.T) {
	resp, err := GetBroadcasterAttributePresets(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Presets) == 0 {
		t.Errorf("expected at least 1 attribute preset from SiteConfig, got 0")
	}
	foundGenre := false
	for _, p := range resp.Presets {
		if p == "Genere Musicale" {
			foundGenre = true
			break
		}
	}
	if !foundGenre {
		t.Errorf("expected to find 'Genere Musicale' in attribute presets")
	}
}

func TestGetAvailabilityOptionsFromSiteConfig(t *testing.T) {
	resp, err := GetAvailabilityOptions(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := config.GetAvailabilityOptions()
	if len(resp.Options) != len(expected) {
		t.Fatalf("expected %d availability options from SiteConfig, got %d", len(expected), len(resp.Options))
	}
	for i, opt := range resp.Options {
		if opt != expected[i] {
			t.Errorf("expected option %d to be '%s', got '%s'", i, expected[i], opt)
		}
	}
}

func TestGetBroadcasterServicesPresetsFromSiteConfig(t *testing.T) {
	resp, err := GetBroadcasterServicesPresets(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := config.GetBroadcasterServices()
	for i, svc := range resp.Services {
		if svc != expected[i] {
			t.Errorf("expected service %d to be '%s', got '%s'", i, expected[i], svc)
		}
	}
}

func TestBroadcasterPhoto_Description(t *testing.T) {
	// 1. Unmarshal from description object
	data1 := []byte(`{"image_url":"https://example.com/img.png","description":{"text":"Foto da concerto","box_x":25,"box_y":50}}`)
	var p1 BroadcasterPhoto
	if err := json.Unmarshal(data1, &p1); err != nil {
		t.Fatalf("failed to unmarshal p1: %v", err)
	}
	if p1.Description.Text != "Foto da concerto" || p1.Alt != "Foto da concerto" {
		t.Errorf("expected Description.Text and Alt to be 'Foto da concerto', got desc='%s', alt='%s'", p1.Description.Text, p1.Alt)
	}
	if p1.Description.Box_X != 25 || p1.Box_X != 25 {
		t.Errorf("expected Box_X 25, got desc=%d, bp=%d", p1.Description.Box_X, p1.Box_X)
	}
	if p1.Description.Box_y != 50 || p1.Box_y != 50 {
		t.Errorf("expected Box_y 50, got desc=%d, bp=%d", p1.Description.Box_y, p1.Box_y)
	}

	// 2. Unmarshal from legacy alt, box_x, box_y
	data2 := []byte(`{"image_url":"https://example.com/img2.png","alt":"Legacy didascalia","box_x":12,"box_y":34}`)
	var p2 BroadcasterPhoto
	if err := json.Unmarshal(data2, &p2); err != nil {
		t.Fatalf("failed to unmarshal p2: %v", err)
	}
	if p2.Description.Text != "Legacy didascalia" {
		t.Errorf("expected Description.Text 'Legacy didascalia', got '%s'", p2.Description.Text)
	}
	if p2.Description.Box_X != 12 || p2.Description.Box_y != 34 {
		t.Errorf("expected Description box 12, 34, got %d, %d", p2.Description.Box_X, p2.Description.Box_y)
	}

	// 3. Marshal output contains both alt/box_x/box_y and description object
	marshaled, err := json.Marshal(p1)
	if err != nil {
		t.Fatalf("failed to marshal p1: %v", err)
	}
	var res map[string]interface{}
	if err := json.Unmarshal(marshaled, &res); err != nil {
		t.Fatalf("failed to unmarshal marshaled result: %v", err)
	}
	if res["alt"] != "Foto da concerto" {
		t.Errorf("expected alt in json, got %v", res["alt"])
	}
	descObj, ok := res["description"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected description object in json, got %T: %v", res["description"], res["description"])
	}
	if descObj["text"] != "Foto da concerto" || descObj["box_x"] != float64(25) || descObj["box_y"] != float64(50) {
		t.Errorf("unexpected description object content: %+v", descObj)
	}
}



func TestValidateHoursComparesTimesNotStrings(t *testing.T) {
	// "9:00" >= "10:00" as strings: must be accepted as a valid range.
	ok := []BroadcasterHours{{Day: "Lunedì", Hours: []BroadcasterTimeSlot{{From: "9:00", To: "10:00"}}}}
	if err := ValidateHours(ok); err != nil {
		t.Fatalf("ValidateHours(9:00-10:00) = %v, want nil", err)
	}
	for _, slot := range []BroadcasterTimeSlot{{From: "9", To: "10:00"}, {From: "25:00", To: "26:00"}, {From: "10:60", To: "11:00"}, {From: "abc", To: "10:00"}} {
		bad := []BroadcasterHours{{Day: "Lunedì", Hours: []BroadcasterTimeSlot{slot}}}
		if err := ValidateHours(bad); err == nil {
			t.Errorf("ValidateHours(%s-%s) = nil, want error", slot.From, slot.To)
		}
	}
}

func TestForViewerRedactsPrivateFields(t *testing.T) {
	owner, _ := uuid.NewV4()
	other, _ := uuid.NewV4()
	newB := func(enabled bool) *Broadcaster {
		return &Broadcaster{UserID: owner, Phone: "+39 123", Enabled: enabled}
	}

	if b := newB(true); !forViewer(b, uuid.Nil, false) || b.Phone != "" || b.UserID != uuid.Nil {
		t.Fatalf("anonymous viewer must see enabled profile without phone/user_id, got %+v", b)
	}
	if b := newB(false); forViewer(b, other, false) {
		t.Fatal("disabled profile must be hidden from other users")
	}
	if b := newB(false); !forViewer(b, owner, false) || b.Phone == "" {
		t.Fatal("owner must see own disabled profile with phone")
	}
	if b := newB(false); !forViewer(b, other, true) || b.Phone == "" {
		t.Fatal("admin must see every profile with phone")
	}
}

func TestSanitizePhotosRejectsScriptURLsAndMarkup(t *testing.T) {
	bad := []BroadcasterPhoto{{ImageURL: "javascript:alert(1)"}}
	if err := sanitizePhotos(bad); err == nil {
		t.Fatal("sanitizePhotos must reject javascript: URLs")
	}
	photos := []BroadcasterPhoto{{ImageURL: "https://cdn.example.com/a.png", Alt: `<b>ok</b><img src=x onerror=alert(1)>`}}
	if err := sanitizePhotos(photos); err != nil {
		t.Fatalf("sanitizePhotos returned %v", err)
	}
	if photos[0].Alt != "<b>ok</b>" {
		t.Fatalf("Alt = %q, want %q", photos[0].Alt, "<b>ok</b>")
	}
}

func TestShortIDGeneratedOnInsertAndUnique(t *testing.T) {
	ctx := context.Background()
	insert := func() uuid.UUID {
		id, _ := uuid.NewV4()
		userID, _ := uuid.NewV4()
		// Same shape as the direct insert in admin.UpdateProfileMemberType: no short_id given.
		if _, err := db.Exec(ctx, `INSERT INTO broadcasters (id, user_id, display_name) VALUES ($1, $2, 'Short ID test')`, id, userID); err != nil {
			t.Fatalf("insert: %v", err)
		}
		return id
	}
	first, second := insert(), insert()

	shortIDs := map[uuid.UUID]string{}
	for _, id := range []uuid.UUID{first, second} {
		var shortID string
		if err := db.QueryRow(ctx, `SELECT short_id FROM broadcasters WHERE id = $1`, id).Scan(&shortID); err != nil {
			t.Fatalf("select short_id: %v", err)
		}
		if !regexp.MustCompile(`^[0-9a-z]{6}$`).MatchString(shortID) {
			t.Fatalf("short_id %q does not match ^[0-9a-z]{6}$", shortID)
		}
		shortIDs[id] = shortID
	}
	if shortIDs[first] == shortIDs[second] {
		t.Fatalf("two broadcasters got the same short_id %q", shortIDs[first])
	}

	// A duplicate value must be rejected by the unique constraint.
	if _, err := db.Exec(ctx, `UPDATE broadcasters SET short_id = $1 WHERE id = $2`, shortIDs[first], second); err == nil {
		t.Fatal("duplicate short_id was accepted")
	}
	// The format check rejects anything that is not 6 chars of [0-9a-z].
	if _, err := db.Exec(ctx, `UPDATE broadcasters SET short_id = 'ABC123' WHERE id = $1`, second); err == nil {
		t.Fatal("short_id with uppercase letters was accepted")
	}

	b, err := getBroadcaster(ctx, first)
	if err != nil {
		t.Fatalf("getBroadcaster: %v", err)
	}
	if b.ShortID != shortIDs[first] {
		t.Fatalf("Broadcaster.ShortID = %q, want %q", b.ShortID, shortIDs[first])
	}
}

func TestLookupByShortIDsReturnsFirstPhotoCaptionAndSize(t *testing.T) {
	ctx := context.Background()
	id, _ := uuid.NewV4()
	userID, _ := uuid.NewV4()
	imageID, _ := uuid.NewV4()
	photos := `[{"image_url": "http://img.example/images/` + imageID.String() + `",
		"description": {"text": "<b>Ciao</b><img src=x onerror=alert(1)>", "box_x": 14, "box_y": 77}},
		{"image_url": "http://img.example/images/second"}]`
	if _, err := db.Exec(ctx, `INSERT INTO broadcasters (id, user_id, display_name, photos, enabled) VALUES ($1, $2, 'Card test', $3::jsonb, TRUE)`, id, userID, photos); err != nil {
		t.Fatalf("insert: %v", err)
	}
	var shortID string
	if err := db.QueryRow(ctx, `SELECT short_id FROM broadcasters WHERE id = $1`, id).Scan(&shortID); err != nil {
		t.Fatalf("select short_id: %v", err)
	}

	resp, err := LookupByShortIDs(ctx, &LookupByShortIDsParams{ShortIDs: []string{shortID, "000000"}})
	if err != nil {
		t.Fatalf("LookupByShortIDs: %v", err)
	}
	if len(resp.Broadcasters) != 1 {
		t.Fatalf("got %d broadcasters, want 1 (unknown short_id omitted)", len(resp.Broadcasters))
	}
	c := resp.Broadcasters[0]
	if c.DisplayName != "Card test" || !strings.HasSuffix(c.ImageURL, imageID.String()) || !c.Enabled {
		t.Fatalf("unexpected card %+v", c)
	}
	if c.Description != "<b>Ciao</b>" {
		t.Fatalf("Description = %q, want sanitized %q", c.Description, "<b>Ciao</b>")
	}
	if c.BoxX != 14 || c.BoxY != 77 {
		t.Fatalf("caption position = (%d, %d), want (14, 77)", c.BoxX, c.BoxY)
	}
	// The image is not in the images service: standard broadcaster size.
	if c.Width != BroadcasterImageSize.Width || c.Height != BroadcasterImageSize.Height || c.Proportion != BroadcasterImageSize.Proportion {
		t.Fatalf("size = %dx%d %q, want standard broadcaster size", c.Width, c.Height, c.Proportion)
	}
}

func TestGetBroadcasterByShortIDHidesDisabledAndPhone(t *testing.T) {
	ctx := context.Background()
	insert := func(name string, enabled bool) string {
		id, _ := uuid.NewV4()
		userID, _ := uuid.NewV4()
		if _, err := db.Exec(ctx, `INSERT INTO broadcasters (id, user_id, display_name, phone, enabled) VALUES ($1, $2, $3, '+41 79 000 00 00', $4)`, id, userID, name, enabled); err != nil {
			t.Fatalf("insert: %v", err)
		}
		var shortID string
		if err := db.QueryRow(ctx, `SELECT short_id FROM broadcasters WHERE id = $1`, id).Scan(&shortID); err != nil {
			t.Fatalf("select short_id: %v", err)
		}
		return shortID
	}

	enabled := insert("Public page", true)
	b, err := GetBroadcasterByShortID(ctx, strings.ToUpper(enabled))
	if err != nil {
		t.Fatalf("GetBroadcasterByShortID(enabled) = %v", err)
	}
	if b.DisplayName != "Public page" || b.Phone != "" || b.UserID != uuid.Nil {
		t.Fatalf("public broadcaster = %+v, want name only, no phone/user_id", b)
	}

	disabled := insert("Hidden page", false)
	for _, id := range []string{disabled, "000000", "bad", "../x"} {
		if _, err := GetBroadcasterByShortID(ctx, id); err == nil {
			t.Errorf("GetBroadcasterByShortID(%q) returned a profile", id)
		}
	}

	// La videochat pubblica ricava il proprietario solo dei profili abilitati.
	owner, err := GetLiveOwner(ctx, enabled)
	if err != nil || owner.DisplayName != "Public page" || owner.UserID == "" {
		t.Fatalf("GetLiveOwner(enabled) = %+v, %v", owner, err)
	}
	for _, id := range []string{disabled, "000000", "bad"} {
		if _, err := GetLiveOwner(ctx, id); errs.Code(err) != errs.NotFound {
			t.Errorf("GetLiveOwner(%q) = %v, want not found", id, err)
		}
	}

	// Il sito pubblico chiede per utente se può andare in diretta.
	mine, err := GetLiveOwnerByUser(ctx, uuid.Must(uuid.FromString(owner.UserID)))
	if err != nil || mine.ShortID != enabled || owner.ShortID != enabled {
		t.Fatalf("GetLiveOwnerByUser = %+v, %v; want short id %q", mine, err, enabled)
	}
	var disabledUser uuid.UUID
	if err := db.QueryRow(ctx, `SELECT user_id FROM broadcasters WHERE short_id = $1`, disabled).Scan(&disabledUser); err != nil {
		t.Fatalf("select disabled user: %v", err)
	}
	ids, err := ShortIDsByUsers(ctx, &ShortIDsByUsersParams{UserIDs: []string{owner.UserID, disabledUser.String(), "bad"}})
	if err != nil || len(ids.ShortIDs) != 1 || ids.ShortIDs[owner.UserID] != enabled {
		t.Fatalf("ShortIDsByUsers = %+v, %v; want only the enabled broadcaster", ids, err)
	}
	for _, id := range []uuid.UUID{disabledUser, uuid.Must(uuid.NewV4())} {
		if _, err := GetLiveOwnerByUser(ctx, id); errs.Code(err) != errs.NotFound {
			t.Errorf("GetLiveOwnerByUser(%s) = %v, want not found", id, err)
		}
	}
}

func TestUpdateBroadcasterVisibilityByOwner(t *testing.T) {
	ctx := context.Background()
	id, _ := uuid.NewV4()
	ownerID, _ := uuid.NewV4()
	strangerID, _ := uuid.NewV4()
	if _, err := db.Exec(ctx, `INSERT INTO broadcasters (id, user_id, display_name, enabled, certified) VALUES ($1, $2, 'Visibility test', TRUE, FALSE)`, id, ownerID); err != nil {
		t.Fatalf("insert: %v", err)
	}
	ownerCtx := auth.WithContext(ctx, auth.UID(ownerID.String()), &authsvc.AuthData{})
	strangerCtx := auth.WithContext(ctx, auth.UID(strangerID.String()), &authsvc.AuthData{})

	// Solo il proprietario (o un admin) cambia la visibilità.
	if _, err := UpdateBroadcasterVisibility(strangerCtx, id, &UpdateVisibilityParams{Enabled: false}); errs.Code(err) != errs.PermissionDenied {
		t.Fatalf("stranger: %v, want permission denied", err)
	}
	b, err := UpdateBroadcasterVisibility(ownerCtx, id, &UpdateVisibilityParams{Enabled: false})
	if err != nil || b.Enabled {
		t.Fatalf("owner hide = %+v, %v", b, err)
	}
	if _, err := GetBroadcasterByShortID(ctx, b.ShortID); err == nil {
		t.Fatal("hidden profile still visible to the public")
	}
	b, err = UpdateBroadcasterVisibility(ownerCtx, id, &UpdateVisibilityParams{Enabled: true})
	if err != nil || !b.Enabled || b.Certified {
		t.Fatalf("owner show = %+v, %v (certified must not change)", b, err)
	}
	if _, err := GetBroadcasterByShortID(ctx, b.ShortID); err != nil {
		t.Fatalf("visible profile not public: %v", err)
	}
}

