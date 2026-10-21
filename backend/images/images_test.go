package images

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"encore.dev/types/uuid"
)

func TestSaveImageParams_Validate(t *testing.T) {
	// Nil params
	if err := (*SaveImageParams)(nil).Validate(); err == nil {
		t.Error("expected error for nil params, got nil")
	}

	// Empty description is allowed for new image with data
	p := &SaveImageParams{
		Description: ImageDescription{},
		Data:        "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==",
	}
	if err := p.Validate(); err != nil {
		t.Errorf("expected valid for empty description, got %v", err)
	}

	// Missing data for new image (uuid is empty)
	p = &SaveImageParams{
		UUID:        "",
		Description: ImageDescription{Text: "A test image"},
		Data:        "",
	}
	if err := p.Validate(); err == nil {
		t.Error("expected error for empty data on new image, got nil")
	}

	// Valid new image (uuid is empty string)
	p = &SaveImageParams{
		UUID:        "",
		Description: ImageDescription{Text: "A test image"},
		Data:        "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==",
	}
	if err := p.Validate(); err != nil {
		t.Errorf("expected valid params, got err: %v", err)
	}

	// Valid update (uuid present, empty data is allowed)
	id, _ := uuid.NewV4()
	p = &SaveImageParams{
		UUID:        id.String(),
		Description: ImageDescription{Text: "Updated description", Box_X: 10, Box_y: 20},
	}
	if err := p.Validate(); err != nil {
		t.Errorf("expected valid update params, got err: %v", err)
	}
}

func TestImageDescription_UnmarshalJSON(t *testing.T) {
	// From string literal
	var d1 ImageDescription
	if err := json.Unmarshal([]byte(`"plain description text"`), &d1); err != nil {
		t.Fatalf("failed to unmarshal string: %v", err)
	}
	if d1.Text != "plain description text" || d1.Box_X != 0 || d1.Box_y != 0 {
		t.Errorf("unexpected d1: %+v", d1)
	}

	// From JSON object
	var d2 ImageDescription
	if err := json.Unmarshal([]byte(`{"text":"fancy text","box_x":15,"box_y":25}`), &d2); err != nil {
		t.Fatalf("failed to unmarshal object: %v", err)
	}
	if d2.Text != "fancy text" || d2.Box_X != 15 || d2.Box_y != 25 {
		t.Errorf("unexpected d2: %+v", d2)
	}
}

func TestImageDescription_SQL(t *testing.T) {
	d := ImageDescription{Text: "hello sql", Box_X: 5, Box_y: 10}
	val, err := d.Value()
	if err != nil {
		t.Fatalf("Value() error: %v", err)
	}
	valStr, ok := val.(string)
	if !ok {
		t.Fatalf("expected string from Value(), got %T", val)
	}

	var scanned ImageDescription
	if err := scanned.Scan(valStr); err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if scanned.Text != "hello sql" || scanned.Box_X != 5 || scanned.Box_y != 10 {
		t.Errorf("unexpected scanned value: %+v", scanned)
	}

	// Legacy plain text scan
	var legacy ImageDescription
	if err := legacy.Scan("legacy text without json"); err != nil {
		t.Fatalf("Scan() legacy error: %v", err)
	}
	if legacy.Text != "legacy text without json" {
		t.Errorf("expected legacy text, got %+v", legacy)
	}
}

func TestParseImageData(t *testing.T) {
	raw := []byte("hello world")
	encoded := base64.StdEncoding.EncodeToString(raw)

	// Plain base64
	data, ct, err := parseImageData(encoded, "image/jpeg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != "hello world" {
		t.Errorf("expected 'hello world', got %q", string(data))
	}
	if ct != "image/jpeg" {
		t.Errorf("expected 'image/jpeg', got %q", ct)
	}

	// Data URI
	dataURI := "data:image/webp;base64," + encoded
	data, ct, err = parseImageData(dataURI, "image/png")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != "hello world" {
		t.Errorf("expected 'hello world', got %q", string(data))
	}
	if ct != "image/webp" {
		t.Errorf("expected 'image/webp', got %q", ct)
	}
}

func TestImageCategoryConstants(t *testing.T) {
	if ImageCategoryGeneral != "general" {
		t.Errorf("expected general, got %s", ImageCategoryGeneral)
	}
	if ImageCategoryAvatar != "avatar" {
		t.Errorf("expected avatar, got %s", ImageCategoryAvatar)
	}
	if ImageCategoryPromo != "promo" {
		t.Errorf("expected promo, got %s", ImageCategoryPromo)
	}
	if ImageCategoryBanner != "banner" {
		t.Errorf("expected banner, got %s", ImageCategoryBanner)
	}
}
