package validate

import (
	"html"
	"regexp"
	"strings"

	xhtml "golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// allowedTags is the formatting produced by the caption editor (Quasar
// q-editor with font, size, bold, italic and underline): <font face/size>,
// <span style> on some browsers, <div>/<p> for lines.
var allowedTags = map[atom.Atom]bool{
	atom.B: true, atom.I: true, atom.U: true, atom.S: true, atom.Strike: true,
	atom.Strong: true, atom.Em: true, atom.Br: true,
	atom.Font: true, atom.Span: true, atom.Div: true, atom.P: true,
}

// droppedWithContent are elements whose text must not leak into the output.
var droppedWithContent = map[atom.Atom]bool{
	atom.Script: true, atom.Style: true, atom.Iframe: true,
	atom.Object: true, atom.Embed: true, atom.Template: true,
	atom.Noscript: true, atom.Textarea: true, atom.Title: true,
	atom.Svg: true, atom.Math: true,
}

// allowedCSS are the style properties kept on span/div/p/font.
var allowedCSS = map[string]bool{
	"font-family": true, "font-size": true, "font-weight": true, "font-style": true,
	"text-decoration": true, "text-decoration-line": true, "color": true, "text-align": true,
}

var (
	// Letters, digits, spaces and the punctuation needed by font lists,
	// sizes and colors; no parentheses, so no url(), expression(), var()...
	safeCSSValue = regexp.MustCompile(`^[a-zA-Z0-9 #%.,'"-]+$`)
	safeFontFace = regexp.MustCompile(`^[a-zA-Z0-9 ,'"-]+$`)
	safeColor    = regexp.MustCompile(`^(#[0-9a-fA-F]{3,8}|[a-zA-Z]+)$`)
	safeFontSize = regexp.MustCompile(`^[1-7]$`)
)

// sanitizeStyle keeps only allowed, safely-valued CSS declarations.
func sanitizeStyle(style string) string {
	var kept []string
	for _, decl := range strings.Split(style, ";") {
		name, value, ok := strings.Cut(decl, ":")
		if !ok {
			continue
		}
		name = strings.ToLower(strings.TrimSpace(name))
		value = strings.TrimSpace(value)
		lower := strings.ToLower(value)
		if !allowedCSS[name] || value == "" || !safeCSSValue.MatchString(value) ||
			strings.Contains(lower, "expression") || strings.Contains(lower, "javascript") {
			continue
		}
		kept = append(kept, name+": "+value)
	}
	return strings.Join(kept, "; ")
}

// cleanAttrs returns the allowed attributes of a kept tag, already escaped.
func cleanAttrs(a atom.Atom, z *xhtml.Tokenizer, hasAttr bool) string {
	var out strings.Builder
	for hasAttr {
		var key, val []byte
		key, val, hasAttr = z.TagAttr()
		k, v := strings.ToLower(string(key)), strings.TrimSpace(string(val))
		keep := false
		switch {
		case k == "style":
			v = sanitizeStyle(v)
			keep = v != ""
		case a == atom.Font && k == "face":
			keep = safeFontFace.MatchString(v)
		case a == atom.Font && k == "size":
			keep = safeFontSize.MatchString(v)
		case a == atom.Font && k == "color":
			keep = safeColor.MatchString(v)
		}
		if keep {
			out.WriteString(" " + k + `="` + html.EscapeString(v) + `"`)
		}
	}
	return out.String()
}

// SanitizeInlineHTML keeps only the formatting produced by the caption editor
// (bold, italic, underline, font face/size/color, lines) with a strict
// allowlist of attributes and CSS properties, and escapes all text. Scripts,
// links, images, event handlers and anything else are removed: the result is
// safe to render with v-html.
func SanitizeInlineHTML(raw string) string {
	if !strings.ContainsAny(raw, "<>&") {
		return raw
	}
	z := xhtml.NewTokenizer(strings.NewReader(raw))
	var out strings.Builder
	var open []atom.Atom
	skipDepth := 0
	for {
		tt := z.Next()
		switch tt {
		case xhtml.ErrorToken:
			for i := len(open) - 1; i >= 0; i-- {
				out.WriteString("</" + open[i].String() + ">")
			}
			return out.String()
		case xhtml.TextToken:
			if skipDepth == 0 {
				out.WriteString(html.EscapeString(string(z.Text())))
			}
		case xhtml.StartTagToken, xhtml.SelfClosingTagToken:
			name, hasAttr := z.TagName()
			a := atom.Lookup(name)
			if droppedWithContent[a] {
				if tt == xhtml.StartTagToken {
					skipDepth++
				}
				continue
			}
			if skipDepth > 0 || !allowedTags[a] {
				continue
			}
			if a == atom.Br {
				out.WriteString("<br>")
				continue
			}
			if tt == xhtml.StartTagToken {
				out.WriteString("<" + a.String() + cleanAttrs(a, z, hasAttr) + ">")
				open = append(open, a)
			}
		case xhtml.EndTagToken:
			name, _ := z.TagName()
			a := atom.Lookup(name)
			if droppedWithContent[a] {
				if skipDepth > 0 {
					skipDepth--
				}
				continue
			}
			if skipDepth > 0 || !allowedTags[a] || a == atom.Br {
				continue
			}
			// Close only a tag that is actually open, so the output stays balanced.
			for i := len(open) - 1; i >= 0; i-- {
				if open[i] == a {
					for j := len(open) - 1; j >= i; j-- {
						out.WriteString("</" + open[j].String() + ">")
					}
					open = open[:i]
					break
				}
			}
		}
	}
}
