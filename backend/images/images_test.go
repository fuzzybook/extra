package images

import (
	"encoding/base64"
	"testing"

	"encore.dev/types/uuid"
)

func TestSaveImageParams_Validate(t *testing.T) {
	// Missing description
	p := &SaveImageParams{
		Description: "",
		Data:        "abc",
	}
	if err := p.Validate(); err == nil {
		t.Error("expected error for empty description, got nil")
	}

	// Missing data for new image (uuid is empty)
	p = &SaveImageParams{
		UUID:        "",
		Description: "A test image",
		Data:        "",
	}
	if err := p.Validate(); err == nil {
		t.Error("expected error for empty data on new image, got nil")
	}

	// Valid new image (uuid is empty string)
	p = &SaveImageParams{
		UUID:        "",
		Description: "A test image",
		Data:        "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==",
	}
	if err := p.Validate(); err != nil {
		t.Errorf("expected valid params, got err: %v", err)
	}

	// Valid update (uuid present, empty data is allowed)
	id, _ := uuid.NewV4()
	p = &SaveImageParams{
		UUID:        id.String(),
		Description: "Updated description",
	}
	if err := p.Validate(); err != nil {
		t.Errorf("expected valid update params, got err: %v", err)
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
