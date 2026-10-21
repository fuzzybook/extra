package validate

import "testing"

func TestHTTPURL(t *testing.T) {
	for _, ok := range []string{"", "https://example.com/a", "http://127.0.0.1:9800/x"} {
		if err := HTTPURL(ok); err != nil {
			t.Errorf("HTTPURL(%q) = %v, want nil", ok, err)
		}
	}
	for _, bad := range []string{"javascript:alert(1)", " JavaScript:alert(1)", "data:text/html,x", "/relative", "https://", "ftp://example.com"} {
		if err := HTTPURL(bad); err == nil {
			t.Errorf("HTTPURL(%q) = nil, want error", bad)
		}
	}
}

func TestImageContentType(t *testing.T) {
	png := []byte("\x89PNG\r\n\x1a\n0000")
	if ct, err := ImageContentType(png); err != nil || ct != "image/png" {
		t.Fatalf("ImageContentType(png) = %q, %v", ct, err)
	}
	for name, data := range map[string][]byte{
		"empty": nil,
		"html":  []byte("<html><script>alert(1)</script></html>"),
		"svg":   []byte(`<svg xmlns="http://www.w3.org/2000/svg"><script>alert(1)</script></svg>`),
		"big":   append(append([]byte{}, png...), make([]byte, MaxImageBytes)...),
	} {
		if _, err := ImageContentType(data); err == nil {
			t.Errorf("ImageContentType(%s) = nil error, want rejection", name)
		}
	}
}

func TestSanitizeInlineHTML(t *testing.T) {
	cases := map[string]string{
		"plain text":                                      "plain text",
		"<b>bold</b> and <i>it</i>":                       "<b>bold</b> and <i>it</i>",
		`<img src=x onerror=alert(1)>hi`:                  "hi",
		`<b onclick="x()">a</b>`:                          "<b>a</b>",
		`<script>alert(1)</script>ok`:                     "ok",
		`<a href="javascript:alert(1)">link</a>`:          "link",
		"<b>unclosed":                                     "<b>unclosed</b>",
		"a<br/>b":                                         "a<br>b",
		"1 < 2 &amp; 3":                                   "1 &lt; 2 &amp; 3",
		`<div><strong>x</strong></div><style>p{}</style>`: "<strong>x</strong>",
	}
	for in, want := range cases {
		if got := SanitizeInlineHTML(in); got != want {
			t.Errorf("SanitizeInlineHTML(%q) = %q, want %q", in, got, want)
		}
	}
}
