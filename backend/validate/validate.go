// Package validate holds input checks shared by several services: URLs that
// end up in href/src attributes, uploaded image bytes and user-written HTML.
package validate

import (
	"net/http"
	"net/url"
	"strings"

	"encore.dev/beta/errs"
)

// HTTPURL accepts an empty string or an absolute http/https URL. Any other
// scheme (javascript:, data:, vbscript:, ...) is rejected, because these
// values are rendered as links and image sources by the frontends.
func HTTPURL(raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return &errs.Error{Code: errs.InvalidArgument, Message: "URL non valido: ammessi solo http/https"}
	}
	return nil
}

// MaxImageBytes is the largest image accepted by the upload endpoints.
const MaxImageBytes = 5 << 20 // 5 MB

var allowedImageTypes = map[string]bool{
	"image/png":  true,
	"image/jpeg": true,
	"image/webp": true,
	"image/gif":  true,
}

// ImageContentType sniffs the content type from the bytes themselves (never
// trusting the client) and returns it if it is an allowed raster format.
// SVG and HTML are rejected because they can carry scripts.
func ImageContentType(data []byte) (string, error) {
	if len(data) == 0 {
		return "", &errs.Error{Code: errs.InvalidArgument, Message: "immagine vuota"}
	}
	if len(data) > MaxImageBytes {
		return "", &errs.Error{Code: errs.InvalidArgument, Message: "immagine troppo grande (massimo 5 MB)"}
	}
	ct := http.DetectContentType(data)
	if !allowedImageTypes[ct] {
		return "", &errs.Error{Code: errs.InvalidArgument, Message: "formato immagine non supportato: ammessi PNG, JPEG, WebP e GIF"}
	}
	return ct, nil
}
