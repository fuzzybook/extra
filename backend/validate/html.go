package validate

import (
	"html"
	"strings"

	xhtml "golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// allowedInlineTags are the formatting tags the photo caption editor produces.
var allowedInlineTags = map[atom.Atom]bool{
	atom.B: true, atom.I: true, atom.U: true,
	atom.Strong: true, atom.Em: true, atom.Br: true,
}

// droppedWithContent are elements whose text must not leak into the output.
var droppedWithContent = map[atom.Atom]bool{
	atom.Script: true, atom.Style: true, atom.Iframe: true,
	atom.Object: true, atom.Embed: true, atom.Template: true,
	atom.Noscript: true, atom.Textarea: true, atom.Title: true,
}

// SanitizeInlineHTML keeps only basic formatting tags (b, i, u, strong, em,
// br) without any attribute, and escapes all text. Everything else is
// removed: the result is safe to render with v-html.
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
			name, _ := z.TagName()
			a := atom.Lookup(name)
			if droppedWithContent[a] {
				if tt == xhtml.StartTagToken {
					skipDepth++
				}
				continue
			}
			if skipDepth > 0 || !allowedInlineTags[a] {
				continue
			}
			if a == atom.Br {
				out.WriteString("<br>")
				continue
			}
			if tt == xhtml.StartTagToken {
				out.WriteString("<" + a.String() + ">")
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
			if skipDepth > 0 || !allowedInlineTags[a] || a == atom.Br {
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
