// Keeps only the formatting produced by the caption editor (Quasar q-editor:
// font face/size, bold, italic, underline, lines) with a strict allowlist of
// attributes and CSS properties; every other element is unwrapped (its text
// kept) and dangerous elements are dropped with their content. Same rules as
// validate.SanitizeInlineHTML in the backend, applied again here as defense in
// depth before v-html.
const ALLOWED_TAGS = new Set(['B', 'I', 'U', 'S', 'STRIKE', 'STRONG', 'EM', 'BR', 'FONT', 'SPAN', 'DIV', 'P']);
const DROPPED_TAGS = new Set(['SCRIPT', 'STYLE', 'IFRAME', 'OBJECT', 'EMBED', 'TEMPLATE', 'NOSCRIPT', 'TEXTAREA', 'TITLE', 'SVG', 'MATH']);

const ALLOWED_CSS = new Set([
  'font-family', 'font-size', 'font-weight', 'font-style',
  'text-decoration', 'text-decoration-line', 'color', 'text-align',
]);
// No parentheses: no url(), expression(), var()...
const SAFE_CSS_VALUE = /^[a-zA-Z0-9 #%.,'"-]+$/;
const SAFE_FONT_FACE = /^[a-zA-Z0-9 ,'"-]+$/;
const SAFE_COLOR = /^(#[0-9a-fA-F]{3,8}|[a-zA-Z]+)$/;
const SAFE_FONT_SIZE = /^[1-7]$/;

function sanitizeStyle(style: string): string {
  return style
    .split(';')
    .map((decl) => {
      const idx = decl.indexOf(':');
      if (idx < 0) return '';
      const name = decl.slice(0, idx).trim().toLowerCase();
      const value = decl.slice(idx + 1).trim();
      const lower = value.toLowerCase();
      if (!ALLOWED_CSS.has(name) || !value || !SAFE_CSS_VALUE.test(value) || lower.includes('expression') || lower.includes('javascript')) {
        return '';
      }
      return `${name}: ${value}`;
    })
    .filter(Boolean)
    .join('; ');
}

function copyAllowedAttrs(src: Element, dst: Element, tag: string) {
  for (const attr of Array.from(src.attributes)) {
    const name = attr.name.toLowerCase();
    const value = attr.value.trim();
    if (name === 'style') {
      const style = sanitizeStyle(value);
      if (style) dst.setAttribute('style', style);
    } else if (tag === 'FONT' && name === 'face' && SAFE_FONT_FACE.test(value)) {
      dst.setAttribute('face', value);
    } else if (tag === 'FONT' && name === 'size' && SAFE_FONT_SIZE.test(value)) {
      dst.setAttribute('size', value);
    } else if (tag === 'FONT' && name === 'color' && SAFE_COLOR.test(value)) {
      dst.setAttribute('color', value);
    }
  }
}

function cleanNode(node: Node, doc: Document): Node[] {
  if (node.nodeType === Node.TEXT_NODE) {
    return [doc.createTextNode(node.textContent ?? '')];
  }
  if (node.nodeType !== Node.ELEMENT_NODE) {
    return [];
  }
  const el = node as Element;
  const tag = el.tagName.toUpperCase();
  if (DROPPED_TAGS.has(tag)) {
    return [];
  }
  const children = Array.from(el.childNodes).flatMap((child) => cleanNode(child, doc));
  if (!ALLOWED_TAGS.has(tag)) {
    return children;
  }
  const clean = doc.createElement(tag.toLowerCase());
  copyAllowedAttrs(el, clean, tag);
  children.forEach((child) => clean.appendChild(child));
  return [clean];
}

export function sanitizeInlineHtml(html: string | null | undefined): string {
  if (!html) return '';
  const doc = new DOMParser().parseFromString(`<body>${html}</body>`, 'text/html');
  const out = document.createElement('div');
  Array.from(doc.body.childNodes)
    .flatMap((child) => cleanNode(child, document))
    .forEach((child) => out.appendChild(child));
  return out.innerHTML;
}

/** Plain text of an HTML fragment, without interpreting it as markup. */
export function stripHtml(html: string | null | undefined): string {
  if (!html) return '';
  return (new DOMParser().parseFromString(html, 'text/html').body.textContent ?? '').trim();
}
