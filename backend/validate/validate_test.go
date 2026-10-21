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
		// Plain text and editor formatting are kept.
		"plain text":                "plain text",
		"<b>bold</b> and <i>it</i>": "<b>bold</b> and <i>it</i>",
		`<font face="Arial Black" size="6">Pippo</font>`:               `<font face="Arial Black" size="6">Pippo</font>`,
		`<font color="#ff0000">rosso</font>`:                           `<font color="#ff0000">rosso</font>`,
		`<div>riga 1</div><div><u>riga 2</u></div>`:                    `<div>riga 1</div><div><u>riga 2</u></div>`,
		`<span style="font-family: Georgia; font-size: 18px">x</span>`: `<span style="font-family: Georgia; font-size: 18px">x</span>`,
		"a<br/>b":       "a<br>b",
		"1 < 2 &amp; 3": "1 &lt; 2 &amp; 3",
		"<b>unclosed":   "<b>unclosed</b>",
		// Dangerous markup, attributes and CSS are removed.
		`<img src=x onerror=alert(1)>hi`:                                        "hi",
		`<b onclick="x()">a</b>`:                                                "<b>a</b>",
		`<script>alert(1)</script>ok`:                                           "ok",
		`<a href="javascript:alert(1)">link</a>`:                                "link",
		`<font face="x" onmouseover="alert(1)">f</font>`:                        `<font face="x">f</font>`,
		`<font size="9" color="red;background:url(x)">f</font>`:                 "<font>f</font>",
		`<span style="background:url(javascript:alert(1))">s</span>`:            "<span>s</span>",
		`<span style="color: expression(alert(1)); font-weight: bold">s</span>`: `<span style="font-weight: bold">s</span>`,
		`<div class="x" id="y"><strong>x</strong></div><style>p{}</style>`:      "<div><strong>x</strong></div>",
		`<svg onload="alert(1)"><b>s</b></svg>`:                                 "",
	}
	for in, want := range cases {
		if got := SanitizeInlineHTML(in); got != want {
			t.Errorf("SanitizeInlineHTML(%q) = %q, want %q", in, got, want)
		}
	}
}
