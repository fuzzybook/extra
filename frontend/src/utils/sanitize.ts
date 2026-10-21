// Keeps only the basic formatting produced by the caption editor, without any
// attribute; every other element is unwrapped (its text kept) and dangerous
// elements are dropped with their content. Same rule as validate.SanitizeInlineHTML
// in the backend, applied again here as defense in depth before v-html.
const ALLOWED_TAGS = new Set(['B', 'I', 'U', 'STRONG', 'EM', 'BR']);
const DROPPED_TAGS = new Set(['SCRIPT', 'STYLE', 'IFRAME', 'OBJECT', 'EMBED', 'TEMPLATE', 'NOSCRIPT', 'TEXTAREA', 'TITLE', 'SVG', 'MATH']);

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
